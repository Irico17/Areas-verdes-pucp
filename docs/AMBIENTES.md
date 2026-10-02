# Ambientes: develop, qa y producción

VerdePUCP corre el mismo backend y las mismas migraciones `db/migrations` (`001`–`048`) en tres ambientes. La base de cada ambiente es independiente. El entrypoint (`backend/docker-entrypoint.sh` mediante el binario `migrate`; el binario `api` no migra) aplica solo las migraciones que falten en `schema_migrations`. No hay `AutoMigrate`, no se monta un `.sql` en `initdb` y no se borra el catastro cargado.

El corte de la EC2 que ya tiene datos sigue en [`RUNBOOK-CORTE-PRODUCCION.md`](RUNBOOK-CORTE-PRODUCCION.md). Este documento no lo ejecuta.

## Matriz

| | develop | qa | produccion |
| --- | --- | --- | --- |
| `APP_ENV` | `develop` | `qa` | `produccion` |
| Swagger y openapi.yaml | encendido | encendido | apagado (404) |
| Logs | debug, consola | info, JSON | info, JSON |
| Pool (open / idle / vida) | 5 / 2 / 15 min | 8 / 3 / 20 min | 10 / 4 / 30 min |
| Datos si la base está vacía | semilla ficticia | semilla ficticia | ETL de `data/raw` |
| Datos si ya hay filas | no se tocan | no se tocan | no se tocan |
| Web (host) | 8088 | 8188 | 8288 |
| API (host) | 8091 | 8191 | 8291 |
| Postgres (host) | 5432 | 5433 | 5434 |
| Base local | `campus_verde_develop` | `campus_verde_qa` | `campus_verde_produccion` |
| Base en la EC2 | `campus_verde` | `campus_verde` | `campus_verde` |
| Contenedores | `campus-develop-db/api/web` | `campus-qa-*` | `campus-produccion-*` |
| Volúmenes | `campus_develop_pg`, `campus_develop_data` | `campus_qa_pg`, `campus_qa_data` | `campus_produccion_pg`, `campus_produccion_data` |
| Red | `campus-develop` | `campus-qa` | `campus-produccion` |
| Proyecto Compose | `campus-develop` | `campus-qa` | `campus-produccion` |
| Cuándo se despliega | push a `develop` o a `backend/arquitectura-equipo`, tras el CI | tag `rc-*` o a mano | solo a mano, con aprobación |
| Antes de cambiar la EC2 | — | — | snapshot EBS, `pg_dump`, conteos |

En una misma máquina los tres stacks conviven: puertos, nombres, volúmenes, redes y bases no se pisan. `make down ENV=qa` no borra volúmenes (`down` no usa `-v`).

`SWAGGER_ENABLED`, `CAMPUS_CORS_ORIGINS`, `LOG_LEVEL`, `LOG_FORMAT`, `SERVER_GIN_MODE`, `DATABASE_MAX_OPEN_CONNS`, `DATABASE_MAX_IDLE_CONNS` y `DATABASE_CONN_MAX_LIFETIME` pisan el default del ambiente. Un asterisco en CORS se descarta. Sin `APP_ENV` el proceso se queda en el comportamiento histórico (pool 10/4/30 min, Swagger según el modo de Gin). El YAML de OpenAPI (`/api/v1/openapi.yaml` y `/areas-verdes/v1/openapi.yaml`) se apaga con 404 únicamente cuando `APP_ENV=produccion`; en develop, qa y local sin variables responde 200 con independencia de `SWAGGER_ENABLED`.

La cookie `Secure` la sigue mandando `CAMPUS_COOKIE_SECURE`. En HTTP (el lab y el compose local) es `false`. Con TLS es `true`. `APP_ENV=produccion` no la enciende sola.

## Local

```bash
make up ENV=develop
make up ENV=qa
make up ENV=produccion
make smoke ENV=develop
make down ENV=qa
```

La primera vez se copia `deploy/env/<ambiente>.env.example` a `deploy/env/<ambiente>.env` (no se versiona). Las claves de esos examples son de demostración (`pando-local` en develop y qa, `produccion-local-demo-2026` en produccion, `campus-develop`, `campus-qa`). No sirven para la EC2.

### Sobrescritura de puertos desde el entorno

En máquinas de desarrollo donde el puerto `5432` ya está ocupado por una instancia local de PostgreSQL (por ejemplo, la base de datos `campus_verde`) o el puerto `8091` está tomado por la API vieja, es posible sobrescribir los puertos de publicación sin modificar los archivos de configuración ni alterar los valores por defecto:

- `POSTGRES_PORT`: puerto de PostgreSQL en el host (por defecto: develop `5432`, qa `5433`, produccion `5434`).
- `API_PORT`: puerto de la API en el host (por defecto: develop `8091`, qa `8191`, produccion `8291`).
- `WEB_PORT`: puerto del frontend/nginx en el host (por defecto: develop `8088`, qa `8188`, produccion `8288`).

**Precedencia:** variables exportadas en el entorno del llamador > variables definidas en `deploy/env/<ambiente>.env` > valores por defecto en `deploy/env/<ambiente>.env.example`.

Ejemplo para levantar develop en una máquina con `5432` y `8091` ocupados:

```bash
POSTGRES_PORT=5442 API_PORT=8093 WEB_PORT=8089 make up ENV=develop
POSTGRES_PORT=5442 API_PORT=8093 WEB_PORT=8089 make smoke ENV=develop
POSTGRES_PORT=5442 API_PORT=8093 WEB_PORT=8089 make rollback ENV=develop
POSTGRES_PORT=5442 API_PORT=8093 WEB_PORT=8089 make down ENV=develop
```

Esta misma precedencia aplica invocando directamente `deploy/deploy.sh`, `deploy/smoke.sh` y `deploy/conteos.sh`.

`make up` sin `ENV` sigue levantando solo el Postgres del `docker-compose.yml` histórico. `make bootstrap` también. El detalle del compose por ambiente está en [`deploy/README.md`](../deploy/README.md).

Smoke (`deploy/smoke.sh`): `GET /health` con `"status":"ok"`, `POST /api/v1/sesion` con la cuenta ficticia `coordinacion`, `GET /api/v1/geo/resumen` y `GET /api/v1/catalogos`. En develop y qa Swagger y openapi.yaml responden; en produccion responden 404.

Rollback local, si el smoke falla y había una imagen anterior de ese ambiente:

```bash
bash deploy/deploy.sh develop --local --rollback
# O indicando un TAG explícito para la API (la imagen web se toma del estado previo):
bash deploy/deploy.sh develop --local --rollback <tag-api>
```

El rollback local revierte tanto la imagen de la API como la de la web (`state/<ambiente>.prev-image` y `state/<ambiente>.prev-image-web`), retagueando ambas y ejecutando `compose up -d --no-build`. Si se pasa un `TAG` explícito a `--rollback`, este aplica a la imagen de la API y la web se recupera del estado guardado (o falla con error si no existe). En `up_local` (`deploy.sh --local` o `make up ENV=...`), si el smoke falla tras levantar los contenedores, se ejecuta este mismo rollback automático a las imágenes previas de API y web.

**Las migraciones son solo hacia adelante:** el rollback revierte los contenedores y sus imágenes de código (API y web), pero **no** revierte la base de datos ni elimina columnas o tablas añadidas por migraciones (no existen migraciones «down»). La imagen anterior corre sobre el esquema de base de datos ya migrado, lo cual es compatible gracias a que todas las migraciones en `db/migrations` son aditivas (columnas nuevas nulas o con `DEFAULT`).

## Qué datos entran

- **develop y qa.** `SEED_PROFILE=etl`, igual que producción: si el catastro está vacío se carga `data/raw` (521 áreas, 534 zonas, capas e inventario) para que el mapa tenga datos reales de forma. Si ya hay filas, no se ejecuta. La semilla mínima `SEED_PROFILE=ficticio` (`deploy/seed/ficticio.sql`) sigue disponible para pruebas rápidas. Los datos de usuarios/cuentas siguen siendo ficticios y las claves son distintas por ambiente.
- **produccion.** `SEED_PROFILE=etl`. La carga de `data/raw` corre solo si el catastro está vacío y no existe `/data/.etl-done`. Si `areas_verdes` ya tiene filas, no corre.

No copiar la base de producción a develop ni a qa. Si hiciera falta un ensayo, se restaura en una base desechable, se enmascaran nombres, cuentas y evidencias, y recién entonces se usa. No hay un script en este repo que haga esa copia.

Las cuentas `norte`, `sur`, `riego`, `coordinacion`, `jefatura` y `admin` son ficticias. En develop, qa y en local sin `APP_ENV` la clave documentada es `pando-local`. En producción (`APP_ENV=produccion`), tanto local como en la EC2, se rechazan claves de laboratorio (`pando-local`, `campus-lab`) o de menos de 16 caracteres en el arranque (en `docker-entrypoint.sh`, en la API y en las migraciones); el example local usa `produccion-local-demo-2026`. En la EC2 la clave es `TF_VAR_dev_password`, distinta en cada ambiente, de 16 caracteres o más, y no puede ser `pando-local` ni `campus-lab`.

## Imágenes

El workflow `ci` publica, en cada push que no es un pull request, dos imágenes inmutables:

- `ghcr.io/<owner>/campus-verde-api:<sha>`
- `ghcr.io/<owner>/campus-verde-web:<sha>`

`<sha>` es el commit completo (40 hex). Un tag `rc-*` publica además ese nombre de tag. No se publica `:latest` como tag de rollback: el rollback usa el SHA (o el `rc-*`).

`GITHUB_TOKEN` alcanza para publicar y para que el job de deploy lea el registro. No hace falta un secret extra de GHCR. El owner va en minúsculas.

## GitHub Actions

| Workflow | Qué hace |
| --- | --- |
| `.github/workflows/ci.yml` | Pruebas de `apps/api`, web, backend con PostGIS, escaneo de seguridad (Trivy fs/image, Gitleaks), prohibición de `AutoMigrate`/borrados/initdb, y publicación de imágenes inmutables por SHA en GHCR |
| `.github/workflows/deploy.yml` | Despliegue en runner self-hosted por ambiente. Valida CI en verde, controla promoción estricta por SHA, solicita aprobación en producción y ejecuta `deploy/host-deploy.sh` |

Disparo de `deploy.yml`:

- **develop:** push automático a `develop` (tras CI en verde) o `workflow_dispatch`.
- **qa:** push de tag `rc-*` o `workflow_dispatch` (exige promoción: despliegue previo exitoso en `develop`).
- **produccion:** solo `workflow_dispatch` manual (exige promoción: despliegue previo exitoso en `qa`, y aprobación humana obligatoria).

### Despliegue con Runners Self-Hosted

Cada instancia EC2 de VerdePUCP corre un runner self-hosted de GitHub Actions con etiquetas fijas según el ambiente:
- **develop:** `[self-hosted, campus-develop]`
- **qa:** `[self-hosted, campus-qa]`
- **produccion:** `[self-hosted, campus-prod]`

El job `deploy` corre directamente dentro de la máquina destino. No requiere credenciales AWS ni variables de Terraform en los secretos de GitHub Actions. El runner ejecuta:

```bash
bash deploy/host-deploy.sh <ambiente> --sha <sha> [--rollback <tag>]
```

`host-deploy.sh` lee su configuración del archivo local `$CAMPUS_HOME/host.env` en la EC2, autentica en GHCR, realiza backup previo en producción, actualiza contenedores, valida con smoke test local y gestiona el rollback automático en caso de fallo.

**Sin omisiones artificiales:** Si el runner self-hosted del ambiente no está disponible, el job permanece en cola de GitHub Actions hasta que se inicie o venza el tiempo límite (`timeout-minutes: 30`). No hay estados "omitido en verde".

### Guía de Promoción entre Ambientes

El flujo de despliegue exige avanzar el mismo SHA verificado a través de los ambientes:

1. **Despliegue inicial en develop:**
   Se produce automáticamente tras hacer push a la rama `develop` una vez que el workflow `ci.yml` finaliza exitosamente.

2. **Promoción a QA:**
   Una vez verificado el funcionamiento en develop, se dispara la promoción hacia qa indicando el SHA:
   ```bash
   gh workflow run deploy.yml -f ambiente=qa -f ref=<sha>
   ```
   O creando un tag de release candidate sobre dicho commit (`git tag rc-1.0.0 <sha> && git push origin rc-1.0.0`).
   El workflow consulta la API de Deployments de GitHub (`gh api repos/:owner/:repo/deployments?sha=<sha>&environment=develop`) y exige que el commit tenga un status `success` en `develop`. Si nunca se desplegó en develop o falló, la promoción se rechaza con error explícito.

3. **Promoción a Producción:**
   Verificado QA, se promueve el mismo commit exacto a producción:
   ```bash
   gh workflow run deploy.yml -f ambiente=produccion -f ref=<sha>
   ```
   El workflow exige que el commit tenga un despliegue previo con status `success` en `qa`.

4. **Aprobación manual de Producción:**
   El ambiente `produccion` está protegido por la política de **Required reviewers**. Al lanzarse el workflow, el job `deploy` entra en estado de espera y notifica a los revisores. Un revisor autorizado debe aprobar el despliegue desde la interfaz de GitHub Actions o vía CLI:
   ```bash
   gh run review <run-id> --approve -c "Despliegue a producción autorizado tras verificación en QA"
   ```
   Adicionalmente, se puede activar un temporizador de espera (*wait-timer*) de cortesía (por ejemplo, 5 minutos) para permitir cancelaciones de último momento.

5. **Rollback:**
   Si se requiere volver a una versión previa:
   ```bash
   gh workflow run deploy.yml -f ambiente=<qa|produccion> -f rollback=<sha-anterior-o-tag-rc>
   ```
   En caso de rollback explícito, el workflow omite la exigencia de despliegue previo del commit por promoción y valida que el tag corresponda a un SHA de 40 dígitos hexadecimales o un tag `rc-*`.

### Variables por Environment en GitHub

El workflow espera las siguientes variables configuradas en cada ambiente de GitHub (**Settings → Environments**):

| Variable | Dónde | Propósito | Ejemplo |
| --- | --- | --- | --- |
| `PUBLIC_URL` | Cada environment | URL pública del servicio. Utilizada para registrar enlaces en el resumen del despliegue y para la verificación post-despliegue (`GET /health`) | `http://198.51.100.1:8088` o `https://campusverde.pucp.edu.pe` |

### Configuración requerida del Operador (con `gh`)

El operador del repositorio debe configurar las protecciones y variables mediante `gh` o la consola web de GitHub (referencia informativa; no ejecutar en pipelines automatizados):

```bash
# 1. Crear los ambientes
gh api -X PUT repos/:owner/:repo/environments/develop
gh api -X PUT repos/:owner/:repo/environments/qa
gh api -X PUT repos/:owner/:repo/environments/produccion

# 2. Configurar revisores requeridos (Required Reviewers) y wait-timer en producción
gh api -X PUT repos/:owner/:repo/environments/produccion \
  -f 'reviewers[][type]=User' -F 'reviewers[][id]=<user-id-o-equipo>' \
  -F 'wait_timer=5'

# 3. Restricción de ramas permitidas para despliegue en producción (Branch Policy)
gh api -X PUT repos/:owner/:repo/environments/produccion \
  -F 'deployment_branch_policy[protected_branches]=true' \
  -F 'deployment_branch_policy[custom_branch_policies]=false'

# 4. Asignar variable PUBLIC_URL por ambiente
gh variable set PUBLIC_URL --env develop --body 'http://<ip-develop>:8088'
gh variable set PUBLIC_URL --env qa --body 'http://<ip-qa>:8188'
gh variable set PUBLIC_URL --env produccion --body 'http://<ip-produccion>:8288'
```

*Nota sobre seguridad de Forks:* En **Settings → Actions → General → Fork pull request workflows**, asegúrese de tener configurado "Require approval for all outside collaborators" para prevenir ejecución no autorizada de acciones en runners del repositorio.

En producción el script (`deploy/deploy.sh produccion --aws`) ejecuta el siguiente orden estricto:

1. `terraform init` y selección de workspace `produccion`.
2. **Antes de `terraform apply`** (sobre la instancia existente en el estado de Terraform, obtenida con `terraform output -raw instance_id`):
   - Snapshot EBS del volumen `campus-verde-data` (filtrado por dicha instancia).
   - Backup lógico `pg_dump -Fc` en `/opt/campus/data/backups/` dentro de la instancia vía SSM.
   - Conteos «antes» de las tablas (el SQL de [`deploy/conteos.sql`](../deploy/conteos.sql), el del §4.4).
   *Nota de seguridad:* Si en producción no existe instancia previa en el estado de Terraform (primera creación de la infraestructura), el despliegue falla de forma segura salvo que se declare la variable de entorno `DEPLOY_PRIMERA_VEZ=1`. En develop y qa no se exige snapshot previo.
3. `terraform apply -auto-approve`: aplica cambios de infraestructura.
4. Construcción y subida de imágenes inmutables (tag = SHA) a ECR.
5. Inyección de secretos en la instancia (`scripts/poner-secretos.sh`).
6. Actualización del compose en la instancia (`patch_compose.py`) y arranque de contenedores (`docker compose pull` y `up -d`).
7. Smoke test (`deploy/smoke.sh`): verifica `/health`, login, lecturas y Swagger apagado. Si falla, revierte automáticamente a las imágenes previas (`PREVIOUS_API` / `PREVIOUS_WEB`).
8. Conteos «después» de tablas y comparación (`deploy/comparar_conteos.py`): valida que las tablas de negocio mantengan exactamente sus filas.

La comparación (`deploy/comparar_conteos.py`) valida que las tablas de datos de negocio mantengan exactamente sus filas. Se excluyen explícitamente las tablas técnicas que cambian de forma legítima durante el despliegue, definidas en [`deploy/conteos.excluir`](../deploy/conteos.excluir) (fuente única de verdad): `schema_migrations` (por migraciones pendientes aplicadas por el entrypoint) y `sesiones` (por el login del smoke test). Tablas nuevas introducidas por migraciones se reportan. Si una tabla de negocio diverge o el smoke falla, se aborta y se revierte a la imagen anterior. No hay migración «down»: las migraciones solo agregan. Volver el binario atrás no borra columnas. Restaurar el dump es el último recurso y lo decide una persona; se pierde lo cargado después del backup.

## Secretos y variables

Créelos en **Settings → Environments** (o `gh secret set` / `gh variable set`) para `develop`, `qa` y `produccion`. Un secret de repositorio con el mismo nombre sigue sirviendo de respaldo; el del environment gana. Producción tiene que tener sus propias claves, distintas de develop y de qa.

### Secrets (los cinco, en cada environment)

| Nombre | Qué es |
| --- | --- |
| `AWS_ACCESS_KEY_ID` | clave temporal del Learner Lab |
| `AWS_SECRET_ACCESS_KEY` | clave temporal del Learner Lab |
| `AWS_SESSION_TOKEN` | token de sesión del lab (caduca con la sesión) |
| `TF_VAR_db_password` | clave de Postgres, 16 caracteres o más; no `campus-lab` ni `pando-local` |
| `TF_VAR_dev_password` | clave de las cuentas locales, 16 caracteres o más; no `campus-lab` ni `pando-local` |

```bash
gh secret set AWS_ACCESS_KEY_ID --env develop
gh secret set AWS_SECRET_ACCESS_KEY --env develop
gh secret set AWS_SESSION_TOKEN --env develop
gh secret set TF_VAR_db_password --env develop
gh secret set TF_VAR_dev_password --env develop
```

Repita con `--env qa` y `--env produccion`. No pegue los valores en el repo ni en el chat.

### Variables (no son secretos)

| Nombre | Dónde | Ejemplo |
| --- | --- | --- |
| `CAMPUS_CORS_ORIGINS` | cada environment | orígenes separados por coma, nunca `*` |
| `CAMPUS_COOKIE_SECURE` | cada environment | `false` en el lab por HTTP; `true` solo con TLS |
| `AWS_REGION` | opcional | `us-east-1` si se omite |

```bash
gh variable set CAMPUS_CORS_ORIGINS --env develop --body 'http://127.0.0.1:8088'
gh variable set CAMPUS_COOKIE_SECURE --env produccion --body 'false'
gh variable set AWS_REGION --env produccion --body 'us-east-1'
```

### Aprobación

1. Settings → Environments → New environment → `produccion`.
2. Required reviewers: al menos una persona.
3. Opcional: lo mismo en `qa`. `develop` no lleva revisores.

Hasta que eso esté guardado, GitHub puede crear el environment en el primer uso **sin** revisores. No dispare producción antes de activar la protección.

## Terraform

Los archivos están en `infra/terraform`. `ambiente` vale `develop`, `qa` o `produccion` (default `produccion`, así los nombres del lab siguen siendo `campus-verde`, `campus-verde-api` y `campus-verde-data`).

Cada ambiente va en su workspace. El estado sigue en el cubo S3 que ya documenta [`DEPLOY-AWS.md`](DEPLOY-AWS.md). Este cambio no ejecuta `init`, `plan`, `apply` ni `workspace new`.

```bash
cd infra/terraform
terraform fmt -check -recursive
terraform init -backend=false
terraform validate
```

Ejemplos sin secretos: `infra/terraform/environments/develop.tfvars.example`, `qa.tfvars.example`, `produccion.tfvars.example`. Las claves entran solo por `TF_VAR_db_password` y `TF_VAR_dev_password`.

El stack que ya está desplegado vive en el workspace `default`. Trátelo como producción hasta que alguien mueva el estado a mano. `user_data_replace_on_change` sigue en false: un apply que solo cambia la imagen no reemplaza la EC2. En una instancia ya creada, cloud-init no se repite; `deploy/deploy.sh` reescribe el tag de la imagen en el compose de la instancia.

`scripts/deploy-learner-lab.sh` es el atajo de producción: exige las tres credenciales y llama a `deploy/deploy.sh produccion --aws`.

## Flujo develop → qa → producción

1. Se integra en `develop`. El CI prueba y publica el SHA. Si queda en verde, `deploy.yml` despliega ese SHA a develop.
2. Cuando develop está aceptado, se promueve a QA vía `workflow_dispatch` indicando el SHA o etiquetando `rc-<algo>` en ese commit. El CI valida y publica el tag, y `deploy.yml` despliega qa tras comprobar el despliegue previo exitoso en develop.
3. Producción: una persona dispara `workflow_dispatch` con `ambiente=produccion` y el `ref` del SHA que pasó por qa. El workflow exige que el commit tenga despliegue exitoso en qa y GitHub pide la aprobación del environment `produccion` (Required reviewers).
4. Para volver atrás: el mismo workflow con `rollback` = el SHA o tag `rc-*` que estaba sirviendo (omite verificación previa por promoción). No se revierten columnas.

El checklist largo del corte (ensayo en copia, evidencias, tag anotado) sigue siendo el de [`RUNBOOK-CORTE-PRODUCCION.md`](RUNBOOK-CORTE-PRODUCCION.md). El workflow automatiza el backup, el snapshot, los conteos y el smoke; no sustituye la decisión de la persona que aprueba.

## Despliegue en el propio host (Runners self-hosted)

Para desacoplar el despliegue de credenciales temporales de AWS desde GitHub, los ambientes cuentan con runners self-hosted en cada EC2 (labels `campus-develop`, `campus-qa`, `campus-prod`) que ejecutan [`deploy/host-deploy.sh`](../deploy/host-deploy.sh):

- **EC2 Compartida:** aloja develop (puerto web 8088, API 8091, Postgres 5432, base `campus_verde_develop`, directorio `/opt/campus/develop`) y qa (puerto web 8188, API 8191, Postgres 5433, base `campus_verde_qa`, directorio `/opt/campus/qa`) con stacks aislados.
- **EC2 Producción:** aloja producción (puerto web 80, API 8091 en localhost, Postgres 5432 en localhost, base `campus_verde`, directorio `/opt/campus`).
- **Secretos locales:** se leen exclusivamente de `$CAMPUS_HOME/host.env` (permisos 600, no versionado). No se inyectan contraseñas desde GitHub Actions.
- **Seguridad en producción:** snapshot EBS del volumen `campus-verde-data` vía IMDSv2, backup lógico `pg_dump` y conteos antes/después con [`deploy/comparar_conteos.py`](../deploy/comparar_conteos.py).
- **Rollback:** automático ante fallos de smoke test o manual indicando el tag previo.

Consulte la arquitectura completa, procedimientos de migración y rotación de claves en [`DEPLOY-RUNNER.md`](DEPLOY-RUNNER.md).

