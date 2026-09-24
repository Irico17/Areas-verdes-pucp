# Estado remoto (opcional, no requerido)

El lab usa estado local en `infra/terraform/terraform.tfstate`. No haga falta S3 ni DynamoDB.

En una cuenta normal, un bootstrap aparte puede crear un cubo y una tabla de lock. No forma parte de este apply y no crea IAM: quien aplica usa el rol que ya tiene.
