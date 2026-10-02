# Decisiones de arquitectura (ADR-lite)

Registro de decisiones del scaffold `campus-verde`. Zona horaria de referencia: America/Lima.

---

## ADR-001 — Repo nuevo, no reescribir el monolito in-place

**Contexto:** Existe recovery Leaflet (`mapa-web-6`) con datos GeoJSON/Sheets. El alcance real es el sistema DP2 de gestión de áreas verdes, no solo un mapa.

**Decisión:** Monorepo `/workspace/campus-verde`; recovery como `data/raw/` + `docs/legacy-recovery/`.

**Consecuencias:** Paridad por ETL + API + PWA nueva, no parchando `script.js`.

---

## ADR-002 — Apps separadas web + api (monolito modular)

**Contexto:** PWA de campo/oficina, auth, catastro, operación, reportes y mapa.

**Decisión:** `apps/web` (PWA) + `apps/api` (backend) + `packages/` auxiliar. Un solo system of record. Despliegue conjunto (compose) o separado.

**Consecuencias:** Contrato HTTP único; el mapa es módulo cliente, no silo.

---

## ADR-003 — PostgreSQL + PostGIS como system of record

**Contexto:** Sheets/Drive dejan de ser fiables (reservas 401).

**Decisión:** Entidades operativas y geometrías en PostgreSQL. PostGIS como **extensión del mismo** Postgres (ver ADR-010). Raw/v1 = etapas ETL.

**Consecuencias:** ETL obligatorio; consultas espaciales e integridad.

---

## ADR-004 — Reservas: mock FAKE tras 401 persistente

**Contexto:** Sheet de agenda de reservas responde HTTP 401 de forma persistente.

**Decisión:** `data/mocks/reservas_agenda.mock.json` etiquetado **FAKE**; no bloquear el baseline.

**Consecuencias:** Sustituir cuando exista CSV/Excel o Sheet re-publicado.

---

## ADR-005 — 3D por extrusión primero; glTF después

**Contexto:** Sin modelo glTF de campus. Backlog CORE prioriza mapa 2D operativo.

**Decisión:** Tras visor 2D + CORE, `fill-extrusion` + OSM buildings. glTF solo si hay asset confiable.

**Consecuencias:** 3D no bloquea RF-29–33.

---

## ADR-006 — Git local only

**Contexto:** Remoto GitHub puede no estar disponible.

**Decisión:** `git init` + commits locales. Sin `remote` obligatorio.

**Consecuencias:** Push manual cuando haya acceso.

---

## ADR-007 — Scaffold de apps diferido

**Contexto:** Prioridad: alinear docs al stack DP2 y conservar datos recovery.

**Decisión:** Placeholders README en `apps/web` y `apps/api`; sin Vite completo ni binario Go en el commit de alineación.

**Consecuencias:** Siguiente iteración = esqueleto Go + PWA.

---

## ADR-008 — Adoptar stack DP2 Go + Gin + GORM (reemplaza placeholder Node)

**Fecha:** 2026-09-24 (America/Lima)

**Contexto:** El scaffold inicial documentaba Node (Fastify/Nest) en `apps/api`. La **Propuesta de arquitectura DP2 v0.2 · Grupo 12** confirma backend **Go + Gin + GORM** (monolito modular), PWA React+TS, PostgreSQL y objetos privados.

**Decisión:** Toda documentación y el placeholder de `apps/api` se alinean a **Go + Gin + GORM**. Se abandona el sketch Node como objetivo. React+TypeScript se mantiene como propuesta de PWA.

**Consecuencias:** `packages/schemas` Zod deja de asumirse como contrato del backend; preferir OpenAPI desde Go o schemas TS espejo. Planes de migración reescritos (fases A–F). No se implementa aún el binario Go en este ADR.

---

## ADR-009 — El mapa es módulo frontend + endpoints geo en la misma API Go

**Fecha:** 2026-09-24

**Contexto:** La arquitectura v0.2 dejaba mapas como integración pendiente; el backlog incluye épica MUST **Módulo CORE — Gestión, supervisión y trazabilidad por mapa** (HUID 29–35). Riesgo de crear un “mini GIS” aparte.

**Decisión:** Un solo backend Go. Capas de catastro y marcadores de actividades se exponen como rutas de los módulos **Catastro** y **Operación**. El visor vive en `apps/web`. **Prohibido** un segundo servicio/backend solo para mapas.

**Consecuencias:** Fase D del plan mapa depende de la API de actividades; inventario recovery alimenta capas, no un producto mapa independiente.

---

## ADR-010 — PostGIS en el mismo PostgreSQL; sin GeoServer

**Fecha:** 2026-09-24

**Contexto:** Necesidad de geometrías (RF-04, RF-06, RNF-11) sin presuponer servidor GIS. La propuesta DP2 indica no asumir servidor GIS dedicado.

**Decisión:** Habilitar extensión **PostGIS** en la instancia PostgreSQL de la aplicación. Lectura/escritura vía GORM/SQL y respuestas GeoJSON desde Gin. No GeoServer, no WMS/WFS obligatorio en MVP.

**Consecuencias:** Índices GIST y migraciones versionadas; tiles vectoriales propias quedan como evolución opcional.

---

## ADR-011 — Datos de recovery = seed/ETL; Sheets fuera de runtime

**Fecha:** 2026-09-24

**Contexto:** GeoJSON/CSV/fotos recuperados de `mapa-web-6` / Sheets públicos; reservas ya inaccesibles.

**Decisión:** `data/raw` → ETL → `data/v1` → PostGIS. Google Sheets/Drive **no** se consultan en producción. Capas recovery de áreas/zonas/sectores = semilla de **Catastro**; tachos/flora/bebederos = overlays posteriores (Fase F). Mock FAKE solo donde falte fuente.

**Consecuencias:** Congelar el legacy como solo lectura; paridad operativa por importación controlada, no por Apps Script.

---

## ADR-012 — Catastro Fase B: EPSG:4326, MultiPolygon y zonas sin PII

**Fecha:** 2026-09-24

**Contexto:** El seed recuperado trae áreas (521 MultiPolygon) y zonas de `jefe_de_grupo.json` (534 Polygon/MultiPolygon, ya en lon/lat CRS84). El campo `jefes` contiene nombres de personas. El visor futuro es MapLibre.

**Decisión:**

- Almacenar y servir **EPSG:4326**. No hace falta reproyectar para MapLibre.
- Normalizar Polygon → MultiPolygon al cargar. `geom` queda **nullable** para el catastro progresivo (RF-04/RF-06).
- Identificadores estables `AV-NNNN` y `Z-NNNN` (índice del archivo + 1), no el serial de la base.
- **No** persistir ni exponer `jefes`. La zona se identifica por `feature_id`, geometría y atributos de lugar (nombre del jardín, uso, riego, código, referencia).
- Jardines de reserva (21) y xerofítica (10) van a `capas_auxiliares`. No se mezclan con el conteo de áreas.
- El ETL vive en el mismo módulo Go (`cmd/etl`). Recarga el catastro semilla (TRUNCATE + insert) de forma idempotente.

**Consecuencias:** `data/v1/zonas.geojson` y `GET /api/v1/geo/zonas` no contienen nombres de jefes. Re-ejecutar `make etl` sustituye esas tablas; no es un cargador incremental de edición humana.

---

## ADR-013 — El esquema vive en `db/migrations`, no en un SQL de initdb

**Fecha:** 2026-10-02

**Contexto:** La base con datos cargados se construyó con migraciones numeradas (`001`–`046`, más `047` y `048`). El equipo describe otro esquema (`schema_v1.1.sql`, antes `schema_nucleo_v0.2.sql`) y un compose que lo monta en `docker-entrypoint-initdb.d`. Ese archivo de núcleo no está en este repositorio. Aplicarlo crearía tablas paralelas junto a las que ya tienen datos.

**Decisión:**

- La única definición del esquema es `db/migrations/*.sql`, con los mismos nombres de archivo que ya registró `schema_migrations.version`.
- La imagen copia esa carpeta a `/opt/campus/migrations` y `MIGRATIONS_DIR` apunta ahí. El entrypoint sale si el directorio no existe.
- `apps/api` queda como referencia de código. Su `migrate` y sus pruebas leen `db/migrations`.
- `db/referencia/` no se monta en Postgres. No se aplica `schema_nucleo_v0.2.sql` ni `schema_v1.1.sql`.
- No hay `AutoMigrate`. Las migraciones nuevas son aditivas e idempotentes. No se hace `DROP`/`TRUNCATE`/`DELETE FROM` de datos cargados.
- El corte en la EC2 (backup, snapshot, conteos, tag de rollback) lo ejecuta una persona con `docs/RUNBOOK-CORTE-PRODUCCION.md`. Este ADR no lo autoriza por sí solo.

**Consecuencias:** Un volumen nuevo de Postgres queda vacío hasta que corre `migrate`. No se crea el núcleo del equipo al lado de nuestras tablas.

---

## ADR-014 — Mapa base OSM; la ortofoto espera el archivo del cliente

**Fecha:** 2026-10-02

**Contexto:** RF-29 y RNF-11 piden una vista aérea del campus y dejan el proveedor del mapa base por definir. La observación del backlog pide dejar escrita la decisión. En `data/raw` no hay ortofoto ni servicio de imágenes aéreas del cliente. La aplicación ya dibuja el campus con MapLibre.

**Decisión:**

- El mapa base es OpenStreetMap, servido como teselas raster en MapLibre. No se agrega Street View.
- No se simula una ortofoto ni se publica una imagen aérea inventada.
- La ortofoto queda pendiente del archivo que entregue el cliente. Cuando exista, se incorpora como capa, sin reemplazar esta decisión por una URL ficticia.
- Si las teselas no cargan o no hay red, la vista de actividades pasa a lista y la pantalla dice por qué. No hay caché de teselas en esta entrega.

**Consecuencias:** El plano cumple el desplazamiento, el zoom y la atribución de OpenStreetMap. La vista aérea del texto de RF-29-CA1 sigue abierta hasta que llegue el archivo.
