# Estado remoto

El root ya no trae el nombre del cubo en `versions.tf`. Cada cuenta de Learner Lab usa el suyo:

- Cubo: `campus-verde-tfstate-<id de la cuenta>`
- Clave de producción: `learner-lab/terraform.tfstate`
- Clave nonprod: `learner-lab-nonprod/terraform.tfstate`
- Región: la de `AWS_REGION` (`us-east-1` o `us-west-2`)
- Cifrado: SSE-S3 (`encrypt = true`)
- Candado: `use_lockfile` de Terraform 1.10 o posterior. No hay tabla DynamoDB ni rol IAM nuevo.

Lo crea `scripts/bootstrap-estado-terraform.sh` (también lo hace el workflow `aprovisionar-cuenta`). El procedimiento completo está en [`docs/CAMBIO-DE-CUENTA-LAB.md`](../../../docs/CAMBIO-DE-CUENTA-LAB.md).
