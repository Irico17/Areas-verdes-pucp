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

