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
| Migraciones `045` y `046` | no contienen `DROP TABLE`, `DROP COLUMN` ni `TRUNCATE` |
| `swag init -g cmd/main.go -o docs/` (swag v1.16.6) | `git diff` de `backend/app/docs` vacío |
| Web (lote 21) | `npm run lint`, `npm test` (75) y `npm run build` en verde. Un segundo build con `VITE_API_BASE=/api/v1` deja la base `/api/v1` en el bundle |

## Pasos 1–4 del §4.4

No se pudieron ejecutar contra datos.

En esta máquina no hay cliente `psql`, ni servidor Postgres/PostGIS, ni Docker, ni un dump de `campus_verde`, ni el script `scripts/copia-bd.sh` apuntando a una instancia viva. Por eso no hay:

- copia `vp_c_22` (ni una segunda copia para la API anterior);
- `schema_migrations` de antes y de después;
- conteos por tabla antes y después;
- salida de `scripts/paridad-api.sh`;
- humo del proxy de Vite (`:5199` → API nueva en `:8092`): login, `GET /areas-verdes/v1/sesion`, `/geo/areas`, `/operacion/actividades`.

No se creó ni se borró ninguna base. No se escribió en `campus_verde`.

## Qué no se probó por falta de Docker

- `docker build -f backend/dockerfile`
- `docker compose up`
- el job `backend` de GitHub Actions (gofmt, vet, tests con el servicio PostGIS, y el build de la imagen)

El entrypoint, el UID 10001 y el `chown` de `user_data.sh.tftpl` / `up.sh` quedan revisados en el script y en el test de entrypoint, no en un contenedor.

## Qué haría el ensayo cuando haya una copia de `campus_verde`

1. `pg_dump -Fc` de `campus_verde` y restore en `vp_c_22` (y otra copia para la API anterior).
2. Conteos de cada tabla de `public` (el SQL del §4.4) y `SELECT version FROM schema_migrations ORDER BY 1`.
3. `MIGRATIONS_DIR` apuntando a `apps/api/migrations`, `go run ./cmd/migrate` del backend nuevo. Tiene que imprimir migración ya aplicada y no cambiar conteos.
4. `go run ./cmd/migrate -necesita-etl` debe salir con código 10 (la base tiene filas). El entrypoint, en ese caso, no llama a `etl`.
5. API anterior y API nueva sobre copias distintas, `scripts/paridad-api.sh`, y los mismos conteos al final.
6. `DROP DATABASE` de las copias `vp_c_*`.
