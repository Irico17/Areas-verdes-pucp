# VerdePUCP — Gestión de Áreas Verdes

VerdePUCP es el sistema de gestión de áreas verdes del campus PUCP (Pando): catastro, labores, solicitudes, riego y reportes, con el mapa sobre la misma API. No es un mapa aislado ni un segundo backend.

Licencia: [MIT](LICENSE).

Fuentes oficiales en [`docs/fuente/`](docs/fuente/). Plan de mapa: [`docs/PLAN-INTEGRACION-MAPA.md`](docs/PLAN-INTEGRACION-MAPA.md). Plan de producto e interfaz: [`docs/PLAN-PRODUCTO-Y-UI.md`](docs/PLAN-PRODUCTO-Y-UI.md).

## Stack

| Capa | Tecnología | Estado |
|------|------------|--------|
| API | **Go + Gin + GORM** — monolito modular | Fase A |
| Datos | **PostgreSQL + PostGIS** en la misma base (sin GeoServer) | Fase B |
| Catastro | ETL `data/raw` → `data/v1` → PostGIS, EPSG:4326 | Fase B |
| PWA | React + TypeScript + Vite + MapLibre | Fase C |
| Objetos | Archivos de evidencia en disco local (`data/evidencias`, no se versiona) | Piloto. Sin bucket S3 |
| Sesión | Cuentas locales y cookie HttpOnly | No es SSO |

Roles de la sesión: **Capataz** (norte, sur, riego), **Coordinación**, **Jefatura** y **Admin**.

Las cuentas semilla (`norte`, `sur`, `riego`, `coordinacion`, `jefatura`, `admin`) y la clave `pando-local` (`CAMPUS_DEV_PASSWORD`) son **solo para desarrollo local**. No son cuentas de la universidad ni el SSO de la PUCP. No las use en un entorno con datos reales.

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

Abre **http://127.0.0.1:4317**. Vite reenvía `/api` y `/health` a la API. Entra con `coordinacion` / `pando-local` (u otra cuenta semilla: `norte`, `sur`, `riego`, `jefatura`, `admin`).

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
| GET | `/api/v1/operacion/capataces` | Equipos ficticios (Equipo Norte, Sur, Riego) |
| GET | `/api/v1/operacion/actividades` | Labores abiertas, GeoJSON Point. `rol=capataz&capataz_id=` filtra |
| POST | `/api/v1/operacion/actividades` | Alta desde un pin. UUID de cliente, idempotente |
| PATCH | `/api/v1/operacion/actividades/{id}/asignacion` | Asignar o reasignar |
| PATCH | `/api/v1/operacion/actividades/{id}/estado` | Cambio de estado |
| POST | `/api/v1/operacion/actividades/{id}/archivar` | Baja lógica |
| GET | `/api/v1/operacion/actividades/{id}/timeline` | Bitácora |
| POST | `/api/v1/sesion` | Ingreso local. Cookie `cv_sesion` HttpOnly |
| GET | `/api/v1/catalogos` | Catálogos (tipos, estados, lugares, especies…) |
| GET/POST/PATCH | `/api/v1/catastro/areas` | Fichas y alta sin geometría |
| GET/POST | `/api/v1/solicitudes` | Solicitudes con código externo |
| GET/POST | `/api/v1/ordenes` | Órdenes de una labor tercerizada |
| GET/POST | `/api/v1/riego` | Riego por sector y turno |
| POST | `/api/v1/evidencias` | Archivo ligado a una labor |
| GET | `/api/v1/reportes/labores` | Reporte básico y conteos. Filtros `zona`, `cuadrilla`, `origen`, `desde`, `hasta`. `formato=csv` o `formato=xls`. Cobertura, rendimiento y métricas de proveedor: definición pendiente. Sin PDF |
| POST | `/api/v1/ia/sugerir-tipo` | Regla local, en proceso. La persona confirma el tipo. No sale del proceso |

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
| [`docs/PLAN-INTEGRACION-MAPA.md`](docs/PLAN-INTEGRACION-MAPA.md) | Fases A–F del visor, hechas |
| [`docs/PLAN-PRODUCTO-Y-UI.md`](docs/PLAN-PRODUCTO-Y-UI.md) | Backlog, interfaz y olas de producto |
| [`docs/PLAN-MIGRACION.md`](docs/PLAN-MIGRACION.md) | Migración legacy → DP2 |
| [`docs/DECISIONES.md`](docs/DECISIONES.md) | ADRs, incluido CRS y PII de zonas |
| [`docs/legacy-recovery/`](docs/legacy-recovery/) | Análisis del monolito Leaflet |

El visor crea labores con un pin (Jefatura y Coordinación), las asigna a un equipo y muestra la bitácora. El rol Capataz solo recibe las de su equipo. Si la API no responde, el alta queda en IndexedDB y se reintenta con el mismo UUID.

La vista **Relieve** extruye las áreas verdes según su superficie y, si se enciende la capa, las huellas de edificios OSM del recinto (`GET /api/v1/geo/edificios`). **Plano** vuelve al relleno 2D.

Inventario opcional (apagado al entrar): bebederos 67, fauna 19, puertas 7, tachos 184, flora 74, cafetos 53, playas 15, vereda 1. `GET /api/v1/geo/inventario` y `GET /api/v1/geo/inventario/{capa}`. La agenda de reservas es ficticia: `GET /api/v1/geo/reservas-mock` (71 ítems, sin hoja de cálculo). No hay JPEG de bebederos en `data/raw`; el mapa lo indica en el popup.

## Despliegue

El destino previsto es un AWS Academy Learner Lab: una EC2 pequeña con Docker Compose (PostGIS, API y nginx). Los pasos, el presupuesto y cómo destruir la instancia están en [`docs/DEPLOY-AWS.md`](docs/DEPLOY-AWS.md). Desde una laptop, con credenciales temporales del lab:

```bash
bash scripts/deploy-learner-lab.sh
```

Esas credenciales no van en el repositorio. `.env`, `*.tfstate`, `*.tfvars` y `.terraform/` están en `.gitignore`.

## Fuera de este corte

SSO institucional y reservas reales. La agenda que se ve es el mock FAKE.
