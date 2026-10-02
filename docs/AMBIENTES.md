# Ambientes: develop, qa y producción

VerdePUCP corre el mismo backend y las mismas migraciones `db/migrations` (`001`–`048`) en tres ambientes. La base de cada ambiente es independiente. El entrypoint (`backend/docker-entrypoint.sh` mediante el binario `migrate`; el binario `api` no migra) aplica solo las migraciones que falten en `schema_migrations`. No hay `AutoMigrate`, no se monta un `.sql` en `initdb` y no se borra el catastro cargado.

El corte de la EC2 que ya tiene datos sigue en [`RUNBOOK-CORTE-PRODUCCION.md`](RUNBOOK-CORTE-PRODUCCION.md). Este documento no lo ejecuta.

## Matriz

| | develop | qa | produccion |
| --- | --- | --- | --- |
| `APP_ENV` | `develop` | `qa` | `produccion` |
| Swagger | encendido | encendido | apagado |
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

`SWAGGER_ENABLED`, `CAMPUS_CORS_ORIGINS`, `LOG_LEVEL`, `LOG_FORMAT`, `SERVER_GIN_MODE`, `DATABASE_MAX_OPEN_CONNS`, `DATABASE_MAX_IDLE_CONNS` y `DATABASE_CONN_MAX_LIFETIME` pisan el default del ambiente. Un asterisco en CORS se descarta. Sin `APP_ENV` el proceso se queda en el comportamiento histórico (pool 10/4/30 min, Swagger según el modo de Gin).

La cookie `Secure` la sigue mandando `CAMPUS_COOKIE_SECURE`. En HTTP (el lab y el compose local) es `false`. Con TLS es `true`. `APP_ENV=produccion` no la enciende sola.

## Local

```bash
make up ENV=develop
make up ENV=qa
make up ENV=produccion
make smoke ENV=develop
make down ENV=qa
```

La primera vez se copia `deploy/env/<ambiente>.env.example` a `deploy/env/<ambiente>.env` (no se versiona). Las claves de esos examples son de demostración (`pando-local`, `campus-develop`, `campus-qa`, `campus-produccion-local`). No sirven para la EC2.

`make up` sin `ENV` sigue levantando solo el Postgres del `docker-compose.yml` histórico. `make bootstrap` también. El detalle del compose por ambiente está en [`deploy/README.md`](../deploy/README.md).

Smoke (`deploy/smoke.sh`): `GET /health` con `"status":"ok"`, `POST /api/v1/sesion` con la cuenta ficticia `coordinacion`, `GET /api/v1/geo/resumen` y `GET /api/v1/catalogos`. En develop y qa Swagger responde; en produccion responde 404.

Rollback local, si el smoke falla y había una imagen anterior de ese ambiente:

```bash
bash deploy/deploy.sh develop --local --rollback
# O indicando un TAG explícito para la API (la imagen web se toma del estado previo):
bash deploy/deploy.sh develop --local --rollback <tag-api>
```

El rollback local revierte tanto la imagen de la API como la de la web (`state/<ambiente>.prev-image` y `state/<ambiente>.prev-image-web`), retagueando ambas y ejecutando `compose up -d --no-build`. Si se pasa un `TAG` explícito a `--rollback`, este aplica a la imagen de la API y la web se recupera del estado guardado (o falla con error si no existe). En `up_local` (`deploy.sh --local` o `make up ENV=...`), si el smoke falla tras levantar los contenedores, se ejecuta este mismo rollback automático a las imágenes previas de API y web.

**Las migraciones son solo hacia adelante:** el rollback revierte los contenedores y sus imágenes de código (API y web), pero **no** revierte la base de datos ni elimina columnas o tablas añadidas por migraciones (no existen migraciones «down»). La imagen anterior corre sobre el esquema de base de datos ya migrado, lo cual es compatible gracias a que todas las migraciones en `db/migrations` son aditivas (columnas nuevas nulas o con `DEFAULT`).

## Qué datos entran

- **develop y qa.** `SEED_PROFILE=ficticio`. Si el catastro está vacío, se insertan el Jardín Ficticio Norte y el Sector Ficticio Norte (`deploy/seed/ficticio.sql`). Si ya hay filas, no se ejecuta. No es un volcado de producción.
- **produccion.** `SEED_PROFILE=etl`. La carga de `data/raw` corre solo si el catastro está vacío y no existe `/data/.etl-done`. Si `areas_verdes` ya tiene filas, no corre.

No copiar la base de producción a develop ni a qa. Si hiciera falta un ensayo, se restaura en una base desechable, se enmascaran nombres, cuentas y evidencias, y recién entonces se usa. No hay un script en este repo que haga esa copia.

Las cuentas `norte`, `sur`, `riego`, `coordinacion`, `jefatura` y `admin` son ficticias. La clave local documentada es `pando-local`. En la EC2 la clave es `TF_VAR_dev_password`, distinta en cada ambiente, de 16 caracteres o más, y no puede ser `pando-local` ni `campus-lab`.

## Imágenes

El workflow `ci` publica, en cada push que no es un pull request, dos imágenes inmutables:

- `ghcr.io/<owner>/campus-verde-api:<sha>`
- `ghcr.io/<owner>/campus-verde-web:<sha>`

`<sha>` es el commit completo (40 hex). Un tag `rc-*` publica además ese nombre de tag. No se publica `:latest` como tag de rollback: el rollback usa el SHA (o el `rc-*`).

`GITHUB_TOKEN` alcanza para publicar y para que el job de deploy lea el registro. No hace falta un secret extra de GHCR. El owner va en minúsculas.

## GitHub Actions

| Workflow | Qué hace |
| --- | --- |
| `.github/workflows/ci.yml` | pruebas de `apps/api`, web, backend, prohibición de `AutoMigrate` / borrados / initdb, y publicación de las imágenes por SHA |
| `.github/workflows/deploy.yml` | despliegue. Espera a que el CI de ese SHA esté en verde |

Disparo automático de `deploy.yml`:

- push a `develop` o a `backend/arquitectura-equipo` → ambiente `develop`
- tag `rc-*` → ambiente `qa`
- producción no sale de un push

A mano:

```bash
gh workflow run deploy --ref backend/arquitectura-equipo -f ambiente=qa -f ref=<sha>
gh workflow run deploy --ref backend/arquitectura-equipo -f ambiente=produccion -f ref=<sha>
gh workflow run deploy --ref backend/arquitectura-equipo -f ambiente=produccion -f rollback=<sha-anterior>
```

`ref` vacío usa el commit de la rama desde la que se dispara el workflow. No dispare el workflow contra `main` mientras `deploy.yml` no esté en esa rama.

El job de deploy usa `environment:` con el nombre del ambiente (`develop`, `qa`, `produccion`). La aprobación manual de producción es la protección **Required reviewers** de ese environment. El YAML no puede crearla: hay que ponerla en GitHub antes del primer despliegue real.

Si faltan las credenciales de AWS, develop y qa terminan en verde sin desplegar (el lab suele estar cerrado). Producción falla si faltan, para no marcar como hecho un despliegue que no ocurrió.

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

1. Se integra en `develop` o en `backend/arquitectura-equipo`. El CI prueba y publica el SHA. Si queda en verde, `deploy.yml` despliega ese SHA a develop.
2. Cuando develop está aceptado, se etiqueta `rc-<algo>` en ese commit. El CI publica el tag y `deploy.yml` despliega qa.
3. Producción: una persona dispara `workflow_dispatch` con `ambiente=produccion` y el `ref` del SHA que pasó por qa. GitHub pide la aprobación del environment. El job hace snapshot, backup, conteos, despliegue, smoke y conteos otra vez.
4. Para volver atrás: el mismo workflow con `rollback` = el SHA que estaba sirviendo. No se revierten columnas.

El checklist largo del corte (ensayo en copia, evidencias, tag anotado) sigue siendo el de [`RUNBOOK-CORTE-PRODUCCION.md`](RUNBOOK-CORTE-PRODUCCION.md). El workflow automatiza el backup, el snapshot, los conteos y el smoke; no sustituye la decisión de la persona que aprueba.
