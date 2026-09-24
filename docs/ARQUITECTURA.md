# Arquitectura — Sistema de gestión de áreas verdes Campus PUCP (Pando)

Alineado a la **Propuesta de arquitectura DP2 v0.2 · Grupo 12** (`docs/fuente/Propuesta_Arquitectura_DP2_v0.2.docx`) y al product backlog (`docs/fuente/Product_Backlog_Areas_Verdes.xlsm`).

> El mapa **no** es un producto aparte ni un segundo backend: es un **módulo del frontend (PWA)** que consume los mismos endpoints geo/operativos de la API Go.

## 1. Producto y alcance

| Ítem | Decisión |
|------|----------|
| Producto | Sistema de gestión de áreas verdes — Campus PUCP (Pando) |
| Frontend | PWA **React + TypeScript** (propuesta de equipo; campo + oficina) |
| Backend | **Go + Gin + GORM** — monolito modular (**decisión confirmada**) |
| DB | **PostgreSQL** con extensión **PostGIS** (no hay servidor GIS / GeoServer aparte) |
| Objetos | Bucket privado tipo S3 (metadatos/claves en Postgres; bytes en el almacén) |
| Offline | IndexedDB + sync; escrituras con **UUID** (idempotencia); network-first en lecturas |
| Ámbito | Campus Pando parametrizable; **sin** multicampus en MVP |
| Operarios | Participan en labores **sin cuenta** obligatoria en MVP |

Orden de implementación del cliente (respetar):

1. **Base operativa** — accesos, catálogos, catastro progresivo, labores y riego; offline/evidencias/idempotencia desde el inicio.
2. **Atención completa** — solicitudes (códigos externos), propia/tercerizada, estados/cierre, reportes y trazabilidad.
3. **Capacidades priorizadas** — import/export, indicadores, IA validada; **mapas y SSO dependen de acuerdos**.

## 2. Vista de contenedores

```
┌──────────────────────────────────────────────────────────────────┐
│  apps/web — PWA React + TypeScript                               │
│  ┌─────────────┐ ┌──────────┐ ┌───────────┐ ┌──────────────────┐ │
│  │ Accesos /   │ │ Catastro │ │ Operación │ │ Seguimiento /    │ │
│  │ catálogos   │ │ inventario│ │ atención  │ │ reportes         │ │
│  └──────┬──────┘ └────┬─────┘ └─────┬─────┘ └────────┬─────────┘ │
│         │             │   ┌─────────┴────────┐       │           │
│         │             │   │ Módulo MAPA      │       │           │
│         │             │   │ (OSM + capas +   │       │           │
│         │             │   │  marcadores act.)│       │           │
│         │             │   └─────────┬────────┘       │           │
│  IndexedDB (offline: catálogos, labores, fotos, pendientes)      │
└─────────┼─────────────┼─────────────┼────────────────┼───────────┘
          └─────────────┴────── HTTPS / JSON (+ GeoJSON) ──────────┘
                                │
                     ┌──────────▼──────────┐
                     │  apps/api           │
                     │  Go · Gin · GORM    │
                     │  monolito modular   │
                     │  auth · módulos ·   │
                     │  sync idempotente   │
                     └──────────┬──────────┘
              ┌─────────────────┼─────────────────┐
              ▼                                   ▼
   ┌─────────────────────┐            ┌─────────────────────┐
   │ PostgreSQL + PostGIS│            │ Object storage      │
   │ (system of record)  │            │ (bucket privado)    │
   └─────────────────────┘            └─────────────────────┘
```

**Estilo:** un solo proceso de API; módulos = paquetes Go (handlers Gin → servicios → GORM). Sin microservicios, Redis, bus de eventos ni BI separado en el piloto.

## 3. Módulos de negocio (backend)

| Módulo | Responsabilidad | Notas |
|--------|-----------------|-------|
| **Accesos / catálogos** | Auth, roles, catálogos configurables (clase/tipo actividad, especies, lugares/sectores, estados, prioridades…) | RF-24; desactivar en vez de borrar |
| **Catastro / inventario** | Ubicaciones, zonas, sectores, áreas, ejemplares; georreferencia **progresiva** (RF-04, RF-06) | Zona/lugar válidos **sin** GPS al registrar |
| **Operación / atención** | Solicitudes, intervenciones/actividades, avances, riego por sector/turno/ciclo, evidencias | Cierre bloqueado sin ejecución completa |
| **Servicios tercerizados** | Órdenes, proveedor, conformidad; atención “En proceso” durante contratación | Acceso directo de proveedores = evolución |
| **Seguimiento / reportes** | Panel, trazabilidad, reportes básico/intermedio/avanzado, indicadores | Consultas a Postgres; sin plataforma BI |
| **Mapa (cliente + geo API)** | Capas catastro + marcadores de actividades; no es dominio de datos propio | Misma API; ver §5 |

Relación clave: una **actividad** (Operación) tiene ubicación (pin o área/zona de Catastro) y capataz asignado; el mapa solo **visualiza y crea** vía esos endpoints.

## 4. Roles (MVP)

| Rol | Cuenta | Capacidades típicas |
|-----|--------|---------------------|
| **Capataz** | Sí | Ver/ejecutar solo actividades asignadas (mapa o lista PWA); avances, evidencias; offline campo |
| **Ingeniería / Coordinación** | Sí | Supervisión, asignación, catastro, solicitudes, mapa CORE |
| **Jefatura** | Sí | Reportes, indicadores, visión de cobertura/cumplimiento |
| **Operario** | No (MVP) | Figurante en personal/avances sin login obligatorio |
| Admin técnico | Sí | Catálogos, parametrización, secretos fuera del código |

Auth prevista: sesiones con cookies Secure/HttpOnly/SameSite (+ CSRF); cuentas propias en MVP; **SSO pendiente** de TI.

## 5. Cómo se enchufa el módulo mapa (una sola API)

El backlog épica **J — Módulo CORE — Gestión, supervisión y trazabilidad por mapa** (HUID 29–35 / RF-29–35, RNF-11) exige:

- Supervisión de actividades activas en mapa (marcadores por estado/tipo, panel filtros, popup).
- Timeline / trazabilidad de actividad.
- Crear/asignar actividades desde el mapa (pin + selector de capataz).
- Capataz: solo asignadas (mapa/lista); offline → degradar a lista.
- Zonas/sectores como capas.
- Reasignar capataz con auditoría; cancelar/archivar (baja lógica).
- Base **OpenStreetMap**; ortofoto del cliente opcional.

### 5.1 Diseño técnico

| Pieza | Dónde | Qué hace |
|-------|-------|----------|
| Visor mapa | `apps/web` (módulo UI) | OSM (+ ortofoto opcional); capas; marcadores; formularios de creación |
| Capas catastro | API Go `/api/v1/geo/...` o equivalentes de Catastro | Polígonos/zonas/sectores desde **PostGIS** |
| Marcadores actividad | API Go de **Operación** (GeoJSON o coords en DTO) | Features derivados de actividades abiertas — **no** de un segundo servicio |
| PostGIS | Extensión en el **mismo** Postgres | Geometría + índices GIST; **no** GeoServer / WMS obligatorio |
| Offline mapa | IndexedDB del PWA | Vistas de campo; sin conexión, lista; sync con UUID |

### 5.2 Inventario recuperado vs actividades

| Origen | Rol en el sistema |
|--------|-------------------|
| `data/raw/areas_verdes.geojson`, zonas jefes (`jefe_de_grupo.json`), sectores/jardines, etc. | **Semilla ETL → Catastro** (capas de fondo / inventario geográfico) |
| Actividades, estados, capataz, timeline | **Operación** — marcadores vivos del mapa CORE |
| Tachos, flora, bebederos, fauna… | Overlays de inventario (**paridad legacy**, no bloquean MVP CORE) |

Sheets/Drive **fuera** de runtime; recovery = ETL/seed.

### 5.3 3D (posterior al CORE 2D)

Extrusión (`fill-extrusion` / alturas) sobre `areas_verdes` + footprints OSM — **Fase E** del plan de mapa. No bloquea RF-29–33.

## 6. Datos y migración

```
data/raw/     →  scripts/etl_*  →  data/v1/  →  migraciones Go/SQL + PostGIS
docs/fuente/  →  contrato de producto (arquitectura + backlog)
```

- Reservas agenda: mock FAKE en `data/mocks/` hasta fuente real (legado 401).
- Fotos: bucket privado; claves en DB.
- PII en `jefe_de_grupo.json`: no exponer en API pública sin minimización.

Detalle de fases mapa/API: `docs/PLAN-INTEGRACION-MAPA.md`. Plan legacy amplio: `docs/PLAN-MIGRACION.md` y `docs/legacy-recovery/`.

## 7. Contratos

- API versionada `/api/v1/...`; GeoJSON (`application/geo+json`) donde aplique.
- `packages/schemas` puede evolucionar a OpenAPI generado desde Go o schemas TS espejo; **no** asumir Zod como contrato del backend Go.
- Sync: UUID de escritura; rechazo de UUID repetido con payload distinto; optimistic concurrency al corregir.

## 8. Despliegue (objetivo, no implementado aún)

- Compose local previsto: `web`, `api` (Go), `db` (Postgres+PostGIS).
- Cloud AWS (RNF backlog); detalles con TI.
- Secretos solo por env. Postgres sin exposición pública.

## 9. Estado de implementación

Fases A y B del plan de mapa están en el repo: API Go (`apps/api`), Compose de PostgreSQL+PostGIS y ETL de catastro (`make bootstrap`). Siguen fuera de este corte:

- PWA / visor mapa (Fase C), actividades en mapa (Fase D), extrusión 3D (Fase E), overlays legacy (Fase F).
- SSO institucional, acceso de proveedores, caso de IA cerrado.
- Servidor GIS dedicado (sigue descartado: PostGIS en la misma base).
