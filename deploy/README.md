# Despliegue por ambiente

`deploy/deploy.sh <develop|qa|produccion> --local` levanta PostGIS, la API y nginx de ese ambiente. Los tres pueden estar arriba a la vez. La matriz de puertos, secretos y el flujo hasta producción está en [`docs/AMBIENTES.md`](../docs/AMBIENTES.md).

```bash
make up ENV=develop
make smoke ENV=develop
make down ENV=develop
```

`--aws` habla con el Learner Lab (Terraform, ECR, SSM). Hace falta `DEPLOY_AWS_CONFIRM=1` y las credenciales temporales. `scripts/deploy-learner-lab.sh` lo exporta y despliega `produccion`. No lo corra sin una sesión del lab abierta y sin querer aplicar cambios en AWS.

La semilla de develop y qa está en [`seed/README.md`](seed/README.md).
