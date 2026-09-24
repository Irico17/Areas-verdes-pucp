# packages/schemas

Contratos auxiliares entre PWA y API.

Con el stack DP2 (**Go + Gin + GORM**), la fuente de verdad del API en este corte es [`apps/api/openapi.yaml`](../../apps/api/openapi.yaml). Este paquete puede alojar:

- Schemas TypeScript (p. ej. Zod) **espejo** para validación en `apps/web`, o
- JSON Schema / tipos generados desde OpenAPI.

Dominios previstos alineados a módulos DP2: accesos/catálogos, catastro/inventario, operación/atención, servicios tercerizados, seguimiento/reportes, y shapes GeoJSON de capas/marcadores del mapa.
