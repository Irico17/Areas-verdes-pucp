# Plan de migración del backend (`apps/api`) a la arquitectura `init/backend`

> **Alcance.** Plan para portar el backend actual de VerdePUCP (`apps/api`, desplegado, v1-propuesta-referencia `115a56d`) a la arquitectura por capas de la rama `init/backend` del repositorio del equipo `GRUPO-12-DP2/-areas-verdes-pucp` (clon local de esa rama, HEAD `1e876dd chore: upd README`).
> Este documento es **solo análisis y plan**: no cambia código ni esquema. Todo lo afirmado se verificó en el código; lo que no se pudo determinar se marca como **(no determinado)**.
> Fecha: 2026-09-27.

---

## 0. Resumen ejecutivo

| Tema | Decisión propuesta |
|---|---|
| Estrategia | **Strangler por módulos**: se crea `backend/` (copia fiel de `init/backend`) junto a `apps/api`; cada lote porta un módulo completo por las capas domain → application → persistence → presentation, verificado por **paridad de respuestas** contra `apps/api` sobre la misma BD. El corte se hace una sola vez al final. |
| Módulo Go | Se conserva el de init: `github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend` en `backend/app/`, para que los archivos pasen al repo del equipo sin reescribir imports. |
| Esquema de BD | **Nuestras 34 migraciones SQL siguen siendo la fuente de verdad** (PostGIS, UUID, datos cargados). `db/schema_nucleo_v0.2.sql` pasa a ser el **modelo conceptual del dominio** (nombres de entidades), mapeado a nuestras tablas con `models` + `mapper`. Cerramos las brechas con **migraciones aditivas e idempotentes (045+)** y vistas de compatibilidad en un esquema `nucleo`. **Nunca** se aplica `schema_nucleo_v0.2.sql` sobre la BD desplegada y nunca se hace `DROP`/`TRUNCATE` de datos cargados. |
| API | Prefijo nuevo `/areas-verdes/v1` (el de init) **más un alias `/api/v1`** durante la transición; `/health` en la raíz se mantiene con su cuerpo actual porque lo consume el healthcheck de la EC2. Formato de error `{"error": "..."}`, cookie `cv_sesion` y nombres de campos JSON **sin cambios**. |
| Roles | Los **códigos técnicos no cambian** (`capataz`, `coordinacion`, `jefatura`, `admin`); cambian los **nombres visibles** a los del backlog v2 (Capataz, Ingeniería/Coordinación, Jefatura, Administrador) en `roles.nombre`, con historial en `cambios`, y `usuarios.rol` pasa de `CHECK` a FK hacia `roles` para que los roles sean configurables (RNF-05). |
| Tamaño | 24 lotes numerados (§9), cada uno acotado para una corrida de Claude Sonnet, con criterios verificables (`make build`, `make test`, paridad de endpoints, conteos de BD). |

---

## 1. Arquitectura de `init/backend` (descripción exacta)

### 1.1 Estado del repositorio

- Historia (de más reciente a más antigua): `1e876dd chore: upd README`, `238bd85 chore: remove unused files`, `21a03d9 feat: add swagger docs to server`, `2ca73f1 faet: init structure for backend`, `0e84bea chore: add backend stack recomendation`, … `466991b docs: agrega README y .gitignore iniciales`.
- `README.md` raíz (línea 7): «backend inicializado. El desarrollo funcional arranca en la Semana 7 (~28-set), Sprint 1».
- **Verificado aquí:** con `GOTOOLCHAIN=auto` (descarga go1.25.3) `go build ./...`, `go vet ./...` y `go test ./...` terminan bien en `backend/app`; **no hay ningún archivo de test** (`[no test files]` en todos los paquetes). Con el Go local 1.24.4 y `GOTOOLCHAIN=local` falla: `go.mod requires go >= 1.25.3`.

### 1.2 Lenguaje, versión y librerías (`backend/app/go.mod`)

| Elemento | Valor |
|---|---|
| Módulo | `github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend` |
| Go | `go 1.25.3` |
| HTTP | `github.com/gin-gonic/gin v1.11.0` |
| ORM / driver | `gorm.io/gorm v1.31.1`, `gorm.io/driver/postgres v1.6.0` (pgx v5.6.0 indirecto) |
| IoC | `go.uber.org/dig v1.19.0` |
| Config | `github.com/joho/godotenv v1.5.1` |
| Logger | `github.com/rs/zerolog v1.35.1` |
| Swagger | `github.com/swaggo/swag v1.16.6`, `github.com/swaggo/gin-swagger v1.6.1`, `github.com/swaggo/files v1.0.1` |

No incluye: bcrypt/`x/crypto`, AWS SDK, YAML, librería de validación aparte de la de Gin (validator/v10 va indirecto) ni librería de migraciones.

### 1.3 Layout completo

Leyenda: ✅ implementado · ⬜ carpeta vacía (solo `.gitkeep`).

```text
(raíz)
├── .env.example                  ✅ POSTGRES_DB/USER/PASSWORD/PORT, DATABASE_URL; app y AWS comentados
├── .github/pull_request_template.md ✅ (no hay workflows de CI)
├── .gitignore                    ✅ bloquea *.pdf/*.docx/*.xlsx…, .env*, backend/app/bin/
├── CONTRIBUTING.md               ✅
├── README.md                     ✅
├── docker-compose.yml            ✅ solo servicio db (postgres:16) + ./db montado en initdb
├── db/schema_nucleo_v0.2.sql     ✅ 36 tablas + FKs + índices + semillas (551 líneas)
├── docs/DECISIONES.md            ✅   docs/onboarding.md ✅   docs/sprint-1-backlog.md ✅
├── frontend/README.md            ✅ «Pendiente de inicializar»
└── backend/
    ├── Makefile                  ✅ build, run, test, tidy, swagger
    ├── README.md                 ✅ guía «group-controller»
    ├── dockerfile                ✅ (en minúsculas)
    ├── openapi/openapi.base.json ✅ generado (OpenAPI 3.0.0)
    ├── openapi/sanitize-openapi.js ✅
    └── app/
        ├── .env.example          ✅ SERVER_*, DATABASE_* (7 variables)
        ├── go.mod / go.sum       ✅
        ├── cmd/main.go           ✅ arranque + anotaciones generales de swag
        ├── cmd/ioc/container.go  ✅ BuildContainer (dig)
        ├── docs/docs.go, swagger.json, swagger.yaml ✅ generados por swag (no editar)
        └── internal/
            ├── application/container.go            ✅ RegisterContainer → return nil
            ├── application/contracts/              ⬜
            ├── application/dto/                    ⬜
            ├── application/services/               ⬜
            ├── application/usecases/               ⬜
            ├── domain/constants/enums/             ⬜
            ├── domain/entities/                    ⬜
            ├── domain/errors/                      ⬜
            ├── infrastructure/container.go         ✅ RegisterContainer → return nil
            ├── persistence/container.go            ✅ Provide(database.NewConnection)
            ├── persistence/database/database.go    ✅
            ├── persistence/mapper/                 ⬜
            ├── persistence/models/                 ⬜
            ├── persistence/repository/postgres/    ⬜
            ├── presentation/container.go           ✅ router + groups + controllers
            ├── presentation/controller/health.controller.go ✅
            ├── presentation/middleware/            ⬜
            ├── presentation/requests/              ⬜
            ├── presentation/routes/routes.go       ✅
            ├── presentation/routes/groups/health.group.go  ✅
            ├── presentation/routes/groups/swagger.group.go ✅
            ├── shared/config/config.go             ✅
            └── shared/logger/logger.go             ✅
```

El `backend/README.md` (línea 144) menciona `shared/` como «Configuración, logging y utilitarios», pero **no existe** `shared/utils`.

### 1.4 Capas y contenedores IoC

- `cmd/ioc/container.go` → `BuildContainer()` crea un `dig.Container` y registra, en orden: `config.New`; un `zerolog.Logger` vía `logger.InitLogger(cfg.Server.GinMode)`; un `*gin.Engine` (`gin.SetMode(cfg.Server.GinMode)` + `gin.New()`); y los registradores de capa `application.RegisterContainer`, `infrastructure.RegisterContainer`, `persistence.RegisterContainer`, `presentation.RegisterContainer`.
- Cada capa expone `func RegisterContainer(*dig.Container) error`. Hoy solo `persistence` (provee `database.NewConnection`) y `presentation` (provee `routes.NewRouter`, `groups.NewHealthGroup`, `groups.NewSwaggerGroup`, `controller.NewHealthController`) registran algo.
- `cmd/main.go` hace `container.Invoke(func(router *routes.Router, engine *gin.Engine, cfg *config.Config) error { router.Setup(); return engine.Run(":"+cfg.Server.Port) })`.
- Como nadie pide `*gorm.DB`, **la conexión a BD nunca se abre** en el estado actual (el comentario de `persistence/container.go` dice «lazy»).
- Dirección de dependencias que sugieren los nombres de carpetas (no hay código que lo muestre): `domain` (entidades, enums, errores) ← `application` (contratos = puertos, DTO, casos de uso, servicios) ← `persistence` (modelos GORM, mappers, repositorios Postgres que implementan los contratos) / `infrastructure` (adaptadores externos) ← `presentation` (controllers, requests, middleware, rutas).

### 1.5 Convenciones de nombres

**Observadas en el código:**

| Elemento | Convención | Ejemplo |
|---|---|---|
| Controller | `internal/presentation/controller/<recurso>.controller.go`; interfaz `I<X>Controller`, struct privado `<x>Controller`, constructor `New<X>Controller() I<X>Controller`; método por operación con bloque godoc de swag | `health.controller.go`: `IHealthController`, `healthController`, `NewHealthController`, `Health(*gin.Context)` |
| DTO de respuesta del controller | tipo exportado en el mismo archivo | `HealthResponse{Status, Timestamp}` |
| Group | `internal/presentation/routes/groups/<recurso>.group.go`; struct `<X>Group{controller}`, `New<X>Group(controller)`, método `Register(router gin.IRouter)` con rutas relativas | `health.group.go` |
| Router | `routes.go`: struct `Router` con un campo por group, `RouterParams` con `dig.In`, `NewRouter(p RouterParams)`, `Setup()` que monta `gin.Recovery()` y llama a `Register(servicePath)` de cada group | `basePath = "/areas-verdes/v1"` |
| Registro IoC | listas de constructores `[]any{...}` en `presentation/container.go` | Groups y Controllers |
| Paquetes | uno por carpeta (`controller`, `groups`, `routes`, `config`, `logger`, `database`) | — |
| Comentarios | godoc en inglés en el código Go; README y docs en español | — |

**No definidas en init** (las carpetas están vacías): nombres de archivos de entidades, DTO, contratos, casos de uso, servicios, modelos, mappers, repositorios, requests y middleware. Este plan propone, por analogía con `*.controller.go` / `*.group.go`, y **pendiente de ratificar por el arquitecto del equipo** (rol de arquitecto definido en `docs/onboarding.md` del equipo):

| Carpeta | Archivo propuesto | Tipos |
|---|---|---|
| `domain/entities` | `<entidad>.entity.go` | `Intervencion`, `ActividadEvento`… |
| `domain/constants/enums` | `<nombre>.enum.go` | `type EstadoIntervencion string` + constantes |
| `domain/errors` | `<modulo>.errors.go` | `var ErrNoEncontrado = errors.New(...)`, `type ValidacionError` |
| `application/dto` | `<entidad>.dto.go` | `IntervencionDTO`, `CrearIntervencionDTO` |
| `application/contracts` | `<entidad>.contract.go` | `IIntervencionRepository`, `IBlobStore` |
| `application/usecases` | `<agregado>.usecase.go` | `IIntervencionUseCase`, `intervencionUseCase`, `NewIntervencionUseCase` |
| `application/services` | `<servicio>.service.go` | servicios transversales: permisos, auditoría (`cambios`) |
| `persistence/models` | `<tabla>.model.go` | struct GORM con `TableName()` apuntando a **nuestra** tabla |
| `persistence/mapper` | `<entidad>.mapper.go` | `ToEntity(model)`, `ToModel(entity)` |
| `persistence/repository/postgres` | `<entidad>.repository.go` | `intervencionRepository` implementa `contracts.IIntervencionRepository` |
| `infrastructure` | `<adaptador>/<x>.go` (p. ej. `storage/s3.storage.go`) | S3/disco, bcrypt, lectores Excel/CSV |
| `presentation/requests` | `<entidad>.request.go` | cuerpos con `binding:"..."` |
| `presentation/middleware` | `<nombre>.middleware.go` | `auth`, `permiso`, `cors`, `bitacora`, `limite_login`, `errores` |

> **Desviación de contratos:** Las interfaces de casos de uso (`I<X>UseCase`) y servicios residen en `application/contracts` (en lugar de `application/usecases`) para mantener a la capa de presentación dependiendo exclusivamente de contratos abstractos.

### 1.6 Acceso a BD y migraciones

- `persistence/database/database.go` → `NewConnection(cfg *config.Config) (*gorm.DB, error)` arma un DSN `host=… port=… user=… password=… dbname=… search_path=<schema> sslmode=…` y llama `gorm.Open(postgres.Open(dsn), &gorm.Config{})`. No configura pool, ni logger de GORM, ni reintentos, ni `DATABASE_URL`.
- **No existe mecanismo de migraciones** en el backend (ni librería, ni runner, ni `AutoMigrate`).
- El esquema vive en `db/schema_nucleo_v0.2.sql` y lo ejecuta **Postgres en su initdb** gracias a `./db:/docker-entrypoint-initdb.d:ro` en `docker-compose.yml`, *solo cuando el volumen está vacío* (el propio compose lo advierte: «para re-aplicar: docker compose down -v»). `CONTRIBUTING.md` exige: «Todo cambio de esquema de BD se refleja en `db/` (migración o actualización del `.sql`)».
- `schema_nucleo_v0.2.sql`: 36 tablas con `bigint GENERATED ALWAYS AS IDENTITY`, nombres en singular (`rol`, `usuario`, `intervencion`, `actividad_evento`, `evidencia`…), coordenadas como `numeric(9,6)` (**sin PostGIS**, sin polígonos), FK al final, 10 índices y semillas (4 roles, 6 estados con color, 3 prioridades, 5 fuentes, 9 clases de actividad). El admin inicial está comentado («crear por migración con hash real»).

### 1.7 Autenticación

**No existe** en código: no hay middleware, ni sesión, ni JWT, ni hash de contraseñas. El esquema núcleo prevé `usuario.password_hash`, `debe_cambiar_password`, `ultimo_acceso`, y RBAC `rol`/`permiso`/`rol_permiso`. `docs/DECISIONES.md` deja **P-09 Autenticación (SSO PUCP vs. JWT propio)** como pendiente del cliente; `.env.example` raíz tiene `# JWT_SECRET=…` comentado; `sprint-1-backlog.md` pide «inicia sesión y recibe token; … contraseña hasheada». El backlog v2 ya la cierra: *RF-01 «Credenciales propias administradas por el jefe de sección (sin SSO): Aprobado por el cliente en la 3.ª reunión»*.

### 1.8 Configuración y logger

- `shared/config/config.go`: `Config{Server{Port, GinMode}, Database{Host, Port, User, Password, Name, Schema, SSLMode}}`; `New()` hace `godotenv.Load()` (ignora error) y lee `SERVER_PORT` (8080), `SERVER_GIN_MODE` (debug), `DATABASE_HOST/PORT/USER/PASSWORD/NAME/SCHEMA/SSL_MODE` con valores por defecto de desarrollo.
- `shared/logger/logger.go`: `InitLogger(mode)` configura el `log.Logger` global de zerolog (consola bonita + nivel debug en `debug`; JSON + nivel info en otro caso), con timestamp y caller. El `zerolog.Logger` se registra en dig pero **ningún componente lo consume**: `main.go` usa el logger global y no hay middleware de logging de peticiones (solo `gin.Recovery()`).

### 1.9 Tests

Ninguno. `make test` = `cd app && go test ./...`.

### 1.10 Estilo de API

- Base: `const basePath = "/areas-verdes/v1"` (`routes.go`). Único endpoint funcional: `GET /areas-verdes/v1/health` → `{"status":"healthy","timestamp":"<RFC3339>"}` (no consulta la BD).
- Swagger: `SwaggerGroup` monta `GET /areas-verdes/v1/swagger` (301 a `/areas-verdes/v1/swagger/index.html`) y `GET /areas-verdes/v1/swagger/*any` (`ginSwagger.WrapHandler`). Se monta **siempre**, también en modo release.
- Anotaciones generales en `cmd/main.go`: `@title Areas Verdes API $SWAGGER_ENV`, `@version 1.0.0`, `@host localhost:8080`, `@BasePath /areas-verdes`. Por eso cada `@Router` lleva `/v1/...` y no `/areas-verdes` (README, paso 1).
- `make swagger`: `swag init -g cmd/main.go -o docs/` → `swagger2openapi app/docs/swagger.json -o openapi/openapi.base.json` → `node openapi/sanitize-openapi.js openapi/openapi.base.json`. Requiere la CLI `swag` v1.16.6, `swagger2openapi` y Node.
- `sanitize-openapi.js`: deja alfanuméricos los nombres de schema (API Gateway lo exige; `controller.HealthResponse` → `controllerHealthResponse`), detecta colisiones y marca con `x-apigw-integration {proxy, stream, timeoutInMillis}` las operaciones que producen `text/event-stream` (pensado para `openapi-to-apigateway-lib`). Indica que el equipo apunta a **AWS API Gateway** delante del backend.
- `openapi.base.json` generado declara `servers: [{"url": "//localhost:8080/areas-verdes"}]`.
- No hay convención definida de formato de error, paginación ni versionado más allá de `/v1`.

### 1.11 Dockerfile y docker-compose

- `backend/dockerfile`: etapa `golang:1.25.3-alpine` (contexto = `backend/`; `COPY app/go.mod app/go.sum`, `go mod download && go mod verify`, `CGO_ENABLED=0 go build -ldflags="-w -s" -o /api ./cmd`); runtime `alpine:3.22` con `ca-certificates tzdata`, usuario no root `app`, `EXPOSE 8080`, `ENTRYPOINT ["/app/api"]`. Un solo binario.
- `docker-compose.yml` (raíz): solo `db` con `postgres:16` (sin PostGIS), `restart: unless-stopped`, puerto `${POSTGRES_PORT:-5432}`, volumen `pgdata`, `./db` en initdb, healthcheck `pg_isready`. No hay servicio de API.

### 1.12 Lo que dicen los documentos

- **README:** PWA única con ruteo por rol (`/capataz/*` mobile-first, `/oficina/*` desktop-first); offline solo capataz (IndexedDB + Background Sync); full cloud AWS; PostgreSQL con modelo núcleo v0.2; backend Go 1.25.3 + Gin + GORM + dig; mapa OSM; roles Jefatura, Ingeniería/Coordinación, Capataz, operario sin cuenta; «Actividad» = entidad `intervencion`; trazabilidad como cadena `actividad_evento`; bajas lógicas «nunca borrado físico»; estados *Por iniciar → En proceso → Ejecutado → Cerrado* + *Cancelado/Archivado*.
- **CONTRIBUTING:** Git Flow simplificado (`main` protegida, `feature/…`, `fix/…`), Conventional Commits, PR con plantilla y revisión ≥1, sin secretos, **sin valores de negocio hardcodeados (RNF-05)**, cambios de BD reflejados en `db/`, decisiones en `docs/DECISIONES.md`.
- **docs/DECISIONES.md:** D-01…D-11 tomadas (PWA, offline, AWS, PostgreSQL, OSM, sin integraciones, actividad = intervención, estados, bajas lógicas, RBAC/secretos, español). Pendientes P-01…P-10; P-01 y P-02 están anotadas en línea como «->Go» y «->GORM».
- **docs/sprint-1-backlog.md:** Sprint 1 (28-set → 9-oct): auth RF-01, roles RF-02/03, alta de usuarios, catálogos RF-24, inventario/zonificación RF-04/06, esqueleto PWA, base offline; DoD con linter + pruebas, cambios de BD en `db/`, despliegue en AWS.
- **docs/onboarding.md:** verificar «36 tablas y semillas»; activar CI; «se puede arrancar con auth propio (JWT) y adaptar luego».

### 1.13 Defectos y huecos de `init/backend` que el port debe cubrir

1. `gin.New()` sin `SetTrustedProxies`: Gin confía en cualquier proxy, así que `ClientIP()` acepta `X-Forwarded-For` falsificado. Nuestro límite de login por IP (`handlers.LimiteLogin`) quedaría evadible. Nuestro `server.New` hace `r.SetTrustedProxies([]string{})` (`apps/api/internal/server/server.go:30`).
2. Swagger UI siempre montado. Detrás de nuestro nginx, la CSP `script-src 'self'` (`apps/web/nginx-headers.conf`) bloquearía los scripts en línea de Swagger UI.
3. Firmas distintas: `HealthGroup.Register(gin.IRouter)` frente a `SwaggerGroup.Register(*gin.RouterGroup)`.
4. El health no mira la BD; el nuestro sí (`database`, `postgis`).
5. Compose con `postgres:16` sin PostGIS; nuestra migración `001_postgis.sql` hace `CREATE EXTENSION postgis`.
6. No hay runner de migraciones, CI, tests ni middleware.
7. `@host localhost:8080` fijo en el swagger generado.

---

## 2. Inventario del backend actual (`apps/api`)

### 2.1 Stack

`apps/api/go.mod`: módulo `campusverde/api`, `go 1.22.2`, `gin v1.10.0`, `gorm v1.25.12`, `gorm.io/driver/postgres v1.5.11`, `golang.org/x/crypto v0.23.0` (bcrypt), `gopkg.in/yaml.v3` (contrato OpenAPI) y `aws-sdk-go-v2` + `service/s3`. El SDK de AWS aparece como `// indirect` aunque `internal/blobs/s3.go` lo importa directamente; `go mod tidy` lo corrige. PostgreSQL 16 + PostGIS 3.4 (`postgis/postgis:16-3.4`). ~23.600 líneas Go + SQL. **Verificado:** `go build ./...` y `go test ./...` pasan (los tests con BD se saltan si falta `MIGRATE_TEST_URL`).

### 2.2 Paquetes de `apps/api/internal`

| Paquete | Archivos principales | Responsabilidad |
|---|---|---|
| `server` | `server.go`, `bitacora.go` | Arma el engine Gin: `Bitacora()` (log de peticiones con id), `Recovery`, `CORS`, `ConCookie`, `LimiteLogin`, resolución de sesión desde la cookie a `c.Set("usuario")`, llamadas `Registrar*` y `NoRoute` → 404 `{"error":"ruta no encontrada"}` |
| `handlers` | 15 archivos | Controladores HTTP por «frente»; `Deps` común; `exige(c, accion)` y `actorDeSesion` (`producto.go:17-49`); CORS/cookie/límite (`seguridad.go`); contrato OpenAPI (`meta.go`) |
| `accesos` | `accesos.go`, `cookie.go`, `limite.go` | Cuentas locales, bcrypt, sesiones opacas (token de 32 bytes, se guarda su SHA-256, 12 h), matriz de permisos en código (`Matriz`), `Ensure` (semillas + reescritura de `permisos`), cookie `cv_sesion` HttpOnly, rate-limit en memoria |
| `catalogos` | `store.go` | Catálogo genérico `catalogos(clase, codigo, nombre, activo, orden)` con baja lógica |
| `catastro` | `store.go`, `modelo.go`, `sector.go` | GeoJSON de áreas, zonas y capas (PostGIS `ST_AsGeoJSON`, bbox), fichas de área, zonas de supervisión, polígonos de cuadrilla, cuadrillas, lugares, especies, ejemplares, recodificación, capas de referencia |
| `inventario` | `store.go` | Overlays de inventario heredado (tabla `inventario` + uniones con capas) |
| `capas` | `store.go` | Frente 2B: tachos, bebederos, puntos PUCP, reservas de jardín, fichas de capas; PATCH parcial; baja lógica; CSV |
| `operacion` | `model.go`, `store.go`, `validate.go` | Labores (`actividades`), asignación, estados, archivo con motivo, timeline de `actividad_eventos`, ficha de escritorio, validación del recinto del campus |
| `atencion` | `store.go`, `poda_vivero.go`, `riego.go`, `export.go`, `indicadores.go`, `ia.go` | Solicitudes, órdenes de servicio, riego, evidencias (metadatos), reporte de labores + exportación CSV/Excel XML, poda, vivero, avances, sugerencia de tipo por palabras clave |
| `evidencias` | `subir.go` | Subida multipart: validación de MIME real, SHA-256 idempotente (409 si cambia el hash), EXIF filtrado, permisos por labor, evento `evidencia` |
| `blobs` | `open.go`, `disk.go`, `s3.go` | Interfaz `Store{Put, Open}`: S3 si hay `EVIDENCIAS_BUCKET`, si no disco |
| `auditoria` | `store.go` | Tabla `cambios` (antes/después), lotes de importación reversibles, edición auditada, historial y timeline |
| `etl` | 22 archivos (~6.000 líneas) | Normalización de `data/raw` → `data/v1`, carga inicial (`Load`, con `TRUNCATE`), carga por lote con upsert (`lote*.go`), importación desde Excel/CSV con vista previa (`importar.go`, `formato.go`), sectores, inventario, anonimización |
| `migrate` | `migrate.go` | Runner propio: `schema_migrations(version)`, orden lexicográfico, una transacción por archivo, `splitSQL` que respeta `$$` |
| `geojson` | `geojson.go`, `format.go` | Tipos FeatureCollection y formateo |
| `models` | `models.go` | 3 structs GORM (`AreaVerde`, `Zona`, `CapaAuxiliar`). El resto usa SQL crudo. El esquema lo definen las migraciones, **nunca `AutoMigrate`** |
| `config` | `config.go` | Variables de entorno (`DATABASE_URL`, `API_ADDR=:8091`, `MIGRATIONS_DIR`, `OPENAPI_PATH`, `EDIFICIOS_PATH`, `RESERVAS_MOCK_PATH`, `CAMPUS_DEV_PASSWORD`, `EVIDENCIAS_DIR`, `EVIDENCIAS_BUCKET`, `CAMPUS_CORS_ORIGINS`, `CAMPUS_COOKIE_SECURE`, `CAMPUS_COOKIE_SAMESITE`, `CAMPUS_LOGIN_MAX`, `DATA_RAW_DIR`, `DATA_V1_DIR`) + carga de `.env` |
| `db` | `db.go` | `db.Open(url)` |

Binarios (`apps/api/cmd`): `api`, `migrate` (aplica SQL y luego `accesos.Ensure`), `etl` (carga inicial), `etl-lote` (upsert sin `TRUNCATE`, `-solo-lectura`) y `sectores` (genera `data/v1/zonas_sector.json` y el SQL de la migración 044).

### 2.3 Endpoints (103 registros método+ruta; el contrato servido declara 64 paths)

Fuente: funciones `Registrar*` de `apps/api/internal/handlers/*.go`, llamadas desde `server.New` (`server.go:68-84`). El contrato OpenAPI está repartido en `apps/api/openapi.yaml` (componentes) y `apps/api/openapi/<tag>.yaml`; `GET /api/v1/openapi.yaml` los une al servirlo, y `handlers/contrato_test.go` exige 64 paths.

| Módulo (registrador, archivo) | Método y ruta |
|---|---|
| Sistema (`RegistrarSistema`, `meta.go:89`) | `GET /health` · `GET /api/v1` · `GET /api/v1/openapi.yaml` |
| Geo (`RegistrarGeo`, `geo.go:142`) | `GET /api/v1/geo/resumen` · `/geo/areas` · `/geo/zonas` · `/geo/capas` · `/geo/capas/:capa` · `/geo/edificios` |
| Inventario heredado (`RegistrarInventario`, `inventario.go:56`) | `GET /api/v1/geo/inventario` · `/geo/inventario/fotos/:name` · `/geo/inventario/:capa` |
| Reservas mock (`RegistrarReservas`, `reservas.go:30`) | `GET /api/v1/geo/reservas-mock` |
| Operación (`RegistrarOperacion`, `operacion.go:51`) | `GET /api/v1/operacion/capataces` · `GET, POST /operacion/actividades` · `PATCH /operacion/actividades/:id/asignacion` · `PATCH …/:id/estado` · `POST …/:id/archivar` · `GET …/:id/timeline` · `PATCH …/:id/ficha` |
| Accesos (`RegistrarAccesos`, `producto.go:58`) | `POST, GET, DELETE /api/v1/sesion` · `GET /api/v1/accesos/usuarios` |
| Catálogos (`RegistrarCatalogos`, `producto.go:71`) | `GET, POST /api/v1/catalogos` · `POST /api/v1/catalogos/:id/desactivar` |
| Fichas de área (`RegistrarCatastro`, `producto.go:84`) | `GET, POST /api/v1/catastro/areas` · `PATCH /api/v1/catastro/areas/:id` |
| Atención (`RegistrarAtencion`, `producto.go:97`) | `GET, POST /api/v1/solicitudes` · `GET, POST /ordenes` · `GET, POST /riego` · `GET, POST /evidencias` · `GET /evidencias/:id/archivo` · `GET /reportes/labores` · `POST /ia/sugerir-tipo` |
| Catastro maestro (`RegistrarEjemplares`, `catastro.go:18`) | `GET, POST /api/v1/catastro/zonas-supervision` · `GET /catastro/poligonos` · `GET, POST /catastro/cuadrillas` · `GET, POST /catastro/lugares` · `GET, POST /catastro/especies` · `GET, POST /catastro/ejemplares` · `GET, POST /catastro/ejemplares/:id/codigos` · `GET /catastro/{fauna, puertas, playas, veredas, xerofiticas, jardines-reserva}` |
| Poda y vivero (`RegistrarPoda`, `poda.go:14`) | `GET, POST /api/v1/podas` · `PATCH /podas/:id` · `POST /podas/:id/archivar` · `GET, POST /vivero` · `PATCH /vivero/:id` · `POST /vivero/:id/archivar` · `PATCH /solicitudes/:id` · `PATCH /ordenes/:id` · `POST /operacion/actividades/:id/avances` |
| Inventario 2B (`RegistrarFrente2B`, `frente2b.go:22`) | bajo `/api/v1/inventario`: `tachos` (GET, POST, PATCH `/:id`, DELETE `/:id`, GET `tachos.csv`) · `bebederos` (GET, POST, PATCH, DELETE) · `puntos` (GET, POST, PATCH, DELETE) · `POST formato/puntos` · `reservas` (GET, POST, PATCH, DELETE) · `capas/:capa` (GET, POST) · `capas/:capa/:id` (PATCH, DELETE) · `GET export/:capa` |
| Auditoría (`RegistrarAuditoria`, `auditoria.go:31`) | `POST /api/v1/lotes` · `POST /lotes/:id/revertir` · `POST /auditoria/ediciones` · `GET /auditoria/cambios` · `GET /auditoria/timeline` |
| Importaciones (`RegistrarImportaciones`, `importaciones.go:24`) | `GET /api/v1/importaciones/entidades` · `POST /api/v1/importaciones` · `POST /importaciones/:id/confirmar` |

Todos los `DELETE` son **bajas lógicas** (`activo=false` / `archivada_en`).

**Hallazgo:** el frontend llama a tres rutas que el router **no registra** (hoy dan 404): `POST /api/v1/catastro/areas/:id/baja` (`apps/web/src/panel/catastro.ts:291`), `PATCH /api/v1/catastro/zonas-supervision/:codigo` (`:311`) y `POST …/zonas-supervision/:codigo/baja` (`:315`). El port no debe «arreglarlas» a escondidas: van como lote aparte (§9, lote 24).

### 2.4 Modelos

Salvo los 3 structs de `internal/models/models.go`, los modelos son structs de lectura y escritura propios de cada store, con SQL crudo vía `db.Raw/Exec` (PostGIS: `ST_AsGeoJSON`, `ST_MakeEnvelope`, `catastro_geom_4326()`, `inventario_geom_4326()`). Ids: `BIGSERIAL` (catastro e inventario), **UUID generado por el cliente** (`actividades`, `solicitudes`, `ordenes_servicio`, `riego_registros`, `evidencias`, `podas`, `vivero_registros`, `personal_labor`, `actividad_avances`: permiten reintentos offline idempotentes) y **TEXT** (`cuadrillas.id` = `cua-…`, `capataces.id` = `cap-…`).

### 2.5 Migraciones

`apps/api/migrations/`: 34 archivos (001–044 con huecos, más `019z_origen_ref_unico.sql`) que crean 43 relaciones: `actividad_avances, actividad_eventos, actividades, areas_verdes, asignaciones_poligono, bebederos, cambios, capas_auxiliares, capataces, catalogos, codigos_historicos, cuadrillas, ejemplares, especies, evidencias, fauna, inventario, jardines_reserva, lotes_importacion, lugares, medidas_palmera, ordenes_servicio, permisos, personal_labor, playas_estacionamiento, podas, poligonos_cuadrilla, poligonos_sector_ref, puertas, puntos_pucp, reservas_jardin, riego_registros, roles, sesiones, solicitudes, tachos, usuarios, veredas_riesgo, vivero_catalogo, vivero_registros, xerofiticas, zonas_supervision` y `zonas` (renombrada a `zonas_origen` por la 012 y reemplazada por la **vista** `zonas` sobre `poligonos_cuadrilla`), además de `schema_migrations`. Son idempotentes (`IF NOT EXISTS`, `ON CONFLICT DO NOTHING`, bloques `DO $$`) y dejan las correcciones de datos en `cambios` (014, 018, 019z, 034). **Excepción histórica:** `006_limpieza.sql` borra filas de prueba («Prueba de pin»). Ya está aplicada y no se vuelve a ejecutar, pero confirma que la regla «sin DELETE» debe revisarse en código.

### 2.6 Autenticación y roles

- `POST /api/v1/sesion {usuario, clave}` → bcrypt → token opaco en la cookie `cv_sesion` (HttpOnly; `Secure` y `SameSite` salen del entorno) y fila en `sesiones(token_hash, usuario_id, expires_at)` con 12 h de vida. El middleware de `server.go:44-53` resuelve la cookie en cada petición.
- Roles: `usuarios.rol TEXT CHECK (rol IN ('capataz','coordinacion','jefatura','admin'))` (005), catálogo `roles(codigo PK, nombre, orden)` con nombres Capataz / Coordinación / Jefatura / Administración (008) y FK `permisos.rol → roles.codigo`.
- Permisos: `accesos.Matriz` en código (`accesos.go:41-46`). `Ensure` **borra y reescribe `permisos` en cada arranque** (`DELETE FROM permisos`, `accesos.go:99`), así que la tabla no es configurable (choca con RF-03 y RNF-05).
- Semillas: 6 cuentas ficticias (`norte`, `sur`, `riego` como capataces; `coordinacion`, `jefatura`, `admin`) con la clave `CAMPUS_DEV_PASSWORD`.
- Literales de rol en Go fuera de `accesos`: `operacion/model.go:6-10`, `evidencias/subir.go:255-262` (`rolEvento`), `handlers/operacion.go:193`, `handlers/producto.go:33`.
- No hay alta, edición ni desactivación de usuarios por API (solo `GET /accesos/usuarios`).

### 2.7 Evidencias y archivos

`POST /api/v1/evidencias` (multipart) → `evidencias.Guardar`: verifica permiso sobre la labor, detecta el MIME real, calcula SHA-256 (mismo id con otro hash → 409), filtra EXIF (fecha, lat, lon, orientación), guarda el binario con `blobs.Store` (S3 si `EVIDENCIAS_BUCKET` carga credenciales, si no disco en `EVIDENCIAS_DIR=/data/evidencias`) y registra la fila en `evidencias` más el evento `evidencia` en `actividad_eventos`. `GET /evidencias/:id/archivo` sirve el binario con `Cache-Control` corto (commit `6a80853`). Límite de 10 MB en nginx (`client_max_body_size 10m`).

### 2.8 ETL e importación

- Carga inicial `cmd/etl` → `etl.Run` → `etl.Load`, que hace `TRUNCATE areas_verdes, poligonos_cuadrilla, capas_auxiliares RESTART IDENTITY CASCADE` (`etl/load.go:47`). Se protege con `negarSiHayDependientes`, que solo mira `ejemplares` y `asignaciones_poligono`; `etl/inventario.go:137` hace `TRUNCATE inventario`.
- El **entrypoint** del contenedor corre `etl` si falta `/data/.etl-done` (`apps/api/docker-entrypoint.sh`). **Riesgo:** si el volumen `/data` cambia o pierde ese archivo y `ejemplares` está vacía, el catastro editado se trunca. Hay que tratarlo en el corte (§8, R-01).
- `cmd/etl-lote`: upsert por `feature_id`/`origen_ref`, sin `TRUNCATE`.
- Importación HTTP: `POST /importaciones` (vista previa de Excel/CSV, guarda el archivo en `lotes_importacion.contenido`), `POST /importaciones/:id/confirmar`, reversión por lote (`POST /lotes/:id/revertir`) con `cambios`. 22 entidades importables (`etl/importar.go:15-38`).

### 2.9 Tests

Tests unitarios en `accesos, atencion, auditoria, blobs, capas, catastro, config, etl, evidencias, geojson, handlers, migrate, operacion, server`. Los de integración (`*_db_test.go`, `migrate/apply_test.go`, `auditoria/*_test.go`) usan `MIGRATE_TEST_URL` (14 referencias). CI: `.github/workflows/ci.yml` (Go 1.22.2 con `gofmt` + `go test`; web con lint, test y build; `docker build` de ambas imágenes; `deploy` manual a Learner Lab).

### 2.10 Dockerfile y entrypoint

`apps/api/Dockerfile`: `golang:1.22-bookworm` construye `api`, `migrate` y `etl`; runtime `debian:bookworm-slim` **como root**; copia `migrations`, `openapi.yaml`, `data/raw`, `data/osm`, `data/mocks`; `API_ADDR=:8091`; `ENTRYPOINT /entrypoint.sh` (`mkdir /data/evidencias /data/v1` → `migrate` (si falla, sale con un mensaje explícito) → `etl` una vez → `exec api`). Contexto de build: la raíz del repo (`docker-compose.yml`, `scripts/deploy-learner-lab.sh:34`).

---

## 3. Mapeo módulo por módulo hacia `init/backend`

Rutas relativas a `backend/app/internal/`. La **entidad de dominio usa el vocabulario del núcleo v0.2** y el **modelo de persistencia apunta a nuestra tabla real** (el mapper los une). Cada group se monta en `/areas-verdes/v1` y en el alias `/api/v1` (§6.1).

### 3.1 Transversal

| Actual | Destino en init/backend |
|---|---|
| `config/config.go` | `shared/config/config.go`: se amplía `Config` con `Database.URL` (`DATABASE_URL`, prioridad sobre las variables sueltas), `Server.TrustedProxies`, `Seguridad{CORSOrigins, CookieSecure, CookieSameSite, LoginMax}`, `Evidencias{Dir, Bucket}`, `Datos{RawDir, V1Dir, EdificiosPath, ReservasPath, FotosDir}`, `Migraciones{Dir}`, `Accesos{DevPassword}`, `Swagger{Enabled}`. Se conserva el `.env` con `godotenv` |
| `db/db.go` | `persistence/database/database.go` (`NewConnection` usa `DATABASE_URL` si viene; pool configurable) |
| `migrate/migrate.go` + `cmd/migrate` | `persistence/database/migrate.go` (`Apply`, `splitSQL`, `ayudaMigracion`, igual tabla `schema_migrations`) + `cmd/migrate/main.go` que usa `ioc.BuildContainer()` |
| `server/server.go` | `presentation/routes/routes.go` (`Setup`: middlewares globales, `NoRoute`, montaje doble) + provider de Gin en `cmd/ioc/container.go` con `SetTrustedProxies` |
| `server/bitacora.go` | `presentation/middleware/bitacora.middleware.go` (usa el `zerolog.Logger` inyectado) |
| `handlers/seguridad.go` (`CORS`, `ConCookie`, `LimiteLogin`) | `presentation/middleware/cors.middleware.go`, `cookie.middleware.go`, `limite_login.middleware.go`; `accesos/limite.go` → `infrastructure/ratelimit/memoria.go` con contrato `application/contracts/limitador.contract.go` |
| middleware de sesión (`server.go:44-53`) | `presentation/middleware/auth.middleware.go` (cookie → `usecases.ISesionUseCase.Resolver` → `c.Set("usuario")`) |
| `exige(c, accion)` (`handlers/producto.go:39`) | `presentation/middleware/permiso.middleware.go` → `RequierePermiso("registrar")`, más el servicio `application/services/permisos.service.go` |
| respuestas `{"error": ...}` repartidas | `domain/errors/*.errors.go` (errores tipados) + `presentation/middleware/errores.middleware.go` (traduce `c.Errors` a HTTP con el **mismo** cuerpo `{"error": "<mensaje en español>"}`) |
| `geojson/` | `domain/entities/feature_collection.entity.go` (tipos) y `application/dto/geojson.dto.go`; `format.go` → `persistence/mapper/geojson.mapper.go` |
| `log.Printf` | `zerolog.Logger` inyectado |

### 3.2 Módulos de negocio

| Módulo actual | Entidad (`domain/entities`) · enums | DTO · contrato (`application`) | Caso de uso / servicio | Modelo (`persistence/models`) · mapper · repositorio | Controller · request · group | Middleware |
|---|---|---|---|---|---|---|
| **Sistema** `handlers/meta.go`, `health.go` | — | `dto/health.dto.go` · `contracts/salud.contract.go` (`PostGISVersion`) | `usecases/salud.usecase.go` | `repository/postgres/salud.repository.go` | `health.controller.go` (existente, ampliado con `database`, `postgis`), `meta.controller.go` · `health.group.go`, `meta.group.go` (índice y alias `openapi.yaml`), `legado.group.go` (`GET /health` en la raíz con el cuerpo actual) | — |
| **Accesos** `accesos/*`, `handlers/producto.go` (`Sesion`) | `usuario.entity.go`, `sesion.entity.go`, `rol.entity.go`, `permiso.entity.go` · `enums/rol.enum.go` (códigos) | `dto/sesion.dto.go` (`UsuarioSesionDTO{id, usuario, nombre, rol, rol_nombre, capataz_id}`) · `contracts/usuario.contract.go`, `sesion.contract.go`, `permiso.contract.go`, `hasher.contract.go` | `usecases/sesion.usecase.go` (Login, Logout, Actual, Resolver), `usecases/usuario.usecase.go` (Listar), `usecases/semilla_accesos.usecase.go` (`Ensure`) · `services/permisos.service.go` | `usuario.model.go` (`usuarios`), `sesion.model.go` (`sesiones`), `permiso.model.go` (`permisos`), `rol.model.go` (`roles`) · `usuario.mapper.go` · `usuario.repository.go`, `sesion.repository.go`, `permiso.repository.go` · `infrastructure/seguridad/bcrypt.hasher.go` | `sesion.controller.go`, `usuario.controller.go` · `requests/sesion.request.go` · `sesion.group.go` (`/sesion`), `accesos.group.go` (`/accesos/usuarios`) | `limite_login`, `cookie` |
| **Catálogos** `catalogos/store.go` | `catalogo.entity.go` · `enums/clase_catalogo.enum.go` | `dto/catalogo.dto.go` · `contracts/catalogo.contract.go` | `usecases/catalogo.usecase.go` | `catalogo.model.go` (`catalogos`) · `catalogo.mapper.go` · `catalogo.repository.go` | `catalogo.controller.go` · `catalogo.request.go` · `catalogo.group.go` | `permiso(catalogos)` en POST |
| **Geo (lectura)** `catastro/store.go` (`Areas`, `Zonas`, `Capa`, `Resumen`, `Capas`), `handlers/geo.go` | `area_verde.entity.go`, `poligono_sector.entity.go` (el «sector» del núcleo), `capa_referencia.entity.go` | `dto/geo.dto.go` (`FiltroGeoDTO{bbox, limit}`, `ResumenDTO`) · `contracts/geo.contract.go`, `contracts/archivo_estatico.contract.go` (edificios OSM) | `usecases/geo.usecase.go` | `area_verde.model.go` (`areas_verdes`), `poligono_cuadrilla.model.go`, `capa_auxiliar.model.go` (vienen de `internal/models/models.go`) · `geo.repository.go` (SQL PostGIS crudo con `db.Raw`) · `infrastructure/archivos/geojson_estatico.go` | `geo.controller.go` · `geo.request.go` (query) · `geo.group.go` (`/geo/...`) | — |
| **Fichas de área** `catastro/store.go` (`Fichas`, `ActualizarFicha`, `CrearSinGeom`), `handlers/producto.go` (`Fichas`) | `area_verde.entity.go` | `dto/area_verde.dto.go` · `contracts/area_verde.contract.go` | `usecases/area_verde.usecase.go` (escribe en `cambios` con `services/auditoria.service.go`) | `area_verde.model.go` · `area_verde.mapper.go` · `area_verde.repository.go` | `area_verde.controller.go` · `area_verde.request.go` · `catastro.group.go` (`/catastro/areas`) | `permiso` |
| **Catastro maestro** `catastro/modelo.go`, `sector.go`, `handlers/catastro.go` | `zona_supervision.entity.go`, `cuadrilla.entity.go` (≈ `personal`/capataz del núcleo), `lugar.entity.go`, `especie.entity.go`, `ejemplar.entity.go`, `codigo_historico.entity.go` · `enums/tipo_vegetacion.enum.go`, `enums/sector.enum.go` | `dto/catastro.dto.go` · `contracts/zona_supervision.contract.go`, `cuadrilla.contract.go`, `lugar.contract.go`, `especie.contract.go`, `ejemplar.contract.go` | `usecases/zona_supervision.usecase.go`, `lugar.usecase.go`, `especie.usecase.go`, `ejemplar.usecase.go` (incluye `Recodificar`); las validaciones de `modelo.go` (`ValidarPuntoCampus`, `NormalizarNombre`…) pasan a métodos de las entidades | modelos por tabla · mappers · repositorios homónimos | `catastro.controller.go` (o uno por recurso) · `catastro.request.go` · `catastro.group.go` (`/catastro/...`) | `permiso` en POST |
| **Inventario heredado** `inventario/store.go`, `handlers/inventario.go` | `elemento_inventario.entity.go` | `dto/inventario.dto.go` · `contracts/inventario.contract.go`, `contracts/foto.contract.go` | `usecases/inventario.usecase.go` | `inventario.model.go` · `inventario.repository.go` · `infrastructure/archivos/fotos_disco.go` (`drive_fotos`) | `inventario.controller.go` · `inventario.group.go` (`/geo/inventario...`) | — |
| **Reservas mock** `handlers/reservas.go` | — | `dto/reserva_mock.dto.go` · `contracts/reservas_mock.contract.go` | `usecases/reservas_mock.usecase.go` | `infrastructure/archivos/reservas_mock.go` | `reservas_mock.controller.go` · `geo.group.go` | — |
| **Inventario 2B** `capas/store.go`, `handlers/frente2b.go` | `tacho.entity.go`, `bebedero.entity.go`, `punto_pucp.entity.go`, `reserva_jardin.entity.go`, `ficha_capa.entity.go` | `dto/inventario_campo.dto.go` · un contrato por entidad | `usecases/tacho.usecase.go`, `bebedero.usecase.go`, `punto.usecase.go`, `reserva_jardin.usecase.go`, `ficha_capa.usecase.go` (PATCH parcial con `clavesPatch`) · `services/exportacion_csv.service.go` | modelos por tabla · mappers · repositorios | `inventario_campo.controller.go` · `inventario_campo.request.go` · `inventario_campo.group.go` (`/inventario/...`) | `permiso` |
| **Operación / labores** `operacion/*`, `handlers/operacion.go`, `CrearAvance` en `poda.go` | `intervencion.entity.go` (≈ núcleo `intervencion`; `ID uuid` ≈ `uuid_cliente`), `actividad_evento.entity.go`, `avance.entity.go`, `personal_labor.entity.go` (≈ `intervencion_personal`), `capataz.entity.go` · `enums/estado_intervencion.enum.go`, `tipo_evento.enum.go`, `ejecutor.enum.go`; errores `ErrNoEncontrada`, `ErrValidacion`, `ErrPuntoFuera` | `dto/intervencion.dto.go` (`CrearIntervencionDTO`, `FiltroIntervencionesDTO`, `TimelineDTO`) · `contracts/intervencion.contract.go`, `actividad_evento.contract.go` | `usecases/intervencion.usecase.go` (Listar, Crear idempotente por UUID, Asignar, CambiarEstado, Archivar con motivo, Timeline, GuardarFicha, CrearAvance, Capataces) | `actividad.model.go` (`actividades`, `geom` leído con `ST_AsGeoJSON`), `actividad_evento.model.go`, `actividad_avance.model.go`, `personal_labor.model.go`, `capataz.model.go` · `intervencion.mapper.go` · `intervencion.repository.go` | `intervencion.controller.go` · `intervencion.request.go` · `operacion.group.go` (`/operacion/...`) | `auth` (el capataz solo ve lo suyo), `permiso(registrar|validar)` |
| **Atención: solicitudes y órdenes** `atencion/store.go`, `handlers/producto.go` (`Atencion`), `EditarSolicitud`/`EditarOrden` de `poda.go` | `solicitud.entity.go`, `servicio_tercerizado.entity.go` (nuestra `ordenes_servicio`) · `enums/estado_solicitud.enum.go`, `fuente.enum.go` | `dto/solicitud.dto.go`, `dto/servicio_tercerizado.dto.go` · contratos homónimos | `usecases/solicitud.usecase.go`, `servicio_tercerizado.usecase.go` | `solicitud.model.go`, `orden_servicio.model.go` · mappers · repositorios | `solicitud.controller.go`, `servicio_tercerizado.controller.go` · requests · `atencion.group.go` (`/solicitudes`, `/ordenes`) | `permiso(solicitudes)` |
| **Riego** `atencion/riego.go` + parte de `store.go` | `turno_riego.entity.go` | `dto/riego.dto.go` · `contracts/riego.contract.go` | `usecases/riego.usecase.go` (`ValidarRiego`; el capataz solo ve su equipo) | `riego_registro.model.go` · `riego.mapper.go` · `riego.repository.go` | `riego.controller.go` · `riego.request.go` · `atencion.group.go` (`/riego`) | `auth`, `permiso` |
| **Poda y vivero** `atencion/poda_vivero.go`, `handlers/poda.go` | `poda.entity.go`, `vivero.entity.go` | `dto/poda.dto.go`, `dto/vivero.dto.go` · contratos | `usecases/poda.usecase.go`, `vivero.usecase.go` | `poda.model.go`, `vivero_registro.model.go`, `vivero_catalogo.model.go` · mappers · repositorios | `poda.controller.go`, `vivero.controller.go` · requests · `poda.group.go`, `vivero.group.go` | `permiso` |
| **Evidencias** `evidencias/subir.go`, `handlers/evidencias.go`, `blobs/*`, evidencias en `atencion/store.go` | `evidencia.entity.go` (con `EventoID` opcional, §4.3) · `enums/tipo_evidencia.enum.go` | `dto/evidencia.dto.go` · `contracts/evidencia.contract.go`, `contracts/almacen_archivos.contract.go` (`Put`, `Open`) | `usecases/evidencia.usecase.go` (Subir idempotente por SHA-256, Listar, Abrir) · `services/exif.service.go` (filtro EXIF y MIME real) | `evidencia.model.go` · `evidencia.mapper.go` · `evidencia.repository.go` · `infrastructure/storage/s3.storage.go`, `disco.storage.go`, `storage/factory.go` (lógica de `blobs.Open`) | `evidencia.controller.go` · `evidencia.request.go` (multipart) · `evidencia.group.go` | `auth`, `permiso` |
| **Reportes e indicadores** `atencion/export.go`, `indicadores.go`, `Reporte`/`Exportar` | `reporte.entity.go` (`Fila`) | `dto/reporte.dto.go` (`FiltroReporteDTO`) · `contracts/reporte.contract.go`, `contracts/exportador.contract.go` | `usecases/reporte.usecase.go` | `reporte.repository.go` (consulta por streaming) · `infrastructure/exportacion/csv.exporter.go`, `excel_xml.exporter.go` | `reporte.controller.go` · `reporte.group.go` (`/reportes/labores`) | `permiso(reportes)` |
| **IA** `atencion/ia.go` | — | `dto/sugerencia.dto.go` · `contracts/sugeridor.contract.go` | `services/sugerencia_tipo.service.go` (heurística local; se deja como **infraestructura** si algún día llama a un servicio externo, RNF-13) | — | `ia.controller.go` · `ia.request.go` · `ia.group.go` | `auth` |
| **Auditoría y lotes** `auditoria/store.go`, `handlers/auditoria.go` | `cambio.entity.go` (≈ núcleo `auditoria`), `lote_importacion.entity.go` · `enums/accion_cambio.enum.go` | `dto/auditoria.dto.go` · `contracts/cambio.contract.go`, `lote.contract.go` | `usecases/auditoria.usecase.go` (Editar, Historial, Timeline), `usecases/lote.usecase.go` (Importar, Revertir) · `services/auditoria.service.go` (el resto de casos de uso lo usan para escribir antes y después) | `cambio.model.go`, `lote_importacion.model.go` · mappers · `cambio.repository.go`, `lote.repository.go` (las funciones `aplicar*` por entidad quedan aquí, porque conocen las tablas) | `auditoria.controller.go`, `lote.controller.go` · requests · `auditoria.group.go` (`/auditoria/...`, `/lotes/...`) | `permiso(validar)` |
| **Importaciones** `etl/importar.go`, `lote*.go`, `formato.go`, `handlers/importaciones.go` | `importacion.entity.go` (vista previa, filas, avisos) · `enums/entidad_importable.enum.go` (las 22) | `dto/importacion.dto.go` · `contracts/lector_archivo.contract.go`, `contracts/carga.contract.go` | `usecases/importacion.usecase.go` (Entidades, Previsualizar, Confirmar) | `infrastructure/etl/` (lectura y normalización puras: `lote_leer.go`, `formato.go`, `normalize.go`, `anon.go`, `ficticio.go`, `inventario.go` sin BD) · `repository/postgres/carga_*.repository.go` (escrituras de `lote.go`, `load.go`, `write.go`, `capas2b.go`, `atencion_import.go`, `sector.go`) | `importacion.controller.go` · `importacion.group.go` | `permiso` |
| **CLIs ETL** `cmd/etl`, `cmd/etl-lote`, `cmd/sectores` | — | — | `usecases/carga_inicial.usecase.go`, `carga_lote.usecase.go`, `sectores.usecase.go` | los mismos repositorios `carga_*` | `cmd/etl/main.go`, `cmd/etl-lote/main.go`, `cmd/sectores/main.go` (usan `ioc.BuildContainer`) | — |

Extensiones al layout de init que requiere el port: `cmd/migrate`, `cmd/etl`, `cmd/etl-lote` y `cmd/sectores` como binarios extra. Son paquetes `main` en subcarpetas de `cmd/`, así que conviven con `cmd/main.go` sin tocarlo.

---

## 4. Modelo de datos y migración sin pérdida de datos

### 4.1 Núcleo v0.2 frente a nuestro esquema

| Núcleo v0.2 | Nuestro equivalente | Diferencias relevantes |
|---|---|---|
| `rol` | `roles` (008) | PK `codigo TEXT` en vez de `id bigint`; sin `descripcion` ni `activo` |
| `permiso`, `rol_permiso` | `permisos(rol, accion)` | Acción como texto, sin maestro de permisos; se reescribe al arrancar |
| `usuario` | `usuarios` | Faltan `email`, `apellidos`, `debe_cambiar_password`, `ultimo_acceso`, `telefono`, `tipo`, `personal_id`; `rol` es texto; tenemos `capataz_id`, `cuadrilla_id` |
| `personal`, `intervencion_personal` | `cuadrillas` (equipos ficticios) y `personal_labor` (nombres por labor) | No hay maestro de personal |
| `auditoria` | `cambios` (+ `lotes_importacion`) | Lo nuestro es más rico (antes/después, reversión); sin `ip` |
| `sede`, `cuartel` | — | Campus Pando implícito; v2 RF-06: cuarteles solo como referencia histórica |
| `sector` | `poligonos_cuadrilla.sector` (slug), `zonas_supervision` (Z1–Z4), `cuadrillas` | No hay tabla `sector`; la geometría es PostGIS |
| `lugar` | `lugares` | `lat/lon double` + `zona_supervision_id` + `nombre_norm` único |
| `area_verde` | `areas_verdes` | `MultiPolygon 4326`, `feature_id`, `codigo` único, `uso`, riego |
| `especie`, `ejemplar` | `especies`, `ejemplares` (+ `codigos_historicos`, `medidas_palmera`) | Historial de códigos en tabla propia (núcleo: `codigo_anterior`) |
| `clase_actividad`, `tipo_actividad`, `estado_atencion`, `prioridad`, `fuente_solicitud` | `catalogos(clase=…)` + `CHECK`s de texto | Catálogo genérico; sin `color`, `es_final`, `frecuencia_dias`, `atributos_schema` |
| `insumo`, `consumo_insumo`, `plaga`, `producto_fitosanitario`, `plaga_especie`, `producto_plaga`, `punto_acopio`, `ficha_tecnica_poda`, `checklist_item` | — | No existen (RF-10 insumos, fitosanitario y checklist quedan pendientes) |
| `empresa` | `ordenes_servicio.empresa` (texto) | Sin maestro de empresas |
| `solicitud`, `referencia_externa` | `solicitudes` (UUID, `codigo_externo` único, `fuente` texto, `actividad_id`) | Una sola referencia externa por solicitud |
| `servicio_tercerizado` | `ordenes_servicio` (UUID, `actividad_id NOT NULL`) | Orden por labor, no por solicitud |
| `intervencion` | `actividades` (UUID, `geom Point`, `tipo`, `estado`, `assigned_capataz_id`, `cuadrilla_id`, `ejecutor propia/tercerizada`, `archivada_en`, `origen_ref`…) | El UUID cumple el papel de `uuid_cliente`; no hay `sync_estado`; estados distintos (`pendiente, en_proceso, bloqueada, cerrada, cancelada, sin_estado`); no existe `Ejecutado` |
| `actividad_evento` | `actividad_eventos` | Tenemos `actor_rol`, `capataz_id`, `nota`, `usuario_id`; faltan `uuid_cliente` y `sync_estado` |
| `evidencia` | `evidencias` (UUID, `ruta`, `sha256`, `lat/lon`, `exif`, `orden_id`) | **Falta `evento_id`** (v2 RF-19/RF-31: evidencia por evento) |
| `turno_riego` | `riego_registros` | Sin `horas` ni `cobertura_pct`; con `zona_supervision_id`, `ciclo`, `superficie_m2` |

Tablas solo nuestras: `capataces` (stub heredado), `zonas_supervision`, `poligonos_cuadrilla`, `zonas_origen` + vista `zonas`, `asignaciones_poligono`, `poligonos_sector_ref`, `capas_auxiliares`, `inventario`, `fauna`, `puertas`, `playas_estacionamiento`, `veredas_riesgo`, `xerofiticas`, `jardines_reserva`, `tachos`, `bebederos`, `puntos_pucp`, `reservas_jardin`, `podas`, `vivero_catalogo`, `vivero_registros`, `actividad_avances`, `sesiones`, `catalogos`, `lotes_importacion`.

**No hay colisión de nombres**: el núcleo usa singulares (`usuario`) y nosotros plurales (`usuarios`). Por eso ejecutar `schema_nucleo_v0.2.sql` sobre nuestra BD *no fallaría*: crearía 36 tablas vacías en paralelo con semillas propias, es decir, **dos fuentes de verdad**. Por eso no se debe hacer.

### 4.2 Decisión: nuestro esquema sigue siendo la fuente de verdad; el núcleo es el modelo de dominio

Justificación:

1. **Datos cargados y editados** viven en nuestras tablas (521 áreas, polígonos, inventario, labores, evidencias, `cambios`, lotes). Copiarlos al núcleo sería un big-bang con conversión de ids (UUID → `bigint identity`) en **9 tablas con FK cruzadas**, reescritura de `cambios.entidad_id` y pérdida de la reversión por lotes, que aplica el «antes» sobre las tablas actuales (`auditoria/store.go`: `aplicarArea`, `aplicarActividad`, `aplicarCatalogo`).
2. **El núcleo no tiene PostGIS** (solo `latitud/longitud numeric`, sin polígonos), y el mapa (RF-29, RF-33, RF-34) depende de `MultiPolygon`, bbox e índices GIST.
3. **Offline:** nuestros UUID generados por el cliente ya dan la idempotencia que el núcleo busca con `uuid_cliente`. Cambiar la PK rompería la cola IndexedDB de la PWA (`apps/web/src/offline/queue.ts`).
4. **init/backend no define migraciones**: su `db/` solo corre en initdb con el volumen vacío, algo inservible sobre una BD desplegada. Nuestro runner es incremental, transaccional y ya está probado (`migrate/apply_test.go`).
5. Sí seguimos a init en lo que define: **capas, IoC, nombres de dominio** (`Intervencion`, `ActividadEvento`, `Solicitud`, `ServicioTercerizado`, `TurnoRiego`…) y **`db/` como lugar del esquema** (las migraciones se mueven a `db/migrations/` con **los mismos nombres de archivo**, así `schema_migrations.version` sigue igual).

Reglas para los modelos GORM: `TableName()` apunta a nuestra tabla; **prohibido `AutoMigrate`** (se verifica en CI con `grep`); las columnas `geometry` no se mapean como campo GORM, se leen con `ST_AsGeoJSON(geom)` y se escriben con `catastro_geom_4326(?)`/`ST_SetSRID(ST_MakePoint(?, ?), 4326)` en SQL crudo dentro del repositorio.

### 4.3 Migraciones aditivas hacia el núcleo (045+)

> **Renumeración (fase 2).** La `045` se usó en la rama `fix/seguridad-datos-permisos` para las etiquetas de rol v2 (`045_roles_v2_etiquetas.sql`, con historial en `cambios`, columnas `roles.descripcion`/`activo` y el permiso `evidencias` de jefatura), y la `046` en el lote 6 para la FK `usuarios.rol → roles.codigo` (`046_usuarios_rol_fk.sql`). Las migraciones del núcleo de esta tabla se corren un número: `047_evidencia_evento.sql`, `048_sync_offline.sql`, `049_estados_nucleo.sql` y `050_vistas_nucleo.sql` (lote 24).
>
> **Nota sobre migración 046 (`046_usuarios_rol_fk.sql`):** La ejecución consecutiva de `NOT VALID` y `VALIDATE CONSTRAINT` dentro de la misma transacción en la migración 046 es intencional y completamente inofensiva en bases de datos con el volumen actual; se mantiene sin editar para preservar los checksums y el estado ya aplicado en entornos de prueba y producción.

Todas idempotentes (`IF NOT EXISTS`, `ON CONFLICT DO NOTHING`, `DO $$ … IF NOT EXISTS (SELECT 1 FROM pg_constraint …)`), columnas nuevas **nulas o con DEFAULT** (para que la imagen anterior siga funcionando si hay rollback), y todo `UPDATE` de datos con su antes/después en `cambios`. Nada de `DROP TABLE`, `DROP COLUMN`, `TRUNCATE` ni `DELETE` de datos cargados.

| Archivo | Contenido | Motivo |
|---|---|---|
| `045_roles_v2.sql` | Nombres v2 en `roles`, `descripcion` y `activo`, FK `usuarios.rol → roles.codigo` en lugar del `CHECK` (§5) | RF-02, RF-03, RNF-05 |
| `046_evidencia_evento.sql` | `ALTER TABLE evidencias ADD COLUMN IF NOT EXISTS evento_id BIGINT REFERENCES actividad_eventos(id)` + índice | RF-19, RF-31 |
| `047_sync_offline.sql` | `actividad_eventos.uuid_cliente UUID` (índice único parcial `WHERE uuid_cliente IS NOT NULL`) y `sync_estado TEXT NOT NULL DEFAULT 'sincronizado'` en `actividades`, `actividad_eventos` y `riego_registros` | RF-12, RNF-01 (núcleo `uuid_cliente`/`sync_estado`) |
| `048_estados_nucleo.sql` | En `catalogos` (clase `estado`): `ejecutado` y `archivado` si faltan, y columnas `color`/`es_final` **nullable** en `catalogos` rellenadas con los colores del núcleo | RF-16, D-08 |
| `049_vistas_nucleo.sql` | `CREATE SCHEMA IF NOT EXISTS nucleo;` y vistas de solo lectura `nucleo.intervencion`, `nucleo.actividad_evento`, `nucleo.evidencia`, `nucleo.solicitud`, `nucleo.servicio_tercerizado`, `nucleo.turno_riego`, `nucleo.area_verde`, `nucleo.ejemplar`, `nucleo.usuario`, `nucleo.rol` con los nombres de columna del núcleo (p. ej. `intervencion.uuid_cliente = actividades.id`, `latitud = ST_Y(geom)`, estado traducido a los nombres del núcleo, `Archivado` si `archivada_en IS NOT NULL`) | Reportes como vistas (README §6); SQL del equipo contra el vocabulario del núcleo, sin tocar `public` |

Los datos maestros que faltan (insumos, fitosanitario, empresas, sede, checklist) **no son parte del port**: se crean después, cuando exista su historia, en tablas nuevas que siguen el DDL del núcleo en `public` (sin choque de nombres) y con importación masiva e historial.

### 4.4 Procedimiento sobre la BD PostGIS desplegada

Se ejecuta en el lote de corte (§9, lote 22) y antes de cada lote que agregue una migración.

1. **Congelar el esquema:** confirmar que `main` desplegado = `115a56d` (o su sucesor) y listar `SELECT version FROM schema_migrations ORDER BY 1;` → `antes-versiones.txt` (hoy deberían ser las 34).
2. **Backup:** en la EC2 (vía SSM), `bash scripts/backup-postgis.sh /opt/campus/data/backups/pre-backend-v2-<fecha>.dump` (`pg_dump -Fc --no-owner --no-acl`, incluye `CREATE EXTENSION`), copiarlo fuera de la instancia (S3 privado o descarga) y tomar un **snapshot EBS** del volumen de datos (`docs/OPERACION.md`). Respaldar además `/opt/campus/data/app/evidencias` si no hay bucket.
3. **Conteo de referencia** (exacto, todas las tablas):
   ```sql
   SELECT table_name,
          (xpath('/row/c/text()', query_to_xml(format('SELECT count(*) AS c FROM %I.%I', table_schema, table_name), false, true, '')))[1]::text::bigint AS filas
   FROM information_schema.tables
   WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
   ORDER BY 1;
   ```
   → `conteos-antes.tsv`. Complementar con `scripts/counts.sh` (geometrías con SRID 4326).
4. **Ensayo en copia:** restaurar el dump en un PostGIS desechable (patrón de `scripts/probar-restauracion.sh` / `restore-postgis.sh`), correr ahí el `migrate` de la **imagen nueva**, volver a contar y comparar: **cada tabla existente debe tener exactamente las mismas filas**; solo pueden aparecer tablas, columnas o vistas nuevas. Correr también el arnés de paridad (§9, lote 4) con la API vieja y la nueva sobre esa copia.
5. **Proteger el ETL:** confirmar que `/opt/campus/data/app/.etl-done` existe; el entrypoint nuevo además **no ejecuta `etl` si `areas_verdes` tiene filas** (chequeo en BD, no por archivo, §8 R-01).
6. **Desplegar** la imagen nueva (tag inmutable por SHA, §7). El entrypoint aplica solo las migraciones pendientes (045+), cada una en su transacción.
7. **Verificar:** `/health` → `"status":"ok"`; `GET /areas-verdes/v1/health`; login con una cuenta semilla; conteos después = conteos antes en las tablas existentes; `schema_migrations` = antes + las nuevas; revisar `cambios` recientes (solo los de 045/048).
8. **Rollback:**
   - *Aplicación:* volver a desplegar el tag anterior de la imagen de `apps/api`. Como las migraciones son aditivas y compatibles hacia atrás, el binario viejo funciona con el esquema nuevo. **Aviso:** el `accesos.Ensure` viejo volverá a reescribir `permisos` desde su `Matriz`, y cualquier permiso configurado con el backend nuevo se perdería (sigue en el dump).
   - *Datos (último recurso):* `pg_restore` del dump en una BD nueva y cambio de `DATABASE_URL`. Se pierde lo registrado después del backup, así que solo se usa ante corrupción y con la decisión del responsable.
   - Las migraciones nuevas **no tienen «down»**: revertir una columna sería un `DROP COLUMN`. Si alguna es incorrecta, se corrige con otra migración aditiva.

---

## 5. Nombres de rol del backlog v2

Fuentes: backlog v2 (`docs/fuente/extraccion/backlog_v2.txt`, líneas 6-8): *RF-02 «Roles diferenciados y configurables para Capataz (campo), Ingeniería/Coordinación y Jefatura (oficina). En el MVP los operarios no requieren cuenta»*; actor de HU-02, HU-19 y RNF-05 = «Administrador del sistema»; RNF-03/04 = «Administrador técnico»; la mayoría de HU usan «Supervisor»; RF-27 «Ingeniero de campo de empresa proveedora» está **fuera del MVP**.

| Código técnico (BD/API, **se conserva**) | `roles.nombre` actual (008) | Nombre v2 (nuevo `roles.nombre`) | Núcleo v0.2 (`rol.nombre`) | Permisos actuales (`accesos.Matriz`) |
|---|---|---|---|---|
| `capataz` | Capataz | **Capataz** | Capataz | consultar, registrar |
| `coordinacion` | Coordinación | **Ingeniería/Coordinación** | Ingeniero/Coordinador | consultar, registrar, validar, solicitudes, reportes |
| `jefatura` | Jefatura | **Jefatura** | Jefe de Sección | consultar, validar, reportes, solicitudes |
| `admin` | Administración | **Administrador** | Administrador | consultar, registrar, validar, reportes, catalogos, solicitudes |

Notas de interpretación:
- «Supervisor» en v2 es un actor genérico (Jefatura + Ingeniería/Coordinación), **no un rol**. «Usuario del sistema/autorizado» = cualquier rol. «Administrador técnico» opera la infraestructura (no es un rol de la app).
- El núcleo v0.2 usa nombres más antiguos («Jefe de Sección», «Ingeniero/Coordinador»). Prevalece el backlog v2 por ser el más reciente; conviene que el equipo actualice la semilla del núcleo (se deja anotado en la sección de riesgos).

**Por qué no se renombran los códigos:** `usuarios.rol`, `permisos.rol` (FK), `actividad_eventos.actor_rol` (**historial inmutable** de trazabilidad), `cambios` y `sesiones` activas, además de 20+ literales en el frontend (`types.ts`, `session.ts`, `App.tsx`, `producto.ts`…). Renombrarlos obligaría a reescribir historia, contra «nunca borrar ni alterar datos cargados».

**`045_roles_v2.sql` (esquema del contenido):**
```sql
ALTER TABLE roles ADD COLUMN IF NOT EXISTS descripcion TEXT;
ALTER TABLE roles ADD COLUMN IF NOT EXISTS activo BOOLEAN NOT NULL DEFAULT TRUE;

-- Antes/después en cambios, solo para las filas que cambian (idempotente).
INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
SELECT 'roles', r.codigo, 'edicion',
       jsonb_build_object('nombre', r.nombre),
       jsonb_build_object('nombre', v.nombre, 'motivo', 'nombres del backlog v2 (RF-02)')
FROM roles r
JOIN (VALUES ('capataz','Capataz'), ('coordinacion','Ingeniería/Coordinación'),
             ('jefatura','Jefatura'), ('admin','Administrador')) v(codigo, nombre)
  ON v.codigo = r.codigo
WHERE r.nombre IS DISTINCT FROM v.nombre;

UPDATE roles r SET nombre = v.nombre
FROM (VALUES (...mismas filas...)) v(codigo, nombre)
WHERE v.codigo = r.codigo AND r.nombre IS DISTINCT FROM v.nombre;

-- usuarios.rol: el CHECK fijo pasa a ser una FK hacia el catálogo (roles configurables).
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'usuarios_rol_fkey') THEN
    ALTER TABLE usuarios ADD CONSTRAINT usuarios_rol_fkey
      FOREIGN KEY (rol) REFERENCES roles (codigo) NOT VALID;
    ALTER TABLE usuarios VALIDATE CONSTRAINT usuarios_rol_fkey;
  END IF;
END $$;
ALTER TABLE usuarios DROP CONSTRAINT IF EXISTS usuarios_rol_chk;  -- quita una restricción, no datos
```

Cambios de código asociados (lote 6):
- `semilla_accesos.usecase.go` (ex-`Ensure`) **ya no hace `DELETE FROM permisos`**: inserta la matriz semilla con `ON CONFLICT DO NOTHING`, así los permisos quedan configurables. La matriz del código solo siembra.
- `GET /sesion` agrega `rol_nombre` (leído de `roles.nombre`), un campo aditivo; el frontend lo muestra y cae a su etiqueta local si no llega.
- Compatibilidad de entrada (opcional): tabla `roles_alias(alias PK, codigo FK)` con `ingenieria_coordinacion→coordinacion`, `jefe_seccion→jefatura`, `administrador→admin`, usada solo al importar usuarios desde Excel.
- Brecha que el port **no** cierra: no hay alta, edición ni desactivación de usuarios por API, y v2 RF-01 pide que el jefe de sección administre las cuentas. Queda como historia posterior, ya dentro de la nueva estructura (`usuario.usecase.go`).

---

## 6. Cambios que necesita el frontend

### 6.1 Contrato de API: diferencias y compatibilidad

| Aspecto | Hoy | Con init/backend | Compatibilidad propuesta |
|---|---|---|---|
| Prefijo | `/api/v1/...` | `/areas-verdes/v1/...` | El backend monta **cada group en ambos prefijos** (`Router.Setup` recorre `[]gin.IRouter{servicePath, legacyPath}`); Swagger solo documenta el nuevo. El alias se retira cuando la bitácora pase ≥14 días sin peticiones a `/api/v1` (las PWA ya instaladas pueden seguir con el bundle viejo hasta que el service worker se actualice, `registerType: "autoUpdate"`). |
| Health | `GET /health` → `{"status":"ok","service":"campus-verde-api","database":"up","postgis":"…"}` | `GET /areas-verdes/v1/health` → `{"status":"healthy","timestamp":…}` | `/health` en la raíz **sin cambios**: lo consumen `campus-healthcheck` (`user_data.sh.tftpl`, busca `"status":"ok"`), nginx y `scripts/healthcheck.sh`. El health nuevo sigue el formato de init y agrega `database`/`postgis`. |
| Errores | `{"error": "mensaje"}` con 400/401/403/404/409/422/503 | No definido | Se mantiene idéntico (middleware de errores). Solo se permite **agregar** `codigo` más adelante. El frontend lo lee en `producto.ts:163-168` (`payload.error`). |
| Auth | Cookie HttpOnly `cv_sesion`, `credentials: "include"` | No existe | Sin cambios; la cookie usa `Path=/`, así que sirve para ambos prefijos. Aceptar además `Authorization: Bearer` es una **decisión abierta** (§8, R-09). |
| Paginación | `limit/offset` en `/catastro/ejemplares` y `limit` en `/geo/*`; el resto devuelve listas completas | No definido | Sin cambios en el port. RNF-07 queda para después con `?limit&offset` → `{items,total}`. |
| Nombres de campos | JSON snake_case, contenedores `{items}`, `{capataces}`, `{usuario}`, FeatureCollection | swag toma los tags `json` | Sin cambios: los DTO copian los tags `json` actuales. Lo verifica el arnés de paridad. |
| Contrato | `GET /api/v1/openapi.yaml` (unión manual, 64 paths) | Swagger generado en `/areas-verdes/v1/swagger/doc.json` | `openapi.yaml` queda como alias durante la transición y después redirige a `doc.json`. El frontend no lo consume. |

### 6.2 Cambios en `apps/web`

1. **Centralizar la base:** en `src/api.ts`, `export const API_BASE = import.meta.env.VITE_API_BASE ?? "/areas-verdes/v1"` y `apiUrl(path)`. Reemplazar las **67 apariciones de `/api/v1`** en 17 archivos: `producto.ts` (20), `panel/catastro.ts` (8), `operacion.ts` (7), `panel/inventarioCapas.ts` (6), `types.ts` (4), `panel/inventarioCapas.test.ts` (4), `panel/importaciones.ts` (3), `panel/Poda.tsx` (3), `App.tsx` (2), `panel/Vivero.tsx` (2), `offline/subida.test.ts` (2), `offline/subida.ts` (1), `map/CampusMap.tsx` (1), `panel/EvidenciasCampo.tsx` (1), `panel/Labores.tsx` (1), `panel/calendario.ts` (1) y `panel/calendario.test.ts` (1). La cola offline guarda **cuerpos**, no URLs (`offline/queue.ts`), así que no hay datos persistidos que migrar.
2. **Proxy de desarrollo:** agregar `"/areas-verdes": { target: "http://127.0.0.1:8091" }` en `vite.config.ts` (bloque `server.proxy`, hoy líneas 67-70).
3. **nginx:** agregar `location /areas-verdes/ { proxy_pass http://api:8091; … }` en `nginx.http.conf` y `nginx.tls.conf`, y mantener `location /api/` mientras exista el alias.
4. **Roles:** `types.ts:76-92` (`ROLES`: «Coordinación» → «Ingeniería/Coordinación», y agregar `admin` → «Administrador»); mostrar `sesion.rol_nombre` cuando llegue. `session.ts`, `App.tsx:57-58` y la unión `Rol` no cambian (los códigos se mantienen). El texto de ayuda de `panel/Modulos.tsx:73` sigue siendo válido.
5. **Rutas inexistentes** (`catastro.ts:291, 311, 315`): fuera del port; se resuelven en el lote 24.

---

## 7. Impacto en despliegue, Docker y CI

| Artefacto | Cambio |
|---|---|
| `backend/dockerfile` (nuevo, basado en el de init) | Contexto = **raíz del repo** (necesita `db/migrations` y `data/`), así que los `COPY` van con el prefijo `backend/app/` (desviación documentada del de init, que usa contexto `backend/`). Etapa `golang:1.25.3-alpine` compila `api` (`./cmd`), `migrate`, `etl` y `etl-lote`. Runtime `alpine:3.22` + `ca-certificates tzdata`, usuario `app` con **UID fijo 10001**; copia `db/migrations`, `data/raw`, `data/osm`, `data/mocks` y `docker-entrypoint.sh` (compatible con `sh` de busybox). `ENV SERVER_PORT=8091 SERVER_GIN_MODE=release MIGRATIONS_DIR=… EVIDENCIAS_DIR=/data/evidencias …` y `EXPOSE 8091`, así nginx, compose y user_data no cambian de puerto. |
| Entrypoint | Igual que hoy (migrate → etl una vez → api), con dos cambios: `etl` solo si `/data/.etl-done` no existe **y** `SELECT count(*) FROM areas_verdes` = 0 (subcomando `migrate -necesita-etl`); y mensaje explícito si `/data` no es escribible. |
| Permisos del volumen | Hoy la API corre como root. Con `USER app`, `/opt/campus/data/app` (propiedad de root en la EC2) no sería escribible (`mkdir /data/evidencias`, `.etl-done`, evidencias en disco). Hace falta un `chown -R 10001:10001 /opt/campus/data/app` en `up.sh` / `user_data.sh.tftpl` (una vez) y lo mismo en el compose local (volumen con nombre: se inicializa con el dueño de la imagen). |
| `docker-compose.yml` (raíz) | `api.build.dockerfile: backend/dockerfile`; variables iguales (`DATABASE_URL` lo soporta el config ampliado) + `SERVER_GIN_MODE`. **La BD sigue en `postgis/postgis:16-3.4` y sin montar `db/` en initdb** (desviación del compose de init: con initdb se crearía el núcleo en paralelo, §4.1). `db/schema_nucleo_v0.2.sql` se copia como **referencia** en `db/referencia/`. |
| `infra/terraform/user_data.sh.tftpl` | Mismo `api_image`; se agrega `SERVER_GIN_MODE: release` y `SWAGGER_ENABLED: "false"`; el `chown` anterior. No cambian puertos ni healthcheck. Un cambio de `user_data` no reemplaza la EC2 (`user_data_replace_on_change = false`), así que en la instancia existente hay que aplicarlo a mano por SSM o en `up.sh`. |
| `scripts/deploy-learner-lab.sh` | `-f backend/dockerfile`; además del `latest`, etiquetar la imagen con el **SHA de git** para poder hacer rollback. |
| `Makefile` raíz, `scripts/bootstrap.sh` | `apps/api` → `backend/app` (`go run ./cmd/migrate`, `./cmd/etl`, `./cmd`); se agregan `make swagger` y `make backend-test` que delegan en `backend/Makefile`. |
| `.github/workflows/ci.yml` | Nuevo job `backend`: `actions/setup-go` con `go-version-file: backend/app/go.mod` (1.25.3), `gofmt -l`, `go vet`, `go test ./...`, prohibición de `AutoMigrate` y de `DROP TABLE|DROP COLUMN|TRUNCATE` en `db/migrations/0(4[5-9]|[5-9])*.sql`, y control de deriva de Swagger (`swag init` + `git diff --exit-code app/docs`). El job de `apps/api` (Go 1.22.2) sigue activo hasta el corte y se elimina después; build de la imagen desde `backend/dockerfile`. Deploy sin cambios de lógica. |
| Swagger en producción | Deshabilitado en release (`SWAGGER_ENABLED=false`): la CSP de nginx bloquearía la UI y además es superficie expuesta. En desarrollo, `http://localhost:8091/areas-verdes/v1/swagger/index.html`. |
| API Gateway (implícito en `sanitize-openapi.js`) | No se usa en el despliegue actual (EC2 + nginx). Se mantiene `make swagger` para que el contrato quede listo si el equipo lo adopta. |
| Documentación | `docs/DESPLIEGUE.md`, `docs/DEPLOY-AWS.md`, `docs/OPERACION.md`, `docs/ARQUITECTURA.md`, `README.md` y `apps/api/README.md` (luego `backend/README.md`) se actualizan en el lote de corte; ADR nuevo en `docs/DECISIONES.md`. |

---

## 8. Riesgos y mitigaciones

| # | Riesgo | Prob. / impacto | Mitigación |
|---|---|---|---|
| R-01 | El `etl` del entrypoint trunca el catastro editado si se pierde `/data/.etl-done` (volumen nuevo, chown, otra ruta) y `ejemplares` está vacía (`etl/load.go:47`) | Media / **crítico** | Chequeo en BD antes de correr `etl` (areas_verdes vacía); `negarSiHayDependientes` pasa a revisar también `areas_verdes`, `cambios` y `actividades`; ensayo en copia; backup previo |
| R-02 | Divergencia de respuestas entre la API vieja y la nueva (campos, orden, códigos HTTP, mensajes) | Alta / alto | Arnés de paridad (lote 4) por módulo; los tests actuales se portan antes que el código; DTO con los mismos tags `json` |
| R-03 | Diferencias entre Go 1.22 → 1.25.3, Gin 1.10 → 1.11 y GORM 1.25 → 1.31 (p. ej. `Row().Scan` con NULL, gestión de tiempo, validador) | Media / medio | Tests existentes portados en cada lote; paridad; los tests de `MIGRATE_TEST_URL` corren en CI con un servicio PostGIS |
| R-04 | Confiar en cualquier proxy (default de Gin en init) → `ClientIP` falsificable → límite de login evadible | Alta si no se corrige / alto | `SetTrustedProxies` en el provider de Gin (lote 2) + test |
| R-05 | Usuario no root sin permisos sobre `/data` → la API no arranca o falla al subir evidencias | Alta / alto | UID fijo + `chown` (§7); prueba en compose local antes del corte |
| R-06 | Runtime Alpine: faltan certificados o zona horaria (S3, fechas `America/Lima`) | Baja / medio | `ca-certificates tzdata` ya van en la imagen de init; test de subida a S3 en el ensayo |
| R-07 | Swagger UI roto por la CSP o expuesto en producción | Media / bajo | `SWAGGER_ENABLED=false` en release |
| R-08 | Dos esquemas en paralelo si alguien levanta el compose de init o aplica `schema_nucleo_v0.2.sql` sobre la BD real | Media / alto | `db/schema_nucleo_v0.2.sql` movido a `db/referencia/` y fuera de initdb; ADR que lo prohíbe; vistas `nucleo.*` para quien necesite ese vocabulario |
| R-09 | El equipo espera JWT/Bearer («recibe token», sprint-1) y nosotros usamos cookie | Media / medio | ADR: cookie HttpOnly con token opaco (más seguro para una PWA, sin tokens en JS). Opcional: aceptar `Authorization: Bearer <token opaco>` en el middleware para clientes que no son navegador. **Decisión del equipo.** |
| R-10 | `accesos.Ensure` viejo reescribe `permisos` si se hace rollback | Baja / bajo | Documentado en §4.4; `permisos` está en el dump |
| R-11 | Vistas `nucleo.*` bloquean `ALTER COLUMN … TYPE` futuros sobre las columnas que usan | Media / bajo | Toda migración que cambie el tipo de una columna usada por una vista la recrea (`CREATE OR REPLACE VIEW` después del cambio; `DROP VIEW` solo afecta la vista, no datos) |
| R-12 | Router con muchos groups (≈20) → `routes.go`/`RouterParams` largos y conflictos de merge entre lotes | Alta / bajo | Cada lote agrega solo su campo, su parámetro y su línea en `Setup`, en el orden del §9. Usar `dig` con grupos de valores sería una desviación de init que habría que aprobar antes |
| R-13 | La API ya no arranca sin BD (init hace fatal en `Invoke`), cuando hoy responde `/health` con `database=down` y 503 | Media / bajo | Aceptado: compose usa `depends_on: service_healthy` y `restart: unless-stopped`; el healthcheck de la EC2 alerta igual («no responde»). Documentar en OPERACION |
| R-14 | Nombres de rol distintos entre el núcleo v0.2 («Jefe de Sección», «Ingeniero/Coordinador») y el backlog v2 | Alta / bajo | Prevalece v2; proponer al equipo actualizar la semilla del núcleo |
| R-15 | Convenciones de archivos no definidas en init (§1.5) → estilos distintos entre integrantes | Alta / medio | Ratificar la tabla del §1.5 como ADR antes del lote 5 |
| R-16 | Rutas llamadas por la PWA que no existen (§2.3) se confunden con regresiones del port | Media / bajo | Lista explícita en el arnés como «404 esperado» hasta el lote 24 |
| R-17 | El port se alarga y el equipo sigue desarrollando en `apps/api` | Media / alto | Congelar features nuevas en `apps/api` desde el lote 5 (solo fixes, replicados en `backend/`); orden de lotes por dependencia |

---

## 9. Lotes de implementación (ordenados)

**Reglas comunes a todos los lotes**
- Trabajar en rama `feature/backend-<n>-<tema>`, un PR por lote (Conventional Commits, plantilla de PR del equipo).
- Criterio común: `cd backend && make build && make test` en verde con Go 1.25.3; `gofmt -l backend/app` vacío; `go vet ./...` limpio; `cd apps/api && go test ./...` sigue en verde (no se toca `apps/api` hasta el lote 23).
- Cada lote porta **primero los tests** del módulo desde `apps/api` y después el código.
- Paridad: con las dos APIs sobre la **misma copia de BD** (vieja en `:8091`, nueva en `:8092`), `scripts/paridad-api.sh` (lote 4) no debe reportar diferencias en las rutas del módulo, en ambos prefijos de la nueva.
- Ninguna migración nueva borra, trunca ni elimina columnas.

### Lote 1: Esqueleto fiel de init/backend
- **Alcance:** copiar `backend/` de `init/backend` (`Makefile`, `README.md`, `dockerfile`, `openapi/`, `app/` completo con los `.gitkeep`) a la raíz de este repo, sin cambios salvo el `README` (aviso de «port en curso»). Mismo `go.mod`.
- **Archivos:** `backend/**` (nuevos).
- **Aceptación:** `make build` genera `app/bin/api`; `make run` + `curl localhost:8080/areas-verdes/v1/health` → `{"status":"healthy",…}`; `diff -r` contra la carpeta `backend/` de un clon de `init/backend` = solo el README; `git status` no muestra cambios fuera de `backend/`.

### Lote 2: Config, logger, Gin seguro y middlewares transversales
- **Alcance:** ampliar `shared/config` (§3.1, incluido `DATABASE_URL`); `database.NewConnection` con URL y pool; provider de Gin con `SetTrustedProxies(cfg)`; middlewares `bitacora` (zerolog), `cors`, `cookie`, `limite_login` y `errores`; `domain/errors/base.errors.go`; `Setup` con `NoRoute` → 404 `{"error":"ruta no encontrada"}`; Swagger condicionado por `SWAGGER_ENABLED`.
- **Archivos:** `shared/config/config.go`, `persistence/database/database.go`, `cmd/ioc/container.go`, `presentation/middleware/*.middleware.go`, `presentation/routes/routes.go`, `infrastructure/ratelimit/memoria.go`, `app/.env.example`.
- **Tests portados:** `config/cookie_test.go`, `server/bitacora_test.go`, `accesos/*limite*` y la parte de CORS/cookie de `server/server_test.go`.
- **Aceptación:** tests en verde; `curl -H 'X-Forwarded-For: 1.2.3.4'` no cambia la clave del limitador (test); ruta inexistente → 404 con el cuerpo indicado; con `SWAGGER_ENABLED=false`, `/areas-verdes/v1/swagger/index.html` → 404.

### Lote 3: Runner de migraciones en persistence
- **Alcance:** `persistence/database/migrate.go` (port literal de `Apply`, `splitSQL` y `ayudaMigracion`) y `cmd/migrate/main.go` con el contenedor IoC, leyendo `MIGRATIONS_DIR` (durante la transición apunta a `apps/api/migrations`, **sin copiar archivos**).
- **Tests portados:** `migrate/migrate_test.go`, `migrate/apply_test.go` (con `MIGRATE_TEST_URL`).
- **Aceptación:** sobre una copia de la BD de producción, `go run ./cmd/migrate` imprime «migración ya aplicada» 34 veces y no cambia nada (conteos idénticos a `conteos-antes.tsv`); sobre una BD vacía crea las 43 relaciones y el conjunto de `schema_migrations` coincide con el que genera `apps/api`.

### Lote 4: Sistema, montaje doble y arnés de paridad
- **Alcance:** health ampliado (`database`, `postgis`) en `/areas-verdes/v1/health`; `legado.group.go` con `GET /health` (cuerpo actual exacto), `GET /api/v1` y `GET /api/v1/openapi.yaml` (unión de `apps/api/openapi*.yaml`, portando `UnirContrato`); `Router.Setup` monta los groups en `/areas-verdes/v1` y `/api/v1`; `scripts/paridad-api.sh VIEJA NUEVA rutas.txt` (login por rol con las cuentas semilla, GET, `jq -S`, campos volátiles excluidos, diff).
- **Archivos:** `controller/health.controller.go`, `controller/meta.controller.go`, `groups/legado.group.go`, `groups/meta.group.go`, `routes/routes.go`, `presentation/container.go`, `scripts/paridad-api.sh`, `scripts/paridad-rutas.txt`.
- **Tests portados:** `handlers/contrato_test.go` (64 paths).
- **Aceptación:** `curl :8092/health` ≡ `curl :8091/health`; el arnés sobre `/health`, `/api/v1` y `/api/v1/openapi.yaml` no muestra diferencias.

### Lote 5: Accesos y sesión (paridad estricta)
- **Alcance:** entidades `Usuario`, `Sesion`, `Rol`, `Permiso`; contratos; `sesion.usecase.go`, `usuario.usecase.go`, `semilla_accesos.usecase.go` (**comportamiento idéntico**, incluido el `DELETE FROM permisos`, que se cambia en el lote 6); `bcrypt.hasher.go`; repositorios; middlewares `auth` y `permiso`; `sesion.group.go`, `accesos.group.go`. `cmd/migrate` llama a la semilla.
- **Tests portados:** `accesos/accesos_test.go`, `accesos/cookie_test.go`, casos de sesión de `server/server_test.go`.
- **Aceptación:** login `norte`/`coordinacion`/`jefatura`/`admin` con `CAMPUS_DEV_PASSWORD` → misma cookie `cv_sesion` (atributos iguales); una sesión creada por la API vieja es válida en la nueva y al revés; `GET /sesion` sin cookie → 401 `{"error":…}` igual; paridad de `/sesion` y `/accesos/usuarios`; `SELECT count(*) FROM usuarios` sin cambios.

### Lote 6: Roles v2 (migración 045 y permisos configurables)
- **Alcance:** `apps/api/migrations/045_roles_v2.sql` (§5; se agrega en la carpeta vigente para que **ambos** backends la vean); semilla de permisos por upsert; `rol_nombre` en la sesión; `enums/rol.enum.go`. Enforcement desde BD diferido a post-corte (enforcement en matriz en memoria por paridad con API previa).
- **Aceptación:** ensayo en copia: `SELECT codigo, nombre FROM roles` = tabla del §5; `usuarios` y `sesiones` con conteo idéntico; `permisos` ≥ antes; `cambios` con exactamente 2 filas nuevas (`coordinacion`: «Coordinación» → «Ingeniería/Coordinación»; `admin`: «Administración» → «Administrador»); `capataz` y `jefatura` no cambian; aplicar dos veces no cambia nada; `apps/api` sigue arrancando y logueando con el esquema nuevo (compatibilidad hacia atrás); `GET /areas-verdes/v1/sesion` incluye `rol_nombre`.

### Lote 7: Catálogos
- **Alcance:** módulo completo (§3.2), rutas `/catalogos`.
- **Aceptación:** paridad `GET /catalogos`, `?clase=estado&activos=1`; `POST` con admin → 201 y con capataz → 403, con los mismos mensajes que la API vieja; `POST /:id/desactivar` deja `activo=false` y no borra filas.

### Lote 8: Geo de lectura y fichas de área
- **Alcance:** `/geo/resumen|areas|zonas|capas|capas/:capa|edificios` y `/catastro/areas` (GET, POST, PATCH) con escritura en `cambios`; se migran los 3 structs de `internal/models`.
- **Tests portados:** `handlers/geo_test.go`, `geojson/*_test.go`, `catastro/modelo_test.go` (parte de fichas).
- **Aceptación:** paridad byte a byte (tras `jq -S`) de `/geo/areas?bbox=-77.09,-12.08,-77.07,-12.06&limit=50` y de `/geo/zonas`; `/geo/resumen` informa 521 áreas; un PATCH de ficha genera una fila en `cambios` igual a la de la API vieja.

### Lote 9: Catastro maestro
- **Alcance:** zonas de supervisión, polígonos, cuadrillas, lugares, especies, ejemplares (paginados), códigos históricos y las 6 capas de referencia.
- **Tests portados:** `catastro/recodificar_test.go`, `catastro/sector_test.go`, el resto de `modelo_test.go`.
- **Aceptación:** paridad de las 19 rutas; `?limit=20&offset=40` en ejemplares idéntico; recodificar crea un `codigos_historicos` sin borrar el código anterior.

### Lote 10: Inventario heredado y reservas mock
- **Alcance:** `/geo/inventario`, `/geo/inventario/:capa`, `/geo/inventario/fotos/:name`, `/geo/reservas-mock`, con adaptadores de archivos en `infrastructure/archivos`.
- **Aceptación:** paridad; una foto servida da el mismo `sha256sum` en ambas APIs; path traversal en `:name` → mismo 400/404 que hoy.

### Lote 11: Inventario de campo (frente 2B)
- **Alcance:** tachos, bebederos, puntos, reservas de jardín, fichas de capas, CSV/export y `formato/puntos`.
- **Tests portados:** `capas/patch_test.go`, `capas/reservas_test.go`, `handlers/reservas_test.go`.
- **Aceptación:** paridad de las 23 rutas; un PATCH parcial no toca campos omitidos (test); `DELETE` deja `activo=false` y el conteo total de filas no baja; `tachos.csv` es idéntico.

### Lote 12: Operación / intervenciones
- **Alcance:** `/operacion/capataces`, `/operacion/actividades` (list/create), `asignacion`, `estado`, `archivar`, `timeline`, `ficha`, `avances`; visibilidad por rol (el capataz solo ve lo suyo); idempotencia por UUID.
- **Tests portados:** `operacion/*_test.go`, `handlers/operacion_test.go`.
- **Aceptación:** paridad de la lista (`?abiertas=1`, `?estado=`, `?cuadrilla_id=`) con sesión de coordinación y de capataz; crear la misma labor dos veces → mismo resultado que hoy (idempotente); cada mutación agrega exactamente un `actividad_eventos` con `usuario_id`; archivar exige motivo.

### Lote 13: Solicitudes, órdenes de servicio y riego
- **Alcance:** `/solicitudes` (GET, POST, PATCH), `/ordenes` (GET, POST, PATCH), `/riego` (GET, POST).
- **Tests portados:** `atencion/riego_test.go`, `atencion/riego_cierre_test.go`.
- **Aceptación:** paridad; `codigo_externo` duplicado → mismo 409; el capataz solo ve el riego de su equipo; no se puede cerrar una labor tercerizada sin orden (test `TestCierreTercerizada`).

### Lote 14: Poda y vivero
- **Alcance:** `/podas*` y `/vivero*` (GET, POST, PATCH, archivar).
- **Aceptación:** paridad; archivar deja `archivada_en` y la fila sigue existiendo.

### Lote 15: Evidencias y almacenamiento
- **Alcance:** `/evidencias` (GET, POST multipart), `/evidencias/:id/archivo`, `infrastructure/storage` (S3 y disco), `services/exif.service.go`.
- **Tests portados:** `evidencias/exif_test.go`, `handlers/evidencias_test.go`, `blobs/disk_test.go`.
- **Aceptación:** subir la misma foto dos veces (mismo id) → 200 idempotente, con otro hash → 409; el archivo descargado tiene el mismo SHA-256; `Cache-Control` igual al de hoy; con `EVIDENCIAS_BUCKET` apuntando a un bucket de prueba escribe en S3 (o se documenta que no se probó en el lab).

### Lote 16: Reportes, exportación e IA
- **Alcance:** `/reportes/labores` (JSON, CSV, Excel XML en streaming), indicadores, `/ia/sugerir-tipo`.
- **Tests portados:** `atencion/export_test.go`, `indicadores_test.go`, `ia_test.go`.
- **Aceptación:** `?formato=csv` y `?formato=xls` dan archivos idénticos (diff) en ambas APIs para el mismo filtro; >300 filas sin recorte (`TestReporteNoRecortaEn300`).

### Lote 17: Auditoría y lotes
- **Alcance:** `/auditoria/ediciones`, `/auditoria/cambios`, `/auditoria/timeline`, `/lotes`, `/lotes/:id/revertir`; `services/auditoria.service.go`, que los lotes 8-16 ya usan.
- **Tests portados:** `auditoria/alta_test.go`, `auditoria/reversion_test.go`.
- **Aceptación:** revertir un lote restaura el «antes» y **no pisa** ediciones posteriores (test); paridad del historial por entidad.

### Lote 18: ETL, lectura y normalización (sin BD)
- **Alcance:** mover a `infrastructure/etl/` las partes puras de `internal/etl` (`normalize.go`, `anon.go`, `ficticio.go`, `formato.go`, `lote_leer.go`, `lote_fuente.go`, `baseline.go`, lectura de `sector.go` e `inventario.go`) detrás de `contracts/lector_archivo.contract.go`.
- **Tests portados:** `normalize_test.go`, `lote_test.go`, `sector_test.go`, `inventario_test.go`, `capas2b_test.go`, `importar_test.go`, `atencion_import_test.go`, `load_test.go` (solo las partes sin BD).
- **Aceptación:** `go run ./cmd/etl -skip-load` genera `data/v1/*` **idéntico** (`diff -r`) al del `cmd/etl -skip-load` viejo.

### Lote 19: ETL, escritura, importaciones y CLIs
- **Alcance:** repositorios `carga_*.repository.go` (lote, load, write, capas2b, atencion_import, sector, inventario); casos de uso `importacion`, `carga_inicial`, `carga_lote`, `sectores`; `/importaciones/*`; `cmd/etl`, `cmd/etl-lote`, `cmd/sectores`. `negarSiHayDependientes` ampliado (R-01) y subcomando `migrate -necesita-etl`.
- **Tests portados:** `*_db_test.go` de `etl` (`MIGRATE_TEST_URL`).
- **Aceptación:** en una copia de producción, `etl-lote -solo-lectura` reporta los mismos conteos que el viejo; `etl` (carga inicial) **se niega** si `areas_verdes` tiene filas; previsualizar y confirmar una importación de prueba da el mismo reporte y la misma cantidad de `cambios` que la API vieja; revertirla deja los conteos como antes.

### Lote 20: Swagger y contrato generado
- **Alcance:** anotaciones swag en todos los controllers (`@Router /v1/...`), `make swagger`, `openapi.base.json` saneado; test que recorre `engine.Routes()` y exige que cada ruta de `/areas-verdes/v1` figure en `swagger.json`; `@host` configurable por el entorno o eliminado.
- **Aceptación:** `make swagger` sin errores ni colisiones en `sanitize-openapi.js`; `swagger.json` con ≥ 62 operaciones de negocio; el test de cobertura de rutas en verde.

### Lote 21: Frontend
- **Alcance:** §6.2 puntos 1-4 (`API_BASE`, proxy de Vite, nginx `/areas-verdes/`, etiquetas de rol y `rol_nombre`).
- **Archivos:** los 17 de `apps/web/src` citados, `apps/web/vite.config.ts`, `apps/web/nginx.http.conf`, `apps/web/nginx.tls.conf`.
- **Aceptación:** `npm run lint && npm test && npm run build` en verde; `grep -rn '"/api/v1\|`/api/v1' apps/web/src` vacío fuera de `api.ts`; con la API nueva, la PWA hace login, carga el mapa, crea una labor y sube una evidencia (capturas `fase2-*.png`); con `VITE_API_BASE=/api/v1` sigue funcionando contra la API vieja.

### Lote 22: Imagen, compose, CI y ensayo de corte
- **Alcance:** §7 completo salvo la mudanza de migraciones: `backend/dockerfile`, entrypoint, UID/chown, compose, `deploy-learner-lab.sh` con tag por SHA, `ci.yml` (job backend), `Makefile` raíz, `bootstrap.sh`. Ensayo completo del §4.4 pasos 1-4 sobre un dump real.
- **TODO (revisión Opus lote 22):** empaquetar en el Dockerfile: binario migrate, carpeta migrations, contrato openapi.yaml y variable de entorno OPENAPI_PATH.
- **Aceptación:** `docker compose up -d --build` local levanta db + api nueva + web; `/health` ok; `scripts/counts.sh` igual antes y después; CI en verde en ambos jobs; informe del ensayo (conteos antes/después, versiones de `schema_migrations`, salida del arnés de paridad) adjunto al PR.

### Lote 23: Corte en producción y mudanza de migraciones
- **Alcance:** `git mv apps/api/migrations/*.sql db/migrations/` (**mismos nombres**), `MIGRATIONS_DIR=/opt/campus/migrations` apuntando al nuevo origen en el dockerfile, `db/schema_nucleo_v0.2.sql` a `db/referencia/`, ADR en `docs/DECISIONES.md`, actualización de `DESPLIEGUE.md`, `DEPLOY-AWS.md`, `OPERACION.md`, `ARQUITECTURA.md` y `README.md`. Ejecutar el §4.4 completo en la EC2.
- **Aceptación:** checklist del §4.4 firmado: backup y snapshot con fecha; conteos antes = después en todas las tablas existentes; `/health` ok desde `campus-healthcheck`; login, mapa, labor y evidencia verificados en producción; tag de rollback anotado. `apps/api` queda **solo como referencia** (código, no datos) y su retiro se hace en un PR aparte, ≥14 días después y sin tráfico en `/api/v1`.

### Lote 24: Alineación con el núcleo v0.2 y huecos del frontend
- **Alcance:** migraciones `046_evidencia_evento.sql`, `047_sync_offline.sql`, `048_estados_nucleo.sql` y `049_vistas_nucleo.sql` (§4.3) con sus campos opcionales en DTO y entidades; decisión e implementación (o eliminación en la PWA) de las 3 rutas inexistentes (§2.3) como bajas lógicas con historial en `cambios`.
- **Aceptación:** cada migración aplicada dos veces sin cambios; conteos de las tablas existentes idénticos; `SELECT count(*) FROM nucleo.intervencion` = `SELECT count(*) FROM actividades`; `apps/api` en su último tag sigue funcionando con el esquema (compatibilidad hacia atrás); las rutas de baja responden 200 y la fila sigue existiendo con `activo=false`.

---

### Anexo: puntos que no se pudieron determinar

- Conteos reales de la BD desplegada, y si `ejemplares`/`asignaciones_poligono` tienen filas en producción (de eso depende la severidad de R-01). No hay acceso a la EC2 ni Docker en este entorno.
- Si existe hoy `/opt/campus/data/app/.etl-done` en la instancia.
- La versión exacta de PostGIS en producción (el compose declara `postgis/postgis:16-3.4`).
- Si el equipo ratifica las convenciones de archivo del §1.5, el uso de cookie frente a Bearer (R-09) y la adopción de API Gateway que sugiere `sanitize-openapi.js`.
