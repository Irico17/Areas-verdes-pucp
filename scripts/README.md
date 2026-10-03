# scripts

| Script | Qué hace |
|--------|----------|
| `actualizar-credenciales-lab.sh` / `.ps1` | Lee el bloque AWS CLI del Learner Lab (stdin o portapapeles) y lo sube al environment `aws-lab` sin imprimir valores. Ver `docs/CAMBIO-DE-CUENTA-LAB.md` |
| `set-aws-secrets.sh` / `set-aws-secrets.ps1` | Sube las tres credenciales temporales del Learner Lab a los secrets del repo. Para una cuenta nueva prefiera `actualizar-credenciales-lab` |
| `bootstrap-estado-terraform.sh` | Crea el cubo de estado `campus-verde-tfstate-<cuenta>` y el cubo de transferencia SSM en la cuenta actual |
| `aprovisionar-cuenta.sh` | plan, apply o destroy parametrizado. Lo llama el workflow `aprovisionar-cuenta` |
| `pausar-cuenta.sh` | Enciende o detiene instancias por etiquetas |
| `generar-inventario-ansible.sh` | Inventario Ansible desde `terraform output -json` |
| `recuperar-estado-terraform.sh` | Imprime `terraform import` si el estado se perdió y los recursos siguen. No ejecuta nada |
| `bootstrap.sh` | `docker compose up -d`, espera, migra y corre el ETL |
| `wait-db.sh` | Espera a que `pg_isready` responda en el servicio `db` |
| `counts.sh` | Conteos de áreas, zonas y capas, con SRID 4326 |
| `runner-instalar.sh` | Instala y registra el runner self-hosted de GitHub Actions en la EC2 (`develop`, `qa` o `produccion`). Lee token de `RUNNER_TOKEN` o SSM, verifica SHA-256 e instala servicio systemd. |
| `cargar-secretos-github.ps1` | Carga en GitHub (gh secret set / gh variable set) los secretos y variables de `.env-github`, que es local y no se versiona. `-DryRun` solo lista los nombres. No contiene valores. |
| `generar-esquema-bd.sh` | Aplica `db/migrations` y escribe `db/esquema.sql` y `docs/BASE-DE-DATOS.md`. `--comprobar` falla si hay deriva. |

El ETL que usa el Makefile está en `backend/app` (`cmd/etl`, `cmd/etl-lote`, `cmd/sectores`, `cmd/migrate`). Desde la raíz, `make bootstrap`, `make etl`, `make etl-lote` y `make counts` llaman a estos scripts o a ese módulo.

