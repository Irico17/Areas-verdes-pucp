# Campus Verde — Gestión de áreas verdes Campus PUCP (Pando)

Monorepo del **DP2 · Grupo 12**: sistema de gestión (catastro, operación, reportes) con el **módulo de mapa** sobre la misma API. No es un mapa aislado ni un segundo backend.

Fuentes oficiales en [`docs/fuente/`](docs/fuente/). Plan de mapa: [`docs/PLAN-INTEGRACION-MAPA.md`](docs/PLAN-INTEGRACION-MAPA.md).

## Stack

| Capa | Tecnología | Estado |
|------|------------|--------|
| API | **Go + Gin + GORM** — monolito modular | Fase A |
| Datos | **PostgreSQL + PostGIS** en la misma base (sin GeoServer) | Fase B |
| Catastro | ETL `data/raw` → `data/v1` → PostGIS, EPSG:4326 | Fase B |
| PWA | React + TypeScript + Vite + MapLibre | Fase C |
| Objetos | Bucket privado tipo S3 | Posterior |

Roles en el visor: **Jefatura**, **Coordinación**, **Capataz**. Es un interruptor local, no SSO.

## Qué hay cargado

| Capa | Origen | Registros | Tabla |
|------|--------|-----------|--------|
| Áreas verdes | `data/raw/areas_verdes.geojson` | **521** | `areas_verdes` |
| Zonas | `data/raw/jefe_de_grupo.json` | **534** | `zonas` |
| Jardines de reserva | `data/raw/jardines_reserva.geojson` | 21 | `capas_auxiliares` |
| Xerofítica | `data/raw/xerofitica.geojson` | 10 | `capas_auxiliares` |

Las zonas **no** guardan el campo `jefes` (nombres de personas). El identificador de zona es `Z-NNNN`. Las áreas usan `AV-NNNN`. Geometría: MultiPolygon **EPSG:4326** (lon/lat), la que consume MapLibre. `geom` puede ser NULL en registros futuros sin GPS (catastro progresivo).

## Cómo correrlo en local

Requisitos: Docker (Compose v2), Go 1.22+, Node 22+ para el visor.

```bash
cp .env.example .env
docker compose up -d
make wait
make migrate
make etl
make api
```

Equivalente: `make bootstrap` y después `make api`.

La API escucha en **http://127.0.0.1:8091**. El visor, en otra terminal:

```bash
cd apps/web && npm install
make web
```

Abre **http://127.0.0.1:4317**. Vite reenvía `/api` y `/health` a la API.

```bash
curl -s http://127.0.0.1:8091/health
curl -s http://127.0.0.1:8091/api/v1/geo/resumen
curl -s -H 'Accept: application/geo+json' http://127.0.0.1:8091/api/v1/geo/areas | head -c 400
curl -s http://127.0.0.1:8091/api/v1/geo/zonas?limit=1
```

Conteos en la base:

```bash
make counts
```

Pruebas del módulo Go (sin base de datos):

```bash
cd apps/api && go test ./...
```

`make etl` **reemplaza** el catastro semilla (TRUNCATE + carga). Es idempotente respecto a `data/raw`. No es un editor de geometrías.

La contraseña de `.env.example` es solo para desarrollo local. `.env` no se versiona.

## Rutas

| Método | Ruta | Respuesta |
|--------|------|-----------|
| GET | `/health` | Proceso, Postgres y PostGIS |
| GET | `/api/v1` | Índice |
| GET | `/api/v1/openapi.yaml` | Contrato |
| GET | `/api/v1/geo/resumen` | Conteos |
| GET | `/api/v1/geo/areas` | FeatureCollection de áreas |
| GET | `/api/v1/geo/zonas` | FeatureCollection de zonas |
| GET | `/api/v1/geo/capas` | Capas auxiliares |
| GET | `/api/v1/geo/capas/{capa}` | `jardines_reserva` o `xerofitica` |

Filtro opcional `bbox=minLon,minLat,maxLon,maxLat` (EPSG:4326) y `limit`.

Detalle del contrato: [`apps/api/openapi.yaml`](apps/api/openapi.yaml) y [`apps/api/README.md`](apps/api/README.md).

## Estructura

```
apps/api          API Go (cmd/api, cmd/migrate, cmd/etl)
apps/web          Visor PWA (React, MapLibre, OSM)
data/raw          recovery, no editar
data/v1           GeoJSON normalizado (salida del ETL)
data/mocks        reservas FAKE
docs/             arquitectura, ADRs, plan de mapa
scripts/          espera de Postgres, bootstrap, conteos
docker-compose.yml
```

## Documentación

| Doc | Contenido |
|-----|-----------|
| [`docs/ARQUITECTURA.md`](docs/ARQUITECTURA.md) | Arquitectura DP2 |
| [`docs/PLAN-INTEGRACION-MAPA.md`](docs/PLAN-INTEGRACION-MAPA.md) | Fases A–F (A, B y C hechas) |
| [`docs/PLAN-MIGRACION.md`](docs/PLAN-MIGRACION.md) | Migración legacy → DP2 |
| [`docs/DECISIONES.md`](docs/DECISIONES.md) | ADRs, incluido CRS y PII de zonas |
| [`docs/legacy-recovery/`](docs/legacy-recovery/) | Análisis del monolito Leaflet |

## Fuera de este corte

SSO institucional, Fase D (actividades en el mapa), Fase E (3D), Fase F (tachos, flora, bebederos…), Google Sheets en runtime y reservas reales (solo el mock FAKE).
