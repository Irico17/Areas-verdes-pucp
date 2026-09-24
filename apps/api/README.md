# apps/api

API del sistema DP2 — **Go + Gin + GORM** (monolito modular). Módulo Go: `campusverde/api`.

El mapa no tiene backend propio: consume estas rutas.

## Layout

```
cmd/api          proceso HTTP
cmd/migrate      aplica apps/api/migrations/*.sql
cmd/etl          data/raw → data/v1 → PostGIS
internal/config  DATABASE_URL, API_ADDR, rutas del repo
internal/db      conexión GORM / Postgres
internal/models  áreas, zonas, capas auxiliares
internal/handlers
internal/catastro lectura GeoJSON
internal/etl     normalización (sin el campo jefes) y carga
internal/migrate
migrations/      001 PostGIS, 002 catastro
openapi.yaml
```

El esquema espacial lo definen las migraciones SQL. GORM abre la conexión y lee las tablas; no se usa AutoMigrate sobre geometría.

## Variables

Ver `/.env.example` en la raíz del repo. Por defecto:

- `DATABASE_URL=postgres://campus:campus@127.0.0.1:5432/campus_verde?sslmode=disable`
- `API_ADDR=:8091`

## Comandos

Desde la raíz del repo (con Postgres ya arriba):

```bash
make migrate
make etl
make api
```

O, dentro de este directorio:

```bash
go run ./cmd/migrate
go run ./cmd/etl
go run ./cmd/api
go test ./...
```

`go run ./cmd/etl --skip-load` solo escribe `data/v1`. `--no-strict` no exige los conteos del baseline (521 / 534 / 21 / 10).

## CRS

EPSG:4326. `ST_AsGeoJSON` sirve lon/lat. Un Polygon de origen se guarda como MultiPolygon. `geom` NULL está permitido para registros sin GPS.
