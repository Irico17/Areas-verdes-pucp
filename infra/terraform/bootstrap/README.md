# Estado remoto

El root usa el backend S3 declarado en `versions.tf`:

- Cubo: `campus-verde-tfstate-890991908027`
- Clave: `learner-lab/terraform.tfstate`
- Región: `us-east-1`
- Cifrado: SSE-S3 (`encrypt = true`)
- Versionado del cubo: activado a mano, fuera de este módulo

No hay tabla DynamoDB de lock y este directorio no crea IAM. Un `apply` a la vez. El cubo se crea con la CLI antes del primer `terraform init`, como dice `docs/DEPLOY-AWS.md`.
