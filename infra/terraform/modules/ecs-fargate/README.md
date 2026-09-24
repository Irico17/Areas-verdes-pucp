# Variante ECS (no es el default)

No la instancia el root de `infra/terraform`. Learner Lab no puede crear roles y el presupuesto de 50 USD no alcanza para ALB + RDS + Fargate encendidos.

Pensada para una cuenta AWS normal, más adelante:

- ECR que ya exista
- ECS Fargate para `api` y `web`, `task_role` y `execution_role` = el ARN de un rol que usted ya tenga (en el lab sería `LabRole`, y aun así el lab suele bloquear ECS)
- Un ALB o CloudFront delante, mismo origen para la cookie
- RDS PostgreSQL pequeño con PostGIS (`CREATE EXTENSION postgis`)
- Sin recursos `aws_iam_*`

Para usarla hay que copiar el módulo a un root propio y pasar los ARN. `enable_ecs` en el root del lab debe seguir en `false`: un `check` de Terraform lo rechaza.
