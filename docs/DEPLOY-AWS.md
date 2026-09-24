# Despliegue en AWS Academy Learner Lab

El destino por defecto es una sola EC2 pequeña en **us-east-1** con el `docker compose` del repo (PostGIS + API + nginx). No se crea ningún rol ni política: la instancia usa el `LabInstanceProfile` que ya trae el lab. La variante ECS + RDS + ALB no se aplica; está descrita en `infra/terraform/modules/ecs-fargate/README.md`.

sa-east-1 queda más cerca de Lima y suele costar más. El lab no lo habilita. Solo us-east-1 y us-west-2.

## Qué hace falta de usted

- Una sesión de Learner Lab abierta (Start Lab). El saldo del curso es de unos 50 USD en total, no al mes.
- Las tres credenciales temporales: access key, secret y **session token**. Caducan cuando termina la sesión (unas 4 horas).
- Docker y Terraform en la laptop, si despliega desde ahí.
- Opcional: un key pair ya creado en el lab, si va a entrar por SSH. El script también intenta SSM.
- No hace falta un dominio. La URL es el Elastic IP en HTTP (puerto 80). El 443 queda abierto para un certificado futuro; nginx escucha en 80.
- No hace falta crear IAM. Si el lab le pide un nombre, use `LabRole` y `LabInstanceProfile`.

## Pasos

1. En Vocareum, **Start Lab**. Espere a que el semáforo esté en verde.
2. **AWS Details** → **AWS CLI** → **Show**.
3. Copie el bloque. Tiene esta forma (los valores cambian cada sesión):

```bash
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_SESSION_TOKEN=...
export AWS_DEFAULT_REGION=us-east-1
```

4. Péguelo en la terminal, en el raíz del repo, y ejecute:

```bash
bash scripts/deploy-learner-lab.sh
```

El script hace `terraform init -reconfigure` y `apply` en `infra/terraform`, construye las dos imágenes, las sube a ECR y pide a la instancia, por SSM, que reinicie `campus.service`. Ese servicio hace `docker compose pull` y `up`. La API, al arrancar, migra y corre el ETL una vez (521 áreas y el resto del catastro) si `/data/.etl-done` no existe.

El estado de Terraform vive en S3, no en la laptop. El cubo es `campus-verde-tfstate-890991908027` (versionado, cifrado SSE-S3, sin acceso público), clave `learner-lab/terraform.tfstate`, región `us-east-1`. No hay tabla de lock: no haga dos `apply` a la vez. El bloque está en `infra/terraform/versions.tf`:

```hcl
backend "s3" {
  bucket  = "campus-verde-tfstate-890991908027"
  key     = "learner-lab/terraform.tfstate"
  region  = "us-east-1"
  encrypt = true
}
```

Si el cubo no existe en una cuenta nueva, créelo antes del primer `init` (el lab no deja crear IAM; este comando no lo hace):

```bash
aws s3api create-bucket --bucket campus-verde-tfstate-890991908027 --region us-east-1
aws s3api put-bucket-versioning --bucket campus-verde-tfstate-890991908027 \
  --versioning-configuration Status=Enabled
```

5. Abra la URL que imprime el script (`http://<elastic-ip>`). Entre con `coordinacion` / `pando-local`, o `norte` para el capataz.

La primera vez la instancia puede tardar unos minutos: cloud-init instala Docker y el binario de Compose v2 (Amazon Linux 2023 no trae el paquete `docker-compose-plugin`), formatea el volumen de datos y arranca el compose. Si el pull ocurre antes de que las imágenes estén en ECR, `systemctl restart campus.service` lo repite.

## Cuando se acaba la sesión

- Las credenciales dejan de servir. Un `apply` a medias hay que repetirlo en la sesión siguiente, con claves nuevas.
- EC2 queda **detenida**. El Elastic IP asociado a una instancia detenida sí se cobra. El disco (raíz de 20 GB y otro de 20 GB para Postgres y archivos) también, aunque la máquina esté apagada.
- Al volver: Start Lab, copiar credenciales nuevas, y o bien encender la instancia en la consola, o volver a correr el script. `user_data` no se repite; Docker arranca solo si la instancia vuelve a encender y el unit `campus.service` está habilitado.

## Parar, encender y destruir desde la laptop

En una sesión nueva del lab, copie credenciales frescas (no sirven las de la sesión anterior) y deje `AWS_DEFAULT_REGION=us-east-1`. El estado ya está en el cubo de arriba.

```bash
cd infra/terraform
terraform init -input=false -reconfigure
```

Detener (la instancia para; el Elastic IP y los discos siguen cobrando):

```bash
aws ec2 stop-instances --instance-ids "$(terraform output -raw instance_id)"
```

Encender en la sesión siguiente:

```bash
aws ec2 start-instances --instance-ids "$(terraform output -raw instance_id)"
```

Destruir el stack (instancia, IP, volúmenes y ECR). El cubo de estado no forma parte del root; bórrelo aparte solo si ya no va a redesplegar:

```bash
terraform destroy -auto-approve
```

## Cómo no quemar el saldo

- Al terminar el día, desde `infra/terraform`, con las credenciales aún vivas:

```bash
terraform destroy -auto-approve
```

Eso suelta la instancia, el IP, los volúmenes y los repositorios ECR (`force_delete`).

- Si solo va a pausar unas horas dentro de la misma sesión, detenga la instancia y sepa que el IP y el disco siguen contando.
- No encienda `enable_ecs`. Un ALB solo ya se acerca a 16–20 USD al mes.

## Costo aproximado del default

Cifras de lista us-east-1, **aproximadas**, si se dejara todo encendido un mes entero. El lab no llega a eso si se destruye al cerrar.

| Pieza | Aprox. al mes |
| --- | --- |
| t3.micro encendida | 8 USD |
| t3.small, si la cambia | 15 USD |
| Dos volúmenes gp3 de 20 GB | 3 USD |
| Elastic IP con instancia encendida | 0 |
| Elastic IP con instancia detenida | 4 USD |
| ECR, unas imágenes pequeñas | < 1 USD |
| **Total holgado del default** | **cerca de 12 USD** |

Cabe en los 50 USD del curso con margen, siempre que no se deje el IP huérfano ni se prenda la variante ECS.

## Terraform a mano

```bash
cd infra/terraform
terraform init
terraform fmt -check -recursive
terraform validate
terraform plan
terraform apply
```

Variables útiles: `instance_type` (`t3.micro` o `t3.small`), `ssh_cidr` (estreche a `x.x.x.x/32`), `ssh_key_name`, `create_evidence_bucket` (el cubo lo lee la API con `LabRole`, sin política nueva).

## CI

`.github/workflows/ci.yml` corre en cada push: `gofmt`, `go test`, lint y build del web, y las dos imágenes. El job de AWS es solo `workflow_dispatch`. Si faltan `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` o `AWS_SESSION_TOKEN`, no despliega y termina en verde.

## Permisos mínimos

En este lab no se pueden crear. El `LabRole` que ya existe tiene que poder, para el script: EC2, VPC (security group y EIP), ECR, SSM `SendCommand`, y S3 si enciende el cubo de evidencias. La instancia asume ese mismo rol vía `LabInstanceProfile` para `ecr:GetAuthorizationToken` y, si aplica, `s3:PutObject` / `s3:GetObject` del cubo de evidencias.
