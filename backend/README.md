# Backend

API de VerdePUCP. Módulo Go `github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend`, Go 1.25.3, Gin, GORM y `dig`. El proceso que se despliega es este. `apps/api` es la API anterior y no es la imagen del CI.

Prefijos HTTP: `/areas-verdes/v1` y el alias `/api/v1`. El mismo router monta los dos.

## Capas

```
backend/app/cmd/                  proceso, migrate, etl, etl-lote, sectores, modelgen
backend/app/internal/presentation controladores, rutas, middleware, peticiones
backend/app/internal/application  casos de uso, servicios, contratos y DTO
backend/app/internal/domain       entidades, enums y errores
backend/app/internal/persistence  GORM, modelos, mappers y repositorios
backend/app/internal/infrastructure  ETL, archivos, almacenamiento, rate limit
backend/app/internal/shared       configuración y logs
backend/dockerfile                imagen de la API
backend/docker-entrypoint.sh      migrate, carga y exec de la API
```

El contenedor de `dig` está en `cmd/ioc`. Una ruta nueva se registra en el controlador, en el grupo de `presentation/routes/groups` y en el contenedor de presentación. Swagger sale de las anotaciones `swag` y se regenera con `make swagger` desde este directorio. No se editan a mano `app/docs/docs.go`, `swagger.json` ni `swagger.yaml`.

OpenAPI que sirve la API: `/areas-verdes/v1/openapi.yaml` y `/api/v1/openapi.yaml`. En `APP_ENV=produccion` responden 404. En develop, QA y en local sin `APP_ENV` responden 200, con independencia de `SWAGGER_ENABLED`. La UI de Swagger, cuando está encendida, queda en `/areas-verdes/v1/swagger/index.html`.

El contrato histórico de la API anterior sigue en `apps/api/openapi.yaml`. `backend/openapi/openapi.base.json` es la base que produce `make swagger`.

## Migraciones

Son aditivas e idempotentes. Cada archivo de `db/migrations` se aplica una vez, en su transacción, y se anota en `schema_migrations`. No hay migraciones «down». No se usa `AutoMigrate`. No se renombran ni se reordenan los archivos.

```bash
cd app
go run ./cmd/migrate
go run ./cmd/migrate -semilla-ficticia
go run ./cmd/migrate -necesita-etl
go run ./cmd/migrate -catastro-incompleto
```

Hace falta `DATABASE_URL` (o el conjunto `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_NAME`) y `CAMPUS_DEV_PASSWORD`. `MIGRATIONS_DIR` por defecto es `db/migrations` en la raíz del repositorio y `/opt/campus/migrations` dentro de la imagen.

`Ensure` crea las seis cuentas ficticias si no existen y no reescribe un hash ya guardado. `-semilla-ficticia` aplica `deploy/seed/ficticio.sql`.

El detalle de la carga y de los flags está en [docs/DATOS-Y-ETL.md](../docs/DATOS-Y-ETL.md).

## etl y etl-lote

```bash
cd app
go run ./cmd/etl
go run ./cmd/etl --skip-load
go run ./cmd/etl --no-strict
go run ./cmd/etl-lote
go run ./cmd/etl-lote -solo-lectura
go run ./cmd/sectores
```

Desde la raíz: `make etl`, `make etl-lote` y `make sectores`.

## Arranque local

Con Postgres ya arriba (desde la raíz, `docker compose up -d db` y `make wait`):

```bash
cp ../.env.example ../.env
cd app
go run ./cmd
```

Por defecto escucha en `:8091` si `API_ADDR` está definido, y si no en el puerto de `SERVER_PORT` (8080). El compose y el `.env.example` de la raíz fijan `API_ADDR=:8091`.

```bash
curl -s http://127.0.0.1:8091/health
```

Dentro de la imagen el entrypoint exige `CAMPUS_DEV_PASSWORD`, migra, decide el ETL y solo entonces ejecuta la API. En `APP_ENV=produccion` rechaza `pando-local`, `campus-lab` y claves de menos de 16 caracteres antes de migrar.

## Pruebas

```bash
cd app
gofmt -l .
go vet ./...
MIGRATE_TEST_URL='postgres://campus:campus@127.0.0.1:5432/campus?sslmode=disable' go test ./...
```

Sin `MIGRATE_TEST_URL`, las pruebas que necesitan base se omiten. Esas pruebas crean una base `vp_c_test_…` y la borran al terminar. No apuntan a una base con datos de servicio.

`make gen-models` (desde `backend/`) corre `cmd/modelgen` contra una base desechable cuyo nombre empieza por `vp_` o `modelgen_`, ya migrada y sin filas en `areas_verdes`. Aborta si la URL apunta a `campus_verde`. La salida va a un temporal o a `MODELGEN_OUT`.

`make swagger` exige `swag` v1.16.6, `swagger2openapi` y Node. El job `backend` del CI compara el resultado con lo versionado.

## Variables de entorno

Los valores de ejemplo están en `/.env.example`. Aquí solo el nombre y para qué sirve. Ninguna clave real.

| Variable | Uso |
| --- | --- |
| `APP_ENV` | `develop`, `qa`, `produccion`, o vacío (pool y Swagger históricos). |
| `API_ADDR` | Dirección de escucha, por ejemplo `:8091`. Si falta, se usa `SERVER_PORT`. |
| `SERVER_PORT` | Puerto si no hay `API_ADDR`. Default del código: `8080`. |
| `SERVER_GIN_MODE` | `debug` o `release`. Vacío: debug en develop y release en el resto. |
| `SERVER_TRUSTED_PROXIES` | Lista separada por comas. Vacío: ninguna. |
| `LOG_LEVEL` | `debug`, `info`, `warn` o `error`. |
| `LOG_FORMAT` | `console` o `json`. |
| `DATABASE_URL` | URL de Postgres. Si está, manda sobre las piezas sueltas. |
| `DATABASE_HOST` | Host. Default del código: `localhost`. |
| `DATABASE_PORT` | Puerto. Default: `5432`. |
| `DATABASE_USER` | Usuario. Default del código: `areasverdes`. El compose local usa `campus`. |
| `DATABASE_PASSWORD` | Clave de Postgres cuando no va dentro de `DATABASE_URL`. |
| `DATABASE_NAME` | Nombre de la base. |
| `DATABASE_SCHEMA` | Esquema. Default: `public`. |
| `DATABASE_SSL_MODE` | `disable`, `require`, o un modo de libpq. `true` equivale a `require`. |
| `DATABASE_MAX_OPEN_CONNS` | Tope de conexiones abiertas. |
| `DATABASE_MAX_IDLE_CONNS` | Conexiones inactivas. |
| `DATABASE_CONN_MAX_LIFETIME` | Vida máxima, en formato de `time.ParseDuration` (`30m`). |
| `CAMPUS_DEV_PASSWORD` | Clave inicial de las cuentas ficticias. Obligatoria para migrar y para el entrypoint. |
| `CAMPUS_CORS_ORIGINS` | Orígenes separados por coma. Un `*` se ignora. |
| `CAMPUS_COOKIE_SECURE` | `true` solo con HTTPS. Cualquier otro valor, o vacío, deja la cookie sin `Secure`. |
| `CAMPUS_COOKIE_SAMESITE` | Default `Lax`. |
| `CAMPUS_LOGIN_MAX` | Intentos de login por minuto. Default 8. |
| `CAMPUS_ENV` | Lo escribe el compose. La cookie `Secure` no depende de esta variable. |
| `SWAGGER_ENABLED` | `true` o `false`. Vacío: encendido en develop y QA, apagado en producción, y según Gin si `APP_ENV` está vacío. |
| `SWAGGER_HOST` | Host que anuncia Swagger. |
| `MIGRATIONS_DIR` | Carpeta de SQL. |
| `SEED_FILE` | Semilla ficticia. Default: `deploy/seed/ficticio.sql`. |
| `SEED_PROFILE` | `etl` o `ficticio`. Lo lee el entrypoint, no el binario de la API. |
| `EVIDENCIAS_DIR` | Carpeta local de evidencias. |
| `EVIDENCIAS_BUCKET` | Cubo privado. Vacío: disco. |
| `DATA_RAW_DIR` | Fuentes del ETL. |
| `DATA_V1_DIR` | Salida normalizada. |
| `EDIFICIOS_PATH` | GeoJSON de edificios OSM. |
| `RESERVAS_MOCK_PATH` | Agenda ficticia. |
| `DRIVE_FOTOS_DIR` | Fotos recuperadas. No se versionan los JPEG. |
| `OPENAPI_PATH` | YAML que sirve la API. Default: `apps/api/openapi.yaml`. |
| `MODELGEN_OUT` | Directorio de salida de `modelgen`. |
| `AWS_REGION` | Región del SDK cuando hay cubo. No es una clave. |
| `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN` | Cadena por defecto del SDK. No van en el repositorio. Con un rol de instancia no hacen falta. |

Postgres del compose, leídas por Docker y no por el proceso Go: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_PORT`.
