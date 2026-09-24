# Plan de migración / integración — Campus Verde (DP2 Grupo 12)

Este documento orienta el paso **legacy recovery → sistema DP2**. El detalle del **mapa** está en [`PLAN-INTEGRACION-MAPA.md`](PLAN-INTEGRACION-MAPA.md) (fases A–F). La arquitectura objetivo está en [`ARQUITECTURA.md`](ARQUITECTURA.md).

## Objetivos

1. Alinear el monorepo al stack confirmado: **PWA React+TS** + **Go/Gin/GORM** + **PostgreSQL/PostGIS** + bucket privado.
2. Sembrar **Catastro** desde GeoJSON/JSON recuperados (`data/raw/`), sin Sheets en runtime.
3. Entregar **base operativa** y **atención** según orden del cliente; el **módulo mapa CORE** se integra sobre la misma API (no un backend paralelo).
4. 3D por extrusión y capas legacy de inventario = posteriores al MVP CORE.

## Estado del baseline (2026-09-24, America/Lima)

| Fuente | Estado |
|--------|--------|
| Arquitectura + backlog Grupo 12 | En `docs/fuente/` |
| GeoJSON / sheets CSV / fotos / legacy-app | `data/raw/` |
| Sheet reservas | 401 → mock FAKE `data/mocks/` |
| `apps/api` | Go + Gin + GORM. `GET /health` y `/api/v1/geo/...` |
| `apps/web` | Placeholder PWA React (Fase C) |
| PostGIS / Docker | Compose `db` (PostGIS 16). Catastro cargado: 521 áreas, 534 zonas |

## Fases macro (producto DP2)

| Orden cliente | Contenido | Relación mapa |
|---------------|-----------|---------------|
| 1. Base operativa | Accesos, catálogos, catastro progresivo, labores, riego; offline + UUID | Fases A–C del plan mapa |
| 2. Atención completa | Solicitudes, propia/tercerizada, cierre, reportes, trazabilidad | Fase D (CORE) cuando exista API de actividades |
| 3. Capacidades priorizadas | Import/export, indicadores, IA validada; SSO/mapas según acuerdos | Fases E–F; ortofoto/SSO opcionales |

## Fases técnicas del repo

Ver checklist detallado en **`PLAN-INTEGRACION-MAPA.md`**:

- **A** Alinear a Go API  
- **B** Seed Catastro PostGIS  
- **C** Visor OSM + capas  
- **D** Supervisión actividades (CORE MUST)  
- **E** Extrusión 3D  
- **F** Overlays inventario legacy  

## Mapeo dominio recovery → DP2

| Dominio recovery | Módulo DP2 |
|------------------|------------|
| áreas / zonas / sectores | Catastro |
| personal / jefes de grupo | Accesos + Personal (operarios sin cuenta) |
| labores futuras / tareas | Operación |
| tachos, flora, bebederos, fauna | Inventario (overlays; Fase F) |
| reservas jardines | Catastro + agenda (mock hasta fuente) |
| monitoreo / vivero CSV | Operación / catálogos según ERS |

## Criterios de salida globales

- API Go responde health + al menos un endpoint geo de catastro.
- PWA muestra mapa OSM con capas seed.
- Actividad creada desde mapa aparece como marcador y respeta rol capataz.
- Cero lecturas a Google Sheets en runtime.

## Riesgos

| Riesgo | Mitigación |
|--------|------------|
| Sketch Node previo en docs | Reescrito; ADR de adopción Go |
| Agenda reservas irrecuperable | Mock FAKE + import manual |
| Scope creep capas | CORE primero; Fase F después |
| Sin glTF | Extrusión/OSM; no bloquear |

## Referencias

- `docs/fuente/` — documentos oficiales Grupo 12  
- `docs/legacy-recovery/` — análisis del monolito Leaflet  
- `docs/DECISIONES.md` — ADRs  
