# Ensayo de corte — lote 22

Fecha: 2026-10-02. Rama `backend/arquitectura-equipo`. Este ensayo cubre los pasos 1–4 del §4.4 del plan **a nivel de base de datos**, sin corte en producción (el lote 23 no se ejecuta).

## Qué se pudo comprobar aquí

| Comprobación | Resultado |
|---|---|
| `gofmt -l` en `backend/app` | vacío |
| `go vet ./...` y `go test ./...` en `backend/app` | en verde. Sin `MIGRATE_TEST_URL`: los tests de BD se saltan |
| `go test ./...` en `apps/api` | en verde (también sin BD de prueba) |
| `make build` (`backend/Makefile`) | genera `backend/app/bin/api` |
| Binarios del Dockerfile con `CGO_ENABLED=0` | `./cmd`, `./cmd/migrate`, `./cmd/etl`, `./cmd/etl-lote`, `./cmd/sectores` compilan |
| `sh -n` | `backend/docker-entrypoint.sh`, `scripts/bootstrap.sh`, `scripts/deploy-learner-lab.sh` |
| Tests del entrypoint | `internal/infrastructure/deploy`: BD con datos (exit 10, no corre ETL), BD vacía (corre ETL), error de `necesita-etl`, y `/data` no escribible (aviso de `chown` UID 10001). El caso de permisos se salta si el proceso es root |
| YAML | `docker-compose.yml` y `.github/workflows/ci.yml` parsean con PyYAML. El compose principal no monta `docker-entrypoint-initdb.d` ni un `.sql` en el servicio `db` |
| `AutoMigrate(` | no aparece en `*.go` de `backend/` ni `apps/api/` |
| Migraciones `045`–`048` | no contienen `DROP TABLE`, `DROP COLUMN` ni `TRUNCATE` de datos; `046` solo suelta un CHECK de roles |
| `swag init -g cmd/main.go -o docs/` (swag v1.16.6) | `git diff` de `backend/app/docs` vacío |
| Web (lote 21) | `npm run lint`, `npm test` (75) y `npm run build` en verde. Un segundo build con `VITE_API_BASE=/api/v1` deja la base `/api/v1` en el bundle |

## Pasos 1–4 del §4.4

Ejecutados después del lote 22 (y revalidados con lotes 23–24) sobre copias desechables `vp_*` de `campus_verde`, sin tocar la base de desarrollo ni la EC2. Ver `/workspace/informe-lotes-23-24.md` y el ensayo previo del lote 22.

- Copias `vp_c_*` / `vp_rev_*` creadas con `pg_dump` en modo lectura y borradas al terminar.
- `schema_migrations` y conteos por tabla antes/después: idénticos salvo las versiones nuevas (`047`/`048` en el tramo del lote 24).
- Paridad GET y escrituras con `scripts/paridad-api.sh` / `paridad-escrituras-*` en verde.
- Humo del proxy Vite → API nueva: login y rutas bajo `/areas-verdes/v1`.

No se escribió en `campus_verde` ni en la EC2.

## Qué no se probó por falta de Docker

- `docker build -f backend/dockerfile`
- `docker compose up`
- el job `backend` de GitHub Actions (gofmt, vet, tests con el servicio PostGIS, y el build de la imagen)

El entrypoint, el UID 10001 y el `chown` de `user_data.sh.tftpl` / `up.sh` quedan revisados en el script y en el test de entrypoint, no en un contenedor.

## Qué haría el ensayo cuando haya una copia de `campus_verde`

1. `pg_dump -Fc` de `campus_verde` y restore en `vp_c_22` (y otra copia para la API anterior).
2. Conteos de cada tabla de `public` (el SQL del §4.4) y `SELECT version FROM schema_migrations ORDER BY 1`.
3. `MIGRATIONS_DIR` apuntando a `db/migrations` (fuente de verdad desde el lote 23; `apps/api/migrations` ya no tiene SQL), `go run ./cmd/migrate` del backend nuevo. Tiene que imprimir migración ya aplicada (o aplicar solo `047`/`048` si faltan) y no cambiar conteos de datos.
4. `go run ./cmd/migrate -necesita-etl` debe salir con código 10 (la base tiene filas). El entrypoint, en ese caso, no llama a `etl`.
5. API anterior y API nueva sobre copias distintas, `scripts/paridad-api.sh`, y los mismos conteos al final.
6. `DROP DATABASE` de las copias `vp_c_*`.
