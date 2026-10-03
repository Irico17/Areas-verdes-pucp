# apps/api

API anterior de VerdePUCP (módulo `campusverde/api`, Go 1.22). Se conserva para `go test ./...` y para `scripts/paridad-api.sh`.

La API que se despliega, y la que publica el CI como `campus-verde-api`, es [backend/app](../../backend/README.md). El esquema ya no vive en esta carpeta: los SQL están en `db/migrations` ([migrations/README.md](migrations/README.md)).

```bash
go test ./...
go run ./cmd/api
go run ./cmd/migrate
go run ./cmd/etl
```

Las pruebas que usan base leen `MIGRATE_TEST_URL`. Si no está, prueban `postgres://campus:campus@127.0.0.1:5432/postgres?sslmode=disable` y omiten el caso cuando no hay Postgres. Crean bases temporales y las borran al terminar.

El contrato de esta API sigue en `openapi.yaml`.
