# Cambiar de cuenta de Learner Lab

Esta guía es para quien tiene que mover VerdePUCP a otra cuenta de AWS Academy Learner Lab (el crédito de unos 50 USD se acabó, o las credenciales de la sesión ya no sirven y la cuenta no vuelve). No hace falta conocer Terraform ni Ansible para seguirla. No hay contraseñas en este documento.

El crédito del lab es del curso, no un presupuesto mensual. Dos máquinas encendidas todo el mes se lo comen. Al terminar la clase, detenga o destruya.

## Qué queda igual y qué cambia

El despliegue de la aplicación (`.github/workflows/deploy.yml`) no usa la cuenta de AWS. Corre en los runners del propio servidor, con las etiquetas `campus-develop`, `campus-qa` y `campus-prod`. Esas etiquetas no cambian al cambiar de cuenta. La URL pública sale de la variable `PUBLIC_URL` de cada environment de GitHub (`develop`, `qa`, `produccion`).

Lo que sí es de una cuenta concreta, y por eso no está escrito en el código:

| Dato | Dónde vive ahora |
| --- | --- |
| Credenciales de la sesión (access key, secret, session token) | Secretos del environment `aws-lab` |
| Región | Variable `AWS_REGION` del environment `aws-lab` (`us-east-1` o `us-west-2`) |
| Cubo del estado de Terraform | `campus-verde-tfstate-<id de la cuenta>`, creado en esa misma cuenta |
| Id de instancia, IP, nombre del cubo | Salidas de Terraform y etiquetas `Project=campus-verde`, `CampusGestion=cuenta-lab` |
| URL que usa el deploy | Variable `PUBLIC_URL` de `develop`, `qa` y `produccion` |
| Etiqueta del runner, si alguna vez hiciera falta otra | Variables de repositorio `RUNNER_LABEL_DEVELOP`, `RUNNER_LABEL_QA`, `RUNNER_LABEL_PRODUCCION`. Vacías, siguen siendo `campus-develop`, `campus-qa` y `campus-prod` |

Auditoría del workflow de despliegue: no tenía id de instancia, ARN, IP, región ni nombre de cubo fijos. El registro de imágenes es `ghcr.io/<dueño del repo>/...`, tomado de `github.repository`, no de la cuenta de AWS. El snapshot de producción busca el volumen por la etiqueta `Name=campus-verde-data` y por el id que la propia instancia lee de los metadatos.

## Por qué las credenciales van en un environment

Pegar la access key en los inputs de un workflow la deja visible en la pantalla del run, en la API y en el resumen, antes de que ningún paso pueda enmascararla. El camino normal es otro:

1. En el Learner Lab: **Start Lab**. Espere el semáforo verde.
2. **AWS Details → AWS CLI → Show**. Copie el bloque (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, región).
3. En su máquina, dentro del repo y con `gh auth login` ya hecho:

```bash
pbpaste | bash scripts/actualizar-credenciales-lab.sh
```

En Windows, PowerShell:

```powershell
Get-Clipboard -Raw | .\scripts\actualizar-credenciales-lab.ps1
```

El script lee el bloque por stdin o, si la terminal está vacía, por el portapapeles. Llama a `gh secret set --env aws-lab` y a `gh variable set AWS_REGION --env aws-lab`. No imprime los valores y no los guarda en el repo.

Esas credenciales caducan cuando termina la sesión del lab (unas cuatro horas). El workflow lo comprueba con `aws sts get-caller-identity` y se detiene si caducaron.

Hay un camino secundario, en el mismo workflow, con el input `usar_credenciales_de_inputs`. El primer paso hace `::add-mask::` para que el log no las muestre. Aun así quedan en el historial del run. Úselo solo si no puede correr el script, y borre la sesión del lab en cuanto termine.

## Lista de comprobación

1. **Environment `aws-lab`.** En GitHub: Settings → Environments → New environment → `aws-lab`. No le ponga revisores: si no, el aprovisionamiento se queda esperando. No es el environment de producción.
2. **Credenciales.** El script de arriba. Compruebe que existen los secretos `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` y `AWS_SESSION_TOKEN`, y la variable `AWS_REGION`.
3. **PAT del runner.** Fine-grained, solo el repositorio `Irico17/Areas-verdes-pucp`, caducidad de 90 días o menos. Permiso **Administration: Read and write**. Con eso alcanza para:
   - `POST /repos/{owner}/{repo}/actions/runners/registration-token`
   - `PUT` de la variable `PUBLIC_URL` en los environments `develop`, `qa` y `produccion`
   No le dé Contents, Secrets, ni permisos de otras repos ni de la cuenta. Guárdelo como secreto `RUNNER_REG_TOKEN_PAT` del environment `aws-lab`. Un PAT clásico con scope `repo` también sirve y es más amplio: prefiera el de grano fino.
4. **Claves de la aplicación para el root de producción.** Secretos `TF_VAR_db_password` y `TF_VAR_dev_password` en `aws-lab`, 16 caracteres o más, distintas de `pando-local` y de `campus-lab`. Terraform las exige y no las escribe en la instancia ni en el user data. La clave que de verdad usa producción va en `/opt/campus/host.env`, a mano.
5. **Environments de despliegue.** `develop`, `qa` y `produccion` tienen que existir. `produccion` mantiene revisor obligatorio y la rama `develop`. Este procedimiento no aprueba ni lanza un despliegue.
6. **Plan.** Actions → aprovisionar-cuenta → Run workflow. `accion=plan`, `ambientes=nonprod` la primera vez. Lea el artefacto de texto. Si dice `forces replacement` y esa máquina tiene datos, no aplique: suba `disco_raiz_gb` / `disco_datos_gb` o el tipo de instancia a los de `infra/terraform/nonprod/cuenta-actual.tfvars.example`.
7. **Apply.** El mismo workflow con `accion=apply` y `confirmar=APLICAR`. `destroy` exige `confirmar=DESTRUIR`. Un plan con reemplazo exige además `acepto_reemplazo=SI`.
8. **Runners.** Con el PAT, el apply registra `campus-develop` y `campus-qa` en la máquina nonprod, y `campus-prod` en la de producción. En Settings → Actions → Runners deben verse en verde.
9. **host.env de producción.** Ansible no escribe esa clave. Entre por SSM y cree `/opt/campus/host.env` (modo 600) a partir de `/opt/campus/host.env.example`. Develop y qa reciben el ejemplo del repo solo si el archivo no existía.
10. **Deploy.** Sin editar archivos: push a `develop`, o `gh workflow run deploy.yml -f ambiente=develop`. Qa y producción siguen con su promoción y, producción, con la aprobación de siempre.

## Qué hace el aprovisionamiento

Workflow: `.github/workflows/aprovisionar-cuenta.yml`.

- Permiso del `GITHUB_TOKEN`: solo `contents: read`. El PAT del runner es un secreto aparte y solo se usa para el token de registro y para `PUBLIC_URL`.
- Concurrencia `cuenta-lab-aws`, sin cancelar un run en curso. El workflow de pausa usa el mismo grupo para no apagar una máquina a mitad de un apply.
- Terraform **1.11.4**, comprobado por SHA-256. `fmt -check`, `init`, `validate`, `plan` guardado y `apply` de ese plan. El binario del plan no se sube. El texto sí, cinco días, después de comprobar que no contiene los secretos del entorno.
- Recursos que el lab ya permite: EC2, security group, EBS, EIP, S3, ECR opcional. El perfil de instancia es el que ya existe, `LabInstanceProfile`. No se crea un rol IAM. El puerto 22 queda cerrado.
- Tamaños por defecto, para cuidar el crédito: nonprod `t3.small` con discos de 20 GB; producción `t3.micro` con discos de 20 GB. Tope 40 GB.
- Después del apply, Ansible (`infra/ansible/`) instala Docker y Compose, el usuario `runner`, `/opt/campus`, los compose de referencia y el servicio del runner. Nginx no se instala en el host: lo trae el contenedor web. No crea unidades de TLS ni de DuckDNS.

Pausa: `.github/workflows/pausar-cuenta.yml`. `stop` exige `DETENER`, `start` exige `ENCENDER`. Busca instancias con `Project=campus-verde`, `CampusGestion=cuenta-lab` y `Stack=nonprod` o `Stack=learner-lab`.

## Dónde está el estado de Terraform

Una cuenta nueva no tiene estado y no debe heredar el de la cuenta anterior: esos id de instancia no existen en la cuenta nueva, y un apply con el estado viejo intentaría gestionar recursos ajenos.

`scripts/bootstrap-estado-terraform.sh` hace, en la cuenta de las credenciales actuales:

- `campus-verde-tfstate-<id>`: versionado, cifrado SSE-S3, sin acceso público. Ahí van dos claves: `learner-lab/terraform.tfstate` (producción) y `learner-lab-nonprod/terraform.tfstate`.
- `campus-verde-ssm-<id>`: sin versionado, los objetos expiran al día. Solo lo usa Ansible para copiar módulos por SSM. Sin versionado, un archivo temporal no se queda en el historial del cubo.

El bloque `backend` del código no lleva nombre de cubo. El script escribe `tmp/backend.hcl` y Terraform lo recibe con `-backend-config`. El candado es el de S3 (`use_lockfile`, Terraform ≥ 1.10), no una tabla DynamoDB. Aun así, no lance dos apply a la vez: el workflow ya lo impide.

**Por qué en la misma cuenta y no en un artefacto de GitHub.** El artefacto vive 5 días, lo puede bajar quien tenga acceso al repo y no sirve para un `terraform plan` desde una laptop la semana siguiente. El cubo S3 versionado sí. El precio es que, si la cuenta muere, el estado muere con ella.

### Si la cuenta se agota

El estado y las máquinas quedan inaccesibles juntos. No copie ese estado a la cuenta nueva.

1. Si todavía puede entrar a la cuenta vieja, lance `aprovisionar-cuenta` con `accion=destroy` y `confirmar=DESTRUIR` para soltar la IP y los discos. El cubo de estado no se borra solo; puede vaciarlo a mano cuando ya no vaya a volver.
2. En la cuenta nueva: secretos nuevos, `accion=plan` y luego `apply`. Empieza de cero. Los datos de Postgres no viajan: si los necesita, saque un `pg_dump` mientras la cuenta vieja aún responde y restáurelo a mano. Este flujo no lo hace.
3. Vuelva a registrar los runners (el apply lo hace si el PAT sigue vigente) y deje que el workflow escriba `PUBLIC_URL`.

### Si el estado se perdió pero las máquinas de ESA cuenta siguen vivas

No haga apply a ciegas: Terraform crearía otra instancia y chocaría con los nombres. Genere los import y léalos antes de ejecutarlos:

```bash
bash scripts/recuperar-estado-terraform.sh --stack nonprod
bash scripts/recuperar-estado-terraform.sh --stack prod
```

El script solo imprime. Después de importar, un `plan` debería salir vacío o casi. La cuenta nueva, vacía, no se toca con estos import: los id son de la cuenta donde corrió el script.

## Costo aproximado

Precios de lista de us-east-1, redondeados, si se deja todo encendido un mes. El lab no llega a un mes si se destruye al cerrar. El crédito de ~50 USD es el total de la cuenta.

| Pieza | Al mes, aprox. |
| --- | --- |
| t3.micro (producción) | 8 USD |
| t3.small (develop+qa) | 15 USD |
| t3.medium, si lo sube | 30 USD |
| Disco gp3, 20 GB | 1,6 USD cada uno |
| IP elástica con la instancia encendida | 0 |
| IP elástica con la instancia detenida | unos 3,6 USD |
| Cubo de estado y ECR | menos de 1 USD |
| **nonprod small + prod micro + cuatro discos de 20 GB, siempre encendidos** | **cerca de 30 USD** |

Detener las instancias (`pausar-cuenta`) ahorra el cómputo y deja de cobrar la IP solo si también la suelta. Con la IP asociada a una instancia detenida, la IP se cobra. `destroy` suelta instancia, IP, discos y ECR. No borra el cubo de estado.

## Ansible, SSM y el token del runner

El acceso al host es Session Manager, con `LabInstanceProfile`. Hace falta el plugin `session-manager-plugin` 1.2.707.0 en el runner de GitHub (el workflow lo instala y comprueba el SHA-256) y el agente SSM en la instancia (Amazon Linux 2023 lo trae; Ansible lo deja habilitado).

El usuario de sesión SSM necesita `sudo` sin contraseña para que Ansible pase a root. En el lab, `ssm-user` lo tiene. Si el playbook se queja de sudo, la sesión del lab no dejó ese sudo: entre por la consola de Session Manager y ejecute `scripts/runner-instalar.sh` a mano.

Permisos que ya trae el rol del lab y que este flujo usa, sin crear políticas: EC2, VPC, S3 (los dos cubos), SSM `GetParameter` en la instancia y `PutParameter` desde el workflow, KMS sobre la clave por defecto de SSM (`alias/aws/ssm`) para el SecureString.

`ansible-playbook --check` es el input `ansible_check` junto con `accion=plan`.

## Si algo sale mal

| Síntoma | Qué hacer |
| --- | --- |
| `ExpiredToken` o «caducaron» | Start Lab, copie el bloque nuevo y vuelva a correr `actualizar-credenciales-lab.sh`. No reutilice el bloque de ayer. |
| El lab rechaza crear un rol o una política | Esperado. El código usa `LabInstanceProfile` y no declara `aws_iam_role`. |
| `BucketAlreadyExists` | El nombre lleva el id de la cuenta. Si choca, otra cuenta lo creó o la región no coincide. No reutilice el cubo de la cuenta anterior. |
| El plan dice `forces replacement` | Hay datos en ese disco. Repita el plan con el tamaño actual o acepte el reemplazo con `acepto_reemplazo=SI`, sabiendo que borra el volumen. |
| El apply termina y el runner sigue offline | Falta `RUNNER_REG_TOKEN_PAT`, el PAT no tiene Administration, o SSM no llegó a Online (espere unos minutos y repita el apply: Terraform no recrea la instancia). Mire `systemctl status 'actions.runner.*'` por SSM. |
| El deploy queda en Queued | El runner no está en línea o la etiqueta no coincide. Settings → Actions → Runners. |
| `PUBLIC_URL` sigue apuntando a la IP vieja | El apply no pudo llamar a `gh variable set` (PAT). Póngala a mano: `gh variable set PUBLIC_URL --env develop --body 'http://<ip>:8088'` y lo mismo para qa (`:8188`) y produccion (`http://<ip>`). |
| Límites del lab (vCPU, EIP, S3) | `destroy` de lo que no use, o pida la cuenta de otro estudiante y empiece por el plan en esa cuenta. No importe el estado viejo. |
| Dos apply a la vez | El segundo espera el candado o falla. Espere a que termine el primero. |

## Comandos útiles, sin secretos en la línea

```bash
bash scripts/aprovisionar-cuenta.sh   # exige ACCION, CONFIRMAR y las variables de entorno; en local, mejor el workflow
bash scripts/pausar-cuenta.sh         # ACCION=stop CONFIRMAR=DETENER
bash infra/ansible/probar-idempotencia.sh
```

Validar sin cuenta:

```bash
terraform -chdir=infra/terraform fmt -check -recursive
terraform -chdir=infra/terraform init -backend=false
terraform -chdir=infra/terraform validate
terraform -chdir=infra/terraform/nonprod init -backend=false
terraform -chdir=infra/terraform/nonprod validate
bash scripts/test_cuenta_lab.sh
```
