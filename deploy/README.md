# Despliegue por ambiente

`deploy/deploy.sh <develop|qa|produccion> --local` levanta PostGIS, la API y nginx de ese ambiente. Los tres pueden estar arriba a la vez. La matriz de puertos, secretos y el flujo hasta producción está en [`docs/AMBIENTES.md`](../docs/AMBIENTES.md).

```bash
make up ENV=develop
make smoke ENV=develop
make down ENV=develop
```

### Sobrescritura de puertos desde el entorno

Si en su máquina ya corren PostgreSQL en el puerto `5432` o la API vieja en `8091`, puede cambiar los puertos publicados directamente desde variables de entorno del llamador sin alterar los archivos de configuración:
- `POSTGRES_PORT` (por defecto: 5432 develop, 5433 qa, 5434 produccion)
- `API_PORT` (por defecto: 8091 develop, 8191 qa, 8291 produccion)
- `WEB_PORT` (por defecto: 8088 develop, 8188 qa, 8288 produccion)

**Precedencia:** entorno del llamador > archivo `.env` del ambiente > `.env.example`.

Ejemplo:
```bash
POSTGRES_PORT=5442 API_PORT=8093 WEB_PORT=8089 make up ENV=develop
POSTGRES_PORT=5442 API_PORT=8093 WEB_PORT=8089 make smoke ENV=develop
POSTGRES_PORT=5442 API_PORT=8093 WEB_PORT=8089 make down ENV=develop
```

`--aws` habla con el Learner Lab (Terraform, ECR, SSM). Hace falta `DEPLOY_AWS_CONFIRM=1` y las credenciales temporales. `scripts/deploy-learner-lab.sh` lo exporta y despliega `produccion`. No lo corra sin una sesión del lab abierta y sin querer aplicar cambios en AWS.

La semilla de develop y qa está en [`seed/README.md`](seed/README.md).

### Claves y seguridad en producción

En producción (`APP_ENV=produccion`), se rechazan claves de laboratorio (`pando-local`, `campus-lab`) y claves con menos de 16 caracteres. El rechazo ocurre de forma temprana en `backend/docker-entrypoint.sh` (antes de migrar) y en la carga de configuración de la API y de `migrate`. Para pruebas locales de producción, `deploy/env/produccion.env.example` provee la clave de demostración `produccion-local-demo-2026`. En AWS/Terraform se inyectan mediante `TF_VAR_db_password` y `TF_VAR_dev_password`.

## Conteos y verificación de negocio

En producción, el despliegue compara los conteos de filas antes y después de aplicar la imagen nueva (`deploy/comparar_conteos.py`). La verificación cubre **únicamente tablas de datos de negocio** (catastro, labores, cuadrillas, catálogos) y cualquier discrepancia aborta el despliegue.

Para evitar falsos positivos por escrituras legítimas durante el despliegue, se excluyen explícitamente las tablas listadas en [`deploy/conteos.excluir`](conteos.excluir) (fuente única de verdad):
- `schema_migrations`: el entrypoint aplica migraciones nuevas pendientes al iniciar.
- `sesiones`: el smoke test post-despliegue ejecuta autenticación (`POST /api/v1/sesion`).

Las tablas de auditoría (`cambios`) no sufren escrituras en el smoke ni en el login. Tablas nuevas agregadas por migraciones se reportan sin provocar error.

## Rollback local

`deploy/deploy.sh <ambiente> --local --rollback [TAG]` revierte tanto la imagen de la API como la de la web al estado previo guardado en `deploy/state/` (`<amb>.prev-image` y `<amb>.prev-image-web`). Si se pasa un `TAG` explícito, este aplica a la API y la imagen web se recupera del estado guardado (fallando si no existe).

Durante `deploy.sh <amb> --local` (o `make up ENV=<amb>`), si el smoke test post-despliegue falla, el script ejecuta automáticamente el rollback a las imágenes previas de API y web.

**Las migraciones son solo hacia adelante:** el rollback local no revierte la base de datos ni ejecuta operaciones destructivas (no hay migraciones «down»). La versión restaurada corre sobre el esquema migrado, compatible gracias a que las migraciones son aditivas.

## Despliegue en el propio host (Runner self-hosted)

Para desplegar directamente dentro de las instancias EC2 sin credenciales de AWS desde GitHub, se utiliza [`deploy/host-deploy.sh`](host-deploy.sh):

```bash
deploy/host-deploy.sh <develop|qa|produccion> --sha <40hex> [--rollback <sha40|rc-*>] [--dry-run]
```

- **Layout:** develop en `/opt/campus/develop`, qa en `/opt/campus/qa`, producción en `/opt/campus`.
- **Configuración:** `$CAMPUS_HOME/host.env` (permisos 600, no versionado). Ejemplos en [`deploy/env/host.env.example.<ambiente>`](env/).
- **Override de Compose:** [`deploy/compose.host.yml`](compose.host.yml) utiliza imágenes inmutables de GHCR, elimina `build`, aísla `db` y `api` en `127.0.0.1` y utiliza bind mounts de datos en disco.
- **Producción:** ejecuta conteos «antes», snapshot EBS del volumen `campus-verde-data` vía IMDSv2, backup lógico `pg_dump`, smoke test y comparación de conteos de tablas de negocio «después».
- **Rollback:** automático ante fallos de smoke test o manual con `--rollback <tag>`.

La guía completa de arquitectura, layout y operación con runners self-hosted se encuentra en [`docs/DEPLOY-RUNNER.md`](../docs/DEPLOY-RUNNER.md).


