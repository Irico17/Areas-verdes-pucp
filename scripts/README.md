# scripts

| Script | Qué hace |
|--------|----------|
| `set-aws-secrets.sh` / `set-aws-secrets.ps1` | Sube las tres credenciales temporales del Learner Lab a los secrets del repo. Ver `docs/DESPLIEGUE.md` |
| `bootstrap.sh` | `docker compose up -d`, espera, migra y corre el ETL |
| `wait-db.sh` | Espera a que `pg_isready` responda en el servicio `db` |
| `counts.sh` | Conteos de áreas, zonas y capas, con SRID 4326 |
| `paridad-api.sh` | Arnés de verificación de paridad entre la API vieja (`apps/api`) y la nueva (`backend/app`). Normaliza rutas, cabeceras y respuestas JSON. Ignora la clave `activo` en fichas de áreas verdes (expuesta por el backend nuevo en F2 pero ausente en la API vieja) sin ocultar otras diferencias. Soporta la directiva `solo-nueva` para pasos exclusivos del backend nuevo. |

El ETL en sí está en `apps/api/cmd/etl` (Go). Desde la raíz, `make bootstrap`, `make etl` y `make counts` llaman a estos scripts o al módulo.

## Arnés de Paridad (`paridad-api.sh`)

- **Comparación GET y escrituras**: Compara respuestas de la API vieja contra la API nueva (status HTTP, Set-Cookie/cabeceras normalizadas y cuerpos JSON ordenados con `jq -S`).
- **Normalización de `activo` en fichas de áreas**: Las fichas de áreas verdes del backend nuevo incluyen `activo: bool` tras la implementación de la baja lógica (F2). Como `apps/api` no expone esa clave, el arnés la excluye al comparar fichas (`.areas[]` y objetos con `feature_id` y metadatos de área) para no generar falsos positivos, preservando todas las demás claves.
- **Directiva `solo-nueva`**: Permite ejecutar pasos que solo aplican al backend nuevo (rutas nuevas como bajas lógicas `POST .../baja`, `PATCH .../zonas-supervision/:codigo` o comprobaciones posteriores a la baja). Sintaxis:
  ```text
  solo-nueva <rol> <METODO> <ruta> [@cuerpo.json] <status_esperado> [fragmento_esperado]
  ```
  Si `fragmento_esperado` inicia con `!`, el arnés verifica que el cuerpo de la respuesta **no** contenga el texto indicado (por ejemplo `!AV-BAJA-AUTO` para confirmar que el área dada de baja ya no aparece en `/catastro/areas` ni en `/geo/areas`).

