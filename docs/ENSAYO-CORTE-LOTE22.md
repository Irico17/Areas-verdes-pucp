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

## Pasos 1–4 del §4.4 (Ensayo ejecutado contra datos reales)

Ejecutados y revalidados tras los lotes 22–24 sobre copias desechables `vp_*` de `campus_verde` y bases vacías, sin tocar la base de desarrollo ni la EC2 (ver `/workspace/informe-lotes-23-24.md` §2 y §4).

Las migraciones las aplica `backend/docker-entrypoint.sh` mediante el binario `migrate` (el binario `api` no migra).

### Resultados de migraciones 001–048 (`db/migrations`, 38 archivos)
- **Copia de `campus_verde` (36 versiones):**
  - 1.ª corrida: aplica `047_evidencia_evento.sql` y `048_uuid_cliente_eventos.sql` («migraciones al día»).
  - 2.ª corrida: 0 migraciones aplicadas («migraciones al día»).
  - **Conteos por tabla antes vs después:** idénticos en las 38 tablas de usuario; única diferencia `schema_migrations` 36 → 38.
  - **Columnas nuevas:** `evidencias.evento_id bigint NULL`, `actividad_eventos.uuid_cliente uuid NULL`; índices `evidencias_evento_id_idx` y `actividad_eventos_uuid_cliente_uidx` presentes.
- **BD vacía:** 1.ª corrida aplica 38 migraciones; 2.ª corrida aplica 0. Total: 48 relaciones. Estructura idéntica (`information_schema`) a la copia migrada.
- Migraciones `047` y `048` aditivas, sin `DROP`, `TRUNCATE` ni `DELETE`.

### Resultados de ejecución del entrypoint (`backend/docker-entrypoint.sh`)
Se ejecutó el script real del entrypoint con los binarios `migrate` y `etl` compilados (`API_BIN=/bin/true`):
- **BD vacía (1.ª corrida):** aplica 38 migraciones, detecta «BD vacía: requiere carga inicial de ETL», ejecuta ETL (`areas=521 zonas=534`), crea `.etl-done`; BD con 521 áreas y 38 versiones.
- **BD vacía (2.ª corrida):** migraciones al día, no repite ETL.
- **Sin `.etl-done` pero BD ya con datos:** detecta «BD con datos … no se ejecuta la carga inicial»; 521 áreas intactas.
- **Copia con datos (36 versiones), sin `.etl-done`:** aplica solo `047` y `048`; no corre ETL; 521 áreas antes y después.
- **Copia (2.ª corrida):** 0 migraciones aplicadas.
- **Sin `CAMPUS_DEV_PASSWORD`:** rc=1 con mensaje claro de error.
- **`MIGRATIONS_DIR` inexistente:** rc=1 (`MIGRATIONS_DIR=/nope no existe`).
- **Migración rota (049 de prueba con tabla inexistente):** rc=1, «FALLO DE MIGRACION: la transaccion se revirtio»; la versión no se registra en `schema_migrations`; la base de datos queda intacta.

## Paridad de API
- **Lecturas GET (237 rutas en `scripts/paridad-rutas.txt`):** 100 % de paridad (0 diferencias) en ambos prefijos (`/api/v1` y `/areas-verdes/v1`).
- **Escrituras (10 archivos en ambos prefijos):** 100 % de paridad y conteos de tablas y `cambios` idénticos.
- **Bajas lógicas (`scripts/paridad-escrituras-bajas.txt`):** verificado con arnés `solo-nueva` (401, 403, 404, 400 mismatch, PATCH 200, bajas 200 con `activo: false`, idempotencia y exclusión en lecturas).

No se escribió en `campus_verde` ni en la EC2.

## Qué no se probó por falta de Docker en este entorno

- `docker build -f backend/dockerfile`
- `docker compose up`
- el job `backend` de GitHub Actions como pipeline completo (se ejecutaron todos sus pasos por separado).

El entrypoint, el UID 10001 y el `chown` de `user_data.sh.tftpl` / `up.sh` quedan revisados en el script y en los tests de entrypoint (`internal/infrastructure/deploy`).
