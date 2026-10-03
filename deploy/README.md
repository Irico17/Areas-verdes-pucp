# Despliegue

Este documento junta el pipeline que hoy está en el repositorio: CI, el workflow `deploy`, los runners, la promoción develop → QA → producción, HTTPS con DuckDNS, el rollback y la verificación. No contiene claves.

La mudanza de cuenta del laboratorio está en [docs/CAMBIO-DE-CUENTA-LAB.md](../docs/CAMBIO-DE-CUENTA-LAB.md). Las protecciones de GitHub están en [docs/PROTECCIONES-REPO.md](../docs/PROTECCIONES-REPO.md). El aprovisionamiento de las máquinas está en [infra/README.md](../infra/README.md).

## Qué corre dónde

```mermaid
flowchart LR
  push["push a develop o tag rc-*"] --> ci["ci.yml"]
  ci --> ghcr["ghcr.io imágenes por SHA"]
  ci --> deploy["deploy.yml"]
  deploy --> promo{"promoción del mismo SHA"}
  promo --> dev["runner campus-develop"]
  promo --> qa["runner campus-qa"]
  promo --> prod["runner campus-prod<br/>con aprobación"]
  dev --> host["deploy/host-deploy.sh"]
  qa --> host
  prod --> host
```

GitHub construye y publica las imágenes. El servidor no usa credenciales de AWS para desplegar: el runner de la propia EC2 hace `docker pull` desde GHCR y levanta Compose. Las credenciales del Learner Lab solo hacen falta para aprovisionar (Terraform) o para entrar por SSM.

## Ramas

| Rama o tag | Efecto |
| --- | --- |
| `develop` | Integración. El push dispara CI y, si queda en verde, despliega develop. |
| `main` | Rama por defecto. No despliega. Avanza por fast-forward desde `develop`. |
| tag `rc-*` | Dispara CI y despliega QA. La imagen también queda etiquetada con ese tag, útil para un rollback. |

Un cambio que solo toca `*.md` o `docs/**` no dispara CI ni despliegue (`paths-ignore`).

## CI

Workflow: `.github/workflows/ci.yml`.

| Job | Qué hace |
| --- | --- |
| `test` | Construcción de las imágenes `campus-verde-api` (`backend/dockerfile`) y `campus-verde-web` (`frontend/Dockerfile`). |
| `frontend` | En `frontend/`: `npm ci`, `oxlint --max-warnings 0`, pruebas unitarias y `tsc -b && vite build`. |
| `backend` | `gofmt`, `go vet` y `go test` de `backend/app` contra PostGIS. Prohíbe `AutoMigrate`, borrados en migraciones desde la 047, SQL montado en `initdb` y borrados en `deploy/seed`. Prueba los scripts de `deploy/`. Regenera Swagger y falla si hay deriva. |
| `escaneo` | Gitleaks y Trivy. No bloquea el pipeline. |
| `imagenes` | Solo en push a `develop` o en tag `rc-*`. Publica `ghcr.io/<dueño>/campus-verde-api:<sha>` y `campus-verde-web:<sha>`. No publica `:latest`. |
| `e2e` | Migra, aplica la semilla ficticia, arranca la API y corre Playwright (`npm run test:e2e`). |

Los permisos son de lectura, salvo `packages: write` en el job de imágenes. Las acciones de terceros van fijadas por SHA. Los workflows de despliegue no se disparan con `pull_request`.

## Workflow de despliegue

`.github/workflows/deploy.yml`. Jobs: `resolver` → `esperar-ci` → `promocion` → `deploy` → `verificar-post-deploy`.

1. **resolver** decide ambiente, SHA y etiqueta del runner. Producción solo entra por `workflow_dispatch`.
2. **esperar-ci** espera a que el workflow `ci` de ese SHA termine en verde y a que el job `frontend` haya concluido en success (`deploy/esperar_ci.sh`).
3. **promocion**: QA exige un despliegue exitoso de ese SHA en develop. Producción exige uno exitoso en QA. Un rollback explícito se salta esta exigencia.
4. **deploy** corre en `[self-hosted, campus-develop|campus-qa|campus-prod]`, entra a GHCR con el token del run y ejecuta `deploy/host-deploy.sh`. El environment de GitHub se llama igual que el ambiente. Producción espera la aprobación configurada en ese environment.
5. **verificar-post-deploy** pide `GET $PUBLIC_URL/health` desde un runner hospedado. El smoke que decide es el del propio host. Si la URL no es alcanzable desde GitHub, el paso avisa y no sustituye al smoke.

Concurrencia por ambiente: no hay dos despliegues del mismo ambiente a la vez. Si el runner está apagado, el job queda en cola hasta el `timeout-minutes: 30`. No se marca en verde por ausencia.

Etiquetas por defecto: `campus-develop`, `campus-qa`, `campus-prod`. Se pueden sustituir con las variables de repositorio `RUNNER_LABEL_DEVELOP`, `RUNNER_LABEL_QA` y `RUNNER_LABEL_PRODUCCION`. La URL pública sale de la variable `PUBLIC_URL` del environment, no de una IP escrita en el workflow.

### Cómo dispararlo

Develop, automático al empujar `develop` con CI en verde. A mano:

```bash
gh workflow run deploy.yml --repo Irico17/Areas-verdes-pucp --ref develop \
  -f ambiente=develop -f ref=<sha-o-rama>
```

QA, con el mismo SHA que ya está en develop:

```bash
gh workflow run deploy.yml --repo Irico17/Areas-verdes-pucp --ref develop \
  -f ambiente=qa -f ref=<sha40>
```

O con un tag: `git tag rc-2026-10-02-1 <sha> && git push origin rc-2026-10-02-1`.

Producción, solo a mano y solo desde la rama `develop`:

```bash
gh workflow run deploy.yml --repo Irico17/Areas-verdes-pucp --ref develop \
  -f ambiente=produccion -f ref=<sha40>
```

El job queda esperando aprobación. En la web: Actions → el run → Review deployments → produccion → Approve and deploy. Por API, el identificador del environment pendiente sale de:

```bash
gh api repos/Irico17/Areas-verdes-pucp/actions/runs/<run_id>/pending_deployments
```

## Qué hace host-deploy.sh

```bash
deploy/host-deploy.sh <develop|qa|produccion> --sha <40hex> [--rollback <sha40|rc-*>] [--dry-run]
```

1. Lee `$CAMPUS_HOME/host.env` (modo 600, fuera del repositorio).
2. Descarga las imágenes de ese SHA.
3. En producción, antes de cambiar contenedores: conteos, `pg_dump -Fc` en `/opt/campus/data/backups/` y snapshot EBS del volumen `campus-verde-data` (IMDSv2 y el rol de la instancia). Si el snapshot no se puede crear, aborta, salvo `DEPLOY_SIN_SNAPSHOT=1` con el volcado ya verificado.
4. Guarda las imágenes en uso. En el primer despliegue de producción detiene el stack antiguo con `stop`, sin `down -v`, y conserva su compose.
5. `docker compose up -d` con datos en bind mounts (`pg/` y `app/`). No borra volúmenes.
6. Smoke (`deploy/smoke.sh`): `GET /health` con `"status":"ok"`, login de `coordinacion`, `GET /api/v1/geo/resumen` y `GET /api/v1/catalogos`. En develop y QA, Swagger responde. En producción, Swagger y el YAML responden 404.
7. Si el smoke falla, rollback automático a las imágenes previas.
8. En producción, `deploy/comparar_conteos.py` compara conteos. Una baja, o una tabla de negocio que desaparece, aborta el despliegue. Los aumentos y las tablas nuevas se anotan y no fallan. Se excluyen `schema_migrations` y `sesiones` (`deploy/conteos.excluir`).

Las migraciones las aplica el entrypoint al arrancar. Son aditivas. El rollback de imagen no revierte el esquema.

### Layout en el servidor

| Ambiente | Directorio | Web pública | API | Postgres |
| --- | --- | --- | --- | --- |
| develop | `/opt/campus/develop` | 8088 | 127.0.0.1:8091 | 127.0.0.1:5432 |
| QA | `/opt/campus/qa` | 8188 | 127.0.0.1:8191 | 127.0.0.1:5433 |
| producción | `/opt/campus` | 80 (443 si hay PEM) | 127.0.0.1:8091 | 127.0.0.1:5432 |

Develop y QA comparten la EC2 nonprod. Producción va en otra EC2. Solo el puerto web (y el 443 en producción) está abierto en el grupo de seguridad. Postgres y la API no se publican. Los runners viven en `/opt/actions-runner-<ambiente>` como servicio systemd.

`compose.host.yml` usa las imágenes de GHCR, quita `build`, ata `db` y `api` a `127.0.0.1` y monta los datos del disco.

Ejemplos de `host.env`, sin claves reales: `deploy/env/host.env.example.develop`, `host.env.example.qa` y `host.env.example.produccion`.

## Ambientes

| | develop | QA | producción |
| --- | --- | --- | --- |
| `APP_ENV` | `develop` | `qa` | `produccion` |
| Swagger y OpenAPI | encendidos | encendidos | 404 |
| Logs | debug, consola | info, JSON | info, JSON |
| Pool (abiertas / inactivas / vida) | 5 / 2 / 15 min | 8 / 3 / 20 min | 10 / 4 / 30 min |
| Carga si la base está vacía | ETL de `data/raw` y semilla ficticia | igual que develop | ETL, sin semilla ficticia |
| Cookie `Secure` | falsa | falsa | verdadera solo con el par PEM |
| URL de referencia | http://34.224.230.64:8088 | http://34.224.230.64:8188 | https://verde-pucp.duckdns.org |

`SWAGGER_ENABLED`, `CAMPUS_CORS_ORIGINS`, `LOG_LEVEL`, `LOG_FORMAT`, `SERVER_GIN_MODE` y las tres variables del pool pisan el default. Un `*` en CORS se descarta. Sin `APP_ENV` el proceso usa el pool histórico (10 / 4 / 30 min) y Swagger según el modo de Gin.

La IP elástica de producción es http://100.51.112.92. Por esa IP el certificado de `verde-pucp.duckdns.org` no coincide. Develop y QA no usan ese nombre. Si la instancia del laboratorio se apaga y la IP cambia, hay que actualizar `PUBLIC_URL` como dice [docs/CAMBIO-DE-CUENTA-LAB.md](../docs/CAMBIO-DE-CUENTA-LAB.md).

En la misma laptop los tres stacks locales conviven. Puertos publicados por `make up ENV=...` (se pueden pisar con `POSTGRES_PORT`, `API_PORT` y `WEB_PORT`; gana el entorno del llamador, luego `deploy/env/<ambiente>.env`, luego el example):

| | develop | QA | producción local |
| --- | --- | --- | --- |
| Web | 8088 | 8188 | 8288 |
| API | 8091 | 8191 | 8291 |
| Postgres | 5432 | 5433 | 5434 |
| Base | `campus_verde_develop` | `campus_verde_qa` | `campus_verde_produccion` |

```bash
make up ENV=develop
make smoke ENV=develop
make down ENV=develop
```

La primera vez se copia `deploy/env/<ambiente>.env.example` a `deploy/env/<ambiente>.env` (no se versiona). `make down` no borra volúmenes.

En producción, `APP_ENV=produccion` rechaza `pando-local`, `campus-lab` y claves de menos de 16 caracteres, tanto para las cuentas como para Postgres. El example local de producción trae una clave de demostración larga; no es la de la EC2.

## Rollback

Imágenes ya publicadas, sin reconstruir:

```bash
gh workflow run deploy.yml --repo Irico17/Areas-verdes-pucp --ref develop \
  -f ambiente=<develop|qa|produccion> \
  -f ref=<sha-actual> \
  -f rollback=<sha40-o-rc-*>
```

`rollback` acepta un SHA de 40 hexadecimales o un tag `rc-*`. En producción también pasa por aprobación. La versión anterior corre sobre el esquema ya migrado.

En local:

```bash
bash deploy/deploy.sh develop --local --rollback
bash deploy/deploy.sh develop --local --rollback <tag-de-la-api>
```

El estado queda en `deploy/state/` (`<amb>.prev-image` y `<amb>.prev-image-web`), que no se versiona. Si el smoke de `make up ENV=...` falla y había imagen previa, el script vuelve atrás solo.

Recuperar datos es otro paso, y lo decide una persona: el `pg_dump` de `/opt/campus/data/backups/` o el snapshot EBS tomado antes del despliegue. Restaurar encima de la base que está en servicio no es el camino; se restaura en una base vacía.

## Dónde viven las claves

| Sitio | Qué hay |
| --- | --- |
| `host.env` en la EC2 | Claves de Postgres y de las cuentas de ese ambiente. Modo 600. |
| Environments `develop`, `qa`, `produccion` | Variable `PUBLIC_URL`. Producción tiene revisor y solo admite la rama `develop`. El deploy no lee secretos de AWS. |
| Environment `aws-lab` | Credenciales temporales, `AWS_REGION`, `TF_VAR_db_password`, `TF_VAR_dev_password` y el PAT del runner. Solo para aprovisionar. |
| Portátil | `deploy/env/<ambiente>.env`, ignorado por git. Plantillas: `deploy/env/*.env.example`. |

No se suben claves al repositorio. `CAMPUS_DEV_PASSWORD` de producción tiene 16 caracteres o más y no es una clave de laboratorio. El nombre de un cubo de evidencias, si se usa, va en el `host.env` de ese ambiente (`EVIDENCIAS_BUCKET`). Vacío significa disco.

## HTTPS con DuckDNS

El nombre `verde-pucp.duckdns.org` es de producción. DuckDNS es un dominio de laboratorio, no el dominio definitivo de la universidad. Develop (`:8088`) y QA (`:8188`) siguen en HTTP.

El cliente es `lego` (imagen fijada en `deploy/tls/comun.sh`), con el proveedor DuckDNS. Nginx es el de `frontend`: si al arrancar ve `/etc/nginx/certs/fullchain.pem` y `privkey.pem`, usa la configuración TLS. No se añade otro proxy.

El token no entra al repositorio. Vive en `/opt/campus/duckdns.env` (modo 600; plantilla `deploy/tls/duckdns.env.example`) o, si se usa el workflow, en el secreto `DUCKDNS_TOKEN`. El workflow `.github/workflows/duckdns-cert.yml` es solo `workflow_dispatch`, corre en `[self-hosted, campus-prod]` y exige `confirmar=EMITIR`. `staging` vale 1 por defecto. No cambia la IP.

Ensayo en el host de producción, con Docker y el archivo ya creado:

```bash
sudo install -d -m 755 /opt/campus/bin
sudo install -m 755 deploy/tls/comun.sh deploy/tls/emitir-certificado.sh \
  deploy/tls/renovar-certificado.sh deploy/tls/actualizar-ip-duckdns.sh \
  /opt/campus/bin/
export DUCKDNS_ENV_FILE=/opt/campus/duckdns.env
export ACME_EMAIL=operador.campus@example.com
export ACME_STAGING=1
bash deploy/tls/emitir-certificado.sh
```

Cuando el ensayo deje un PEM y el journal no muestre el token, repetir con `ACME_STAGING=0`. DNS-01 no necesita que el nombre ya apunte a la EC2. La cuenta ACME queda en `/opt/campus/letsencrypt` (modo 700). Nginx solo monta `/opt/campus/certs`.

Para servir TLS la primera vez: dejar los dos PEM, poner en `host.env` `CAMPUS_COOKIE_SECURE=true`, `PUBLIC_URL=https://verde-pucp.duckdns.org` y el origen CORS de ese nombre, y volver a desplegar. `host-deploy.sh` incluye `compose.tls.yml` solo si los dos PEM existen. Sin ellos no publica el 443. HSTS usa un año: un navegador que ya vio el encabezado seguirá pidiendo HTTPS.

Renovación: `deploy/tls/renovar-certificado.sh` no llama a Let's Encrypt si faltan más de 30 días. Las unidades de ejemplo están en `deploy/tls/systemd/` (`campus-cert-renew.timer`, 03:00 y 15:00).

Cambiar la IP, al final y a mano, no desde un runner:

```bash
bash deploy/tls/actualizar-ip-duckdns.sh --dry-run --ip 100.51.112.92
bash deploy/tls/actualizar-ip-duckdns.sh --ip 100.51.112.92
```

El script rechaza `CI=true` y `GITHUB_ACTIONS=true`. `--auto` es para un timer en el propio host (`campus-duckdns-ip.timer`). No se habilita antes de decidir el cambio.

Comprobación en la EC2, con el certificado real:

```bash
curl -sS -o /dev/null -w '%{http_code}\n' http://127.0.0.1/health
curl -sS --resolve verde-pucp.duckdns.org:443:127.0.0.1 \
  -o /dev/null -w '%{http_code}\n' https://verde-pucp.duckdns.org/health
```

`/health` en el puerto 80 no se redirige: el timer de salud sigue recibiendo 200. El resto del 80 pasa a HTTPS. Sintaxis de nginx: `bash deploy/tls/test_nginx_tls.sh`.

## Backup, restauración y salud

Volcado diario, en la instancia:

```bash
DATABASE_URL='postgres://campus:CLAVE@127.0.0.1:5432/campus_verde' \
  bash scripts/backup-postgis.sh /opt/campus/backups/campus.dump
```

Es `pg_dump -Fc` e incluye PostGIS. La copia tiene que salir del disco de la instancia. Restaurar solo en una base vacía:

```bash
TARGET_DATABASE_URL='postgres://campus:CLAVE@127.0.0.1:5432/campus_restore' \
  bash scripts/restore-postgis.sh /opt/campus/backups/campus.dump
```

`scripts/probar-restauracion.sh` ensaya ese camino en un Postgres local, sin tocar la base de servicio.

`scripts/healthcheck.sh` pide `GET /health` sin seguir redirecciones. Si `status` no es `ok`, sale con código 1. En la EC2 el timer lo corre cada cinco minutos.

La API escribe una línea por petición (`request_id`, método, ruta, estado, latencia). No escribe el cuerpo, la query ni la cookie. En el servidor: `docker logs campus-<ambiente>-api`.

## Si algo falla

| Síntoma | Qué mirar |
| --- | --- |
| El job queda en cola | Settings → Actions → Runners. En la EC2, por SSM: `systemctl status 'actions.runner.*'`. |
| El registro del runner caducó | Token con `gh api -X POST repos/Irico17/Areas-verdes-pucp/actions/runners/registration-token`, parámetro SSM `/campus/runner-token/<ambiente>` y `scripts/runner-instalar.sh`. Borrar el parámetro después. |
| `esperar-ci` falla | El CI de ese SHA está en rojo o no existe. |
| `promocion` rechazada | Ese SHA no tiene un despliegue exitoso en el ambiente anterior. |
| `docker pull` con 401 o «not found» | El job `imagenes` no publicó ese SHA, o el paquete de GHCR no está enlazado al repositorio. |
| Smoke falla y hay rollback | `docker logs` de la API. Un reinicio en bucle suele ser la migración o la base. |
| Mapa sin áreas | La carga no corrió. El catastro publicado son 521 áreas y 534 sectores. |
| 502 en la web | La API no está arriba. |
| Snapshot EBS imposible | El rol de la instancia no puede crearlo. Hace falta el `pg_dump` antes de usar `DEPLOY_SIN_SNAPSHOT=1`. |
| Disco o memoria en producción | La EC2 de producción es pequeña (1 GB, con swap). `docker image prune` limpia imágenes viejas. No borra volúmenes ni `data/`. |

Caducar las credenciales de AWS no detiene un despliegue ya servido por el runner. Solo afecta a Terraform y al SSM desde fuera. Se renuevan con `scripts/actualizar-credenciales-lab.sh`.

Reglas que el pipeline no rompe: no hay `docker compose down -v`, ni `TRUNCATE` ni `DROP` de datos cargados en una migración nueva, ni force push.

## Camino desde la laptop

`scripts/deploy-learner-lab.sh` sigue en el repositorio. Exige `DEPLOY_AWS_CONFIRM=1`, las tres credenciales temporales y `TF_VAR_db_password` / `TF_VAR_dev_password`, y llama a `deploy/deploy.sh produccion --aws` (Terraform, ECR y SSM). No es el camino de `deploy.yml`. No se lanza sin una sesión del laboratorio abierta y sin querer aplicar cambios.

`scripts/set-aws-secrets.sh` sube esas tres credenciales a secretos del repositorio. Para una cuenta nueva el camino es el environment `aws-lab` y `scripts/actualizar-credenciales-lab.sh`, no ese script.

Parar o destruir la cuenta, sin borrar discos a ciegas, es el workflow `pausar-cuenta` y, si hace falta, `aprovisionar-cuenta` con `destroy`. El detalle está en [infra/README.md](../infra/README.md).

## Costo, en orden de magnitud

No es una cotización. El Learner Lab tiene un saldo de curso (unos 50 USD en total, no al mes) y apaga la instancia al cerrar la sesión. Una IP elástica asociada a una instancia detenida se cobra. Dos discos también.

Dejado encendido un mes en us-east-1, el default del laboratorio (t3.micro y dos discos) ronda los 12 USD. Una cuenta propia, sin NAT ni balanceador, ronda 25–45 USD (t3.small, discos, snapshots y un cubo pequeño de evidencias). La variante ECS de `infra/terraform/modules/ecs-fargate/` no se aplica: un balanceador solo ya se acerca a 16–20 USD.
