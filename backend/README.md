> **Port en curso (VerdePUCP).** Esta carpeta es una copia fiel de `backend/` de la rama `init/backend` del equipo (HEAD `1e876dd`). El backend actual (`apps/api`) se porta aquí por lotes siguiendo `docs/PLAN-MIGRACION-BACKEND.md`; hasta el corte, `apps/api` sigue siendo el backend desplegado.

# Backend - Áreas Verdes PUCP

API del Sistema de Gestión de Áreas Verdes del campus PUCP. Usa Go 1.25.3,
Gin, GORM/PostgreSQL y `dig` para inyección de dependencias, con arquitectura
por capas basada en el servicio de referencia del equipo.

## Requisitos

- Go 1.25.3
- PostgreSQL 16 con PostGIS 3.4+ (a diferencia de la API anterior, la nueva API valida la conexión a la base de datos al iniciar y falla si no está disponible; en producción el entrypoint corre `migrate` antes de iniciar la API)
- Docker

## Ejecución local

```bash
cd backend
cp app/.env.example app/.env
make run
```

El servidor inicia por defecto en `http://localhost:8080`.

## API y Swagger

La API usa el prefijo `/areas-verdes/v1`. La interfaz de Swagger está en
`http://localhost:8080/areas-verdes/v1/swagger/index.html` y la definición JSON
que carga está en
`http://localhost:8080/areas-verdes/v1/swagger/doc.json`.

Para que un endpoint nuevo forme parte de la API y aparezca en Swagger, hay que
conectarlo en dos lugares: el registro de rutas y las anotaciones del controller.
El grupo organiza y monta rutas Gin; el controller implementa las operaciones.

### Flujo para agregar un group-controller

1. **Crear el controller** bajo `app/internal/presentation/controller/`.
   Define su interfaz, implementación y constructor (por ejemplo,
   `NewPlantsController`). Anota cada método HTTP exportado para `swaggo/swag`:

   ```go
   // List godoc
   // @Summary Listar áreas verdes
   // @Description Devuelve las áreas verdes registradas
   // @Tags areas-verdes
   // @Produce json
   // @Success 200 {array} AreaResponse
   // @Router /v1/areas-verdes [get]
   func (c *plantsController) List(ctx *gin.Context) { /* ... */ }
   ```

   Usa anotaciones `@Param`, `@Accept`, `@Produce`, `@Success`, `@Failure` y
   `@Router` según el contrato. `@Router` debe incluir `/v1`, pero no
   `/areas-verdes`: ese prefijo ya está declarado como `@BasePath` en
   `app/cmd/main.go`. Los tipos de request/response referenciados deben estar
   exportados para que Swagger pueda describirlos.

2. **Crear un route group** en
   `app/internal/presentation/routes/groups/`. Recibe el controller por
   constructor y registra los métodos y paths en `Register`, usando paths
   relativos al prefijo común:

   ```go
   type PlantsGroup struct{ controller controller.IPlantsController }

   func NewPlantsGroup(c controller.IPlantsController) *PlantsGroup {
       return &PlantsGroup{controller: c}
   }

   func (g *PlantsGroup) Register(router gin.IRouter) {
       router.GET("/areas-verdes", g.controller.List)
   }
   ```

3. **Registrar las dependencias en `app/internal/presentation/container.go`.**
   Añade `groups.NewPlantsGroup` a la lista de Groups y
   `controller.NewPlantsController` a la lista de Controllers. `dig` construye
   el group e inyecta el controller.

4. **Conectar el group al router** en
   `app/internal/presentation/routes/routes.go`: añádelo a `Router`, a
   `RouterParams` y a `NewRouter`, y llama `r.plantsGroup.Register(servicePath)`
   dentro de `Setup`. `servicePath` ya representa `/areas-verdes/v1`.

5. **Regenerar los documentos** desde `backend/`:

   ```bash
   make swagger
   ```

   El target genera `app/docs/docs.go`, `app/docs/swagger.json` y
   `app/docs/swagger.yaml`, y actualiza `openapi/openapi.base.json`. Requiere la
   CLI `swag` (`go install github.com/swaggo/swag/cmd/swag@v1.16.6`),
   `swagger2openapi` y Node.js. No edites manualmente los archivos de
   `app/docs/`: son generados. La aplicación ya importa ese paquete desde
   `app/cmd/main.go` y monta el UI mediante `SwaggerGroup`.

6. **Reiniciar `make run` y comprobar** el endpoint y Swagger. La ruta real del
   ejemplo sería `GET /areas-verdes/v1/areas-verdes`; en Swagger debe aparecer
   bajo la etiqueta `areas-verdes`. Si el UI abre pero muestra “Failed to load
   API definition”, abre `.../swagger/doc.json`: un 500 suele indicar que la
   especificación generada o registrada no es válida; comprueba que ejecutaste
   `make swagger` y que los archivos `app/docs/` corresponden a las anotaciones
   actuales.

### Health check

Endpoint de ejemplo actualmente expuesto:

```bash
curl http://localhost:8080/areas-verdes/v1/health
```

Respuesta esperada:

```json
{
  "status": "healthy",
  "timestamp": "2026-09-24T20:00:00Z"
}
```

## Comandos

```bash
make run        # Ejecuta la API
make build      # Genera app/bin/api
make test       # Ejecuta las pruebas
make tidy       # Ordena las dependencias del módulo
make swagger    # Regenera la documentación swagger
make gen-models # Genera modelos GORM y ejecuta control de deriva contra BD desechable
```

### Control de deriva y generación de modelos (`make gen-models`)

`make gen-models` ejecuta `cmd/modelgen` como control de deriva de nuestros modelos en `internal/persistence/models/*.model.go` y genera modelos GORM en un directorio temporal o `$MODELGEN_OUT`.

Por seguridad, `modelgen` cuenta con guardas estrictas que abortan si:
1. Apunta a la base de datos de datos reales (`campus_verde`).
2. El nombre de la base de datos no comienza con el prefijo desechable `vp_` o `modelgen_`.
3. La base de datos no fue construida con nuestras migraciones (falta `schema_migrations`).
4. La tabla `areas_verdes` contiene datos (debe ser una BD vacía creada por `cmd/migrate` sin ETL).

**Flujo de uso:**
```bash
# 1. Crear una base de datos desechable con prefijo vp_ o modelgen_
psql "postgres://campus:campus@127.0.0.1:5432/campus_verde?sslmode=disable" -c "CREATE DATABASE vp_modelgen_x"

# 2. Aplicar nuestras migraciones con cmd/migrate
(cd app && DATABASE_URL="postgres://campus:campus@127.0.0.1:5432/vp_modelgen_x?sslmode=disable" MIGRATIONS_DIR="../../db/migrations" go run ./cmd/migrate)

# 3. Ejecutar gen-models apuntando a la base desechable
DATABASE_URL="postgres://campus:campus@127.0.0.1:5432/vp_modelgen_x?sslmode=disable" make gen-models

# 4. Eliminar la base de datos desechable
psql "postgres://campus:campus@127.0.0.1:5432/campus_verde?sslmode=disable" -c "DROP DATABASE vp_modelgen_x"
```

## Estructura

```text
backend/
├── app/
│   ├── cmd/                    # Entry point y composition root (dig)
│   └── internal/
│       ├── application/        # Casos de uso, servicios, contratos y DTOs
│       ├── domain/             # Entidades, constantes y errores de dominio
│       ├── infrastructure/     # Adaptadores de servicios externos
│       ├── persistence/        # GORM, modelos, mappers y repositorios
│       ├── presentation/       # HTTP, controladores, middleware y rutas
│       └── shared/             # Configuración, logging y utilitarios
├── dockerfile
└── Makefile
```

La configuración se obtiene de variables de entorno; los
valores de `.env.example` son solo para desarrollo y nunca se deben versionar
secretos reales.
