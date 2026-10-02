# Despliegue por ambiente

`deploy/deploy.sh <develop|qa|produccion> --local` levanta PostGIS, la API y nginx de ese ambiente. Los tres pueden estar arriba a la vez. La matriz de puertos, secretos y el flujo hasta producción está en [`docs/AMBIENTES.md`](../docs/AMBIENTES.md).

```bash
make up ENV=develop
make smoke ENV=develop
make down ENV=develop
```

`--aws` habla con el Learner Lab (Terraform, ECR, SSM). Hace falta `DEPLOY_AWS_CONFIRM=1` y las credenciales temporales. `scripts/deploy-learner-lab.sh` lo exporta y despliega `produccion`. No lo corra sin una sesión del lab abierta y sin querer aplicar cambios en AWS.

La semilla de develop y qa está en [`seed/README.md`](seed/README.md).

## Conteos y verificación de negocio

En producción, el despliegue compara los conteos de filas antes y después de aplicar la imagen nueva (`deploy/comparar_conteos.py`). La verificación cubre **únicamente tablas de datos de negocio** (catastro, labores, cuadrillas, catálogos) y cualquier discrepancia aborta el despliegue.

Para evitar falsos positivos por escrituras legítimas durante el despliegue, se excluyen explícitamente las tablas listadas en [`deploy/conteos.excluir`](conteos.excluir) (fuente única de verdad):
- `schema_migrations`: el entrypoint aplica migraciones nuevas pendientes al iniciar.
- `sesiones`: el smoke test post-despliegue ejecuta autenticación (`POST /api/v1/sesion`).

Las tablas de auditoría (`cambios`) no sufren escrituras en el smoke ni en el login. Tablas nuevas agregadas por migraciones se reportan sin provocar error.

