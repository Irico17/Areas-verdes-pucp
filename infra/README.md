# Aprovisionamiento

Terraform crea las máquinas del Learner Lab. Ansible les instala Docker, el layout de `/opt/campus` y el runner. El despliegue de la aplicación no pasa por aquí: lo hace el runner, y está descrito en [deploy/README.md](../deploy/README.md).

La guía para cambiar de cuenta, con la lista de comprobación, es [docs/CAMBIO-DE-CUENTA-LAB.md](../docs/CAMBIO-DE-CUENTA-LAB.md). Las protecciones del repositorio son [docs/PROTECCIONES-REPO.md](../docs/PROTECCIONES-REPO.md).

El laboratorio no deja crear roles IAM ni un proveedor OIDC. La instancia usa el `LabInstanceProfile` que la cuenta ya trae. El puerto 22 queda cerrado. La región es `us-east-1` o `us-west-2`.

## Qué hay en este directorio

| Ruta | Qué aprovisiona |
| --- | --- |
| `terraform/` | Producción: una EC2, IP elástica, disco de datos, grupo de seguridad. Estado en `learner-lab/terraform.tfstate`. |
| `terraform/nonprod/` | Una EC2 compartida para develop (web 8088) y QA (web 8188). Estado en `learner-lab-nonprod/terraform.tfstate`. |
| `terraform/bootstrap/` | Notas del cubo de estado. Lo crea `scripts/bootstrap-estado-terraform.sh`. |
| `terraform/modules/ecs-fargate/` | Variante que no se aplica. No encender `enable_ecs`. |
| `terraform/environments/*.tfvars.example` | Tamaños de ejemplo, sin claves. |
| `ansible/` | Playbook `site.yml`: Docker, usuario `runner`, `/opt/campus` y el servicio del runner. |

El cubo de estado se llama `campus-verde-tfstate-<id de la cuenta>` (versionado, SSE-S3, sin acceso público). El candado es el de S3 (`use_lockfile`, Terraform 1.10 o posterior). No hay tabla DynamoDB. El nombre no está escrito en `versions.tf`: Terraform lo recibe con `-backend-config`.

Ansible copia módulos por el cubo `campus-verde-ssm-<cuenta>`, sin versionado y con expiración al día. La conexión es SSM (`amazon.aws.aws_ssm`). El token del runner no viaja en el playbook: el workflow lo deja en `/campus/runner-token/<ambiente>` y `scripts/runner-instalar.sh` lo lee. Al terminar, el parámetro se borra.

Tamaños por defecto, para cuidar el crédito: nonprod `t3.small` y producción `t3.micro`, discos de 20 GB, tope 40 GB. `user_data_replace_on_change` está en falso: un apply que solo cambia la imagen no recrea la EC2.

## Workflows

| Workflow | Qué hace |
| --- | --- |
| `.github/workflows/aprovisionar-cuenta.yml` | `plan`, `apply` o `destroy`. Environment `aws-lab`. |
| `.github/workflows/pausar-cuenta.yml` | Enciende o detiene instancias por etiqueta, sin borrar discos. |

Los dos comparten la concurrencia `cuenta-lab-aws` para no apagar una máquina a mitad de un apply.

Credenciales: no se pegan en los inputs. En el laboratorio, Start Lab → AWS Details → AWS CLI → Show, y en el repositorio:

```bash
pbpaste | bash scripts/actualizar-credenciales-lab.sh
```

Eso escribe los tres secretos y `AWS_REGION` en el environment `aws-lab`. Caducan con la sesión (unas cuatro horas). El workflow lo comprueba con `aws sts get-caller-identity`.

También hacen falta, en `aws-lab` y nunca en el repositorio:

- `TF_VAR_db_password` y `TF_VAR_dev_password`: 16 caracteres o más, distintas de `pando-local` y de `campus-lab`. Terraform las exige y no las escribe en el user data. La clave que usa producción va en `/opt/campus/host.env`, a mano.
- `RUNNER_REG_TOKEN_PAT`: PAT de grano fino, solo este repositorio, Administration lectura y escritura, para el token de registro del runner y para actualizar `PUBLIC_URL`.

Disparo: Actions → aprovisionar-cuenta. La primera vez, `accion=plan` y `ambientes=nonprod`. Si el plan dice `forces replacement` y esa máquina tiene datos, no se aplica: se suben los discos o el tipo a los de `terraform/nonprod/cuenta-actual.tfvars.example`.

`apply` exige `confirmar=APLICAR`. `destroy` exige `confirmar=DESTRUIR`. Un plan con reemplazo exige además `acepto_reemplazo=SI`.

Después del apply, Ansible instala Docker y Compose, el usuario `runner`, los compose de referencia y el servicio del runner. No instala nginx en el host (lo trae el contenedor web) ni las unidades de TLS. Develop y QA reciben el `host.env` de ejemplo solo si el archivo no existía. El de producción lo crea el operador.

Pausa: `pausar-cuenta`. `stop` exige `DETENER` y `start` exige `ENCENDER`. Busca `Project=campus-verde`, `CampusGestion=cuenta-lab` y `Stack=nonprod` o `Stack=learner-lab`.

## Comprobar sin AWS

```bash
cd infra/terraform
terraform fmt -check -recursive
terraform init -backend=false
terraform validate
```

```bash
cd infra/ansible
ansible-playbook --syntax-check -i inventory/hosts.example.yml site.yml
bash probar-idempotencia.sh
```

`inventory/hosts.yml` lo genera el workflow y no se versiona. El ejemplo sí.

Estos comandos no despliegan la aplicación ni aprueban un deploy. El push a `develop`, o `gh workflow run deploy.yml`, viene después, con los runners en verde.
