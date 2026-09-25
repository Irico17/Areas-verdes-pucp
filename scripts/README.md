# scripts

| Script | Qué hace |
|--------|----------|
| `set-aws-secrets.sh` / `set-aws-secrets.ps1` | Sube las tres credenciales temporales del Learner Lab a los secrets del repo. Ver `docs/DESPLIEGUE.md` |
| `bootstrap.sh` | `docker compose up -d`, espera, migra y corre el ETL |
| `wait-db.sh` | Espera a que `pg_isready` responda en el servicio `db` |
| `counts.sh` | Conteos de áreas, zonas y capas, con SRID 4326 |

El ETL en sí está en `apps/api/cmd/etl` (Go). Desde la raíz, `make bootstrap`, `make etl` y `make counts` llaman a estos scripts o al módulo.
