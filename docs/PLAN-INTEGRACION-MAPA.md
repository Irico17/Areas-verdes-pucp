# Plan de integración del módulo mapa — alineado a DP2 Grupo 12

Respeta el **orden del cliente**: (1) base operativa → (2) atención completa → (3) capacidades priorizadas; mapas dependen de acuerdos, pero el backlog ya define la épica CORE (MUST).

Fuentes: `docs/fuente/Propuesta_Arquitectura_DP2_v0.2.docx`, backlog épica J (HUID 29–35), recovery en `data/raw/`.

## Qué es MVP mapa vs después

| | MVP / CORE (MUST) | Después / no bloqueante |
|---|-------------------|-------------------------|
| Stack | Misma API Go; PostGIS en el mismo Postgres; PWA | GeoServer, tiles propias, glTF campus |
| Capas | Zonas/sectores + áreas catastro seed | Tachos, flora, bebederos, fauna, playas… |
| Operación | Marcadores de actividades, crear/asignar, panel, timeline, reasignar, baja lógica | Extrusión 3D, ortofoto cliente, Street View (explícitamente no) |
| Capataz | Solo asignadas; offline → lista + sync | Edición geométrica avanzada en campo |
| Base cartográfica | OpenStreetMap | Ortofoto OSG si hay acuerdo |

## Fases

### Fase A — Alinear repo a API Go *(hecha, 2026-09-24)*

**Objetivo:** dejar de tratar `apps/api` como Node.

- [x] Docs y README: Go + Gin + GORM.
- [x] Esqueleto Go (`apps/api/cmd/api`, `internal/config`, `internal/db`, `internal/handlers`, `internal/models`) con Gin + GORM contra Postgres.
- [x] Contratos `/api/v1` en `apps/api/openapi.yaml` y en `GET /api/v1`.

**MVP:** sí (fundación). **Mapa:** aún no.

### Fase B — Seed Catastro desde GeoJSON recuperado *(hecha, 2026-09-24)*

**Objetivo:** Postgres+PostGIS con catastro usable como capas.

- [x] Extensión PostGIS + tablas de áreas y zonas (MultiPolygon EPSG:4326, `geom` nullable = geometría progresiva). Capas opcionales en `capas_auxiliares`.
- [x] ETL `data/raw` → `data/v1` → load (`make etl`):
  - `areas_verdes.geojson` → **521** features (`areas_verdes`, ids `AV-NNNN`)
  - zonas desde `jefe_de_grupo.json` → **534** features (`zonas`, ids `Z-NNNN`). Se omite el campo `jefes` (nombres de personas).
  - `jardines_reserva.geojson` → **21**; `xerofitica.geojson` → **10** (no bloquean el MUST)
- [x] Endpoints lectura GeoJSON FeatureCollection:
  - `GET /api/v1/geo/areas`
  - `GET /api/v1/geo/zonas`
  - auxiliares: `GET /api/v1/geo/capas` y `GET /api/v1/geo/capas/{capa}`

**CRS:** EPSG:4326 (lon/lat), el que espera MapLibre / RFC 7946.  
**MVP catastro (RF-04/06):** sí. **Marcadores de actividad:** no aún.

### Fase C — Visor mapa en PWA (OSM + capas catastro) *(hecha, 2026-09-24)*

**Objetivo:** mapa 2D de consulta sobre catastro, sin depender aún del CRUD completo de actividades.

- [x] Scaffold PWA React+TS en `apps/web`.
- [x] Base OSM; capas on/off (zonas, sectores, áreas, jardines y xerofítica).
- [x] Auth stub + roles para quién ve el visor (Jefatura / Coordinación / Capataz, sin SSO).

**MVP visor:** sí como base de CORE. **Crear actividad desde mapa:** Fase D.

### Fase D — Supervisión de actividades en mapa (épica CORE) *(hecha, 2026-09-24)*

**Depende de:** API de actividades/asignaciones/trazabilidad (base operativa + atención).

- [x] Marcadores de actividades no cerradas (color=estado, letra=tipo); panel filtros; popup.
- [x] Timeline de eventos (RF-30).
- [x] Crear/asignar desde mapa: pin + selector capataz (RF-31).
- [x] Vista capataz restringida mapa/lista + cola local IndexedDB con UUID idempotente (RF-32, RNF-01).
- [x] Reasignación con auditoría (RF-35); cancelar/archivar lógica (RF-34).

**MVP CORE mapa:** **sí — esto es el MUST del backlog.**  
Implementar **después** (o en paralelo controlado) de tener modelo de actividades; no antes de Fase A–B estables.

### Fase E — 3D extrusión *(hecha, 2026-09-24)*

- [x] `fill-extrusion` / alturas sobre `areas_verdes` (heurística por `area_m2`). La vista plana sigue con `fill`.
- [x] Footprints OSM buildings del campus (`data/osm/edificios_pando.geojson`, 483 huellas, © OpenStreetMap).

**MVP CORE:** **no.** Mejora UX post-piloto 2D.

### Fase F — Paridad legacy (overlays inventario)

- [ ] Tachos, flora/cafetos, bebederos (+ fotos), fauna, playas, puertas, etc. como capas opcionales de inventario.
- [ ] Sustituir cualquier resto de Sheets; mantener mock reservas hasta fuente real.

**MVP CORE:** **no.** No bloquean supervisión por mapa.

## Mapeo recovery → módulos DP2

| Dato recovery | Módulo destino | Fase |
|---------------|----------------|------|
| `areas_verdes`, xerofítica, vereda | Catastro | B / C |
| `jefe_de_grupo` (zonas) | Catastro (+ referencia personal) | B / C |
| `jardines_reserva` + mock agenda | Catastro / operación reservas si aplica | B; agenda cuando exista fuente |
| Actividades (nuevas en API) | Operación → marcadores mapa | D |
| Tachos, flora, bebederos, fauna… | Inventario overlays | F |
| Legacy `script.js` Leaflet | Solo referencia; no runtime | — |

## Criterios de salida

| Fase | Listo cuando… |
|------|----------------|
| A | Repo y docs hablan solo de Go API; esqueleto Gin/GORM y OpenAPI publicados. **Hecho.** |
| B | PostGIS: 521 áreas, 534 zonas, CRS 4326; capas opcionales 21 + 10. **Hecho.** |
| C | Demo: OSM + 2–3 capas catastro desde API. **Hecho.** |
| D | Demo: crear actividad con pin, verla como marcador, capataz solo ve las suyas. **Hecho.** |
| E | Extrusión visible sin romper 2D |
| F | Capas legacy must-have del inventario en checklist opcional |

## Riesgos

| Riesgo | Mitigación |
|--------|------------|
| Arquitectura decía “mapas pendientes” vs backlog MUST | Tratar mapa CORE como **módulo priorizado** una vez exista Operación; no segundo sistema |
| Scope 54 capas día 1 | Fases B–D primero; F después |
| PII en jefes de grupo | ETL con campos mínimos; no API pública |
| Ortofoto / capas OSG sin acuerdo | OSM siempre; ortofoto opcional |
| Dual backend “geo” | **Prohibido** — ADR en `DECISIONES.md` |
