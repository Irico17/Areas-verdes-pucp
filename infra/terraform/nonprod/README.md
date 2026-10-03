# Infraestructura Compartida Nonprod (develop y qa)

Este directorio contiene las definiciones de Terraform para aprovisionar **una única instancia EC2 compartida** que aloja simultáneamente los ambientes `develop` y `qa` de VerdePUCP en el entorno AWS Academy Learner Lab (`us-east-1`).

Debido a las restricciones de AWS Academy Learner Lab (cuenta compartida de laboratorio):
- No es posible crear roles de IAM nuevos ni proveedores de identidad OIDC.
- Se reutiliza el perfil de instancia preexistente `LabInstanceProfile`.
- Los despliegues continuos posteriores de GitHub Actions no utilizan credenciales de AWS: la instancia ejecuta runners *self-hosted* dedicados por cada ambiente (`campus-develop` y `campus-qa`).

---

## Recursos aprovisionados

1. **Security Group (`campus-verde-nonprod`)**:
   - Ingress TCP restringido **exclusivamente** a los puertos web:
     - `8088`: frontend web de `develop` (desde `0.0.0.0/0`).
     - `8188`: frontend web de `qa` (desde `0.0.0.0/0`).
   - La API (`8091`/`8191`) y PostgreSQL (`5432`/`5433`) **no** se publican a internet: se comunican por redes internas de Docker.
   - El puerto `22` (SSH) está **cerrado** por defecto (`ssh_cidr = ""`). El acceso administrativo y operativo se realiza mediante AWS Systems Manager (SSM Session Manager).
   - Egress total (`0.0.0.0/0`) para permitir descargas de paquetes, imágenes de GHCR y comunicación del runner de GitHub Actions.

2. **Instancia EC2 (`campus-verde-nonprod`)**:
   - AMI: Amazon Linux 2023 x86_64 (`al2023-ami-2023*-x86_64`).
   - Tipo: `t3.small` por defecto (el crédito del lab es corto). `t3.medium` si los dos stacks se quedan sin memoria. La máquina que ya existía con 30/40 GB está en `cuenta-actual.tfvars.example`.
   - Disco raíz: 20 GB gp3, cifrado. Bajar un disco ya creado lo reemplaza.
   - Metadatos IMDSv2 obligatorios: `http_tokens = "required"`, `http_put_response_hop_limit = 1`.
   - Protección contra reemplazo accidental: `user_data_replace_on_change = false` y `lifecycle { ignore_changes = [ami, user_data] }`.

3. **Volumen EBS persistente (`campus-verde-nonprod-data`)**:
   - Tipo gp3 de 20 GB por defecto, cifrado (`/dev/sdf`).
   - Formateado con XFS solo en el primer arranque y montado persistentemente por UUID en `/opt/campus` (`nofail` en `/etc/fstab`).

4. **Dirección IP Elástica (`aws_eip`)**:
   - Asociada a la instancia EC2 para mantener estables las IPs y URLs públicas (`develop` en `http://<ip>:8088` y `qa` en `http://<ip>:8188`).

---

## ⚠️ Advertencia crítica sobre `terraform destroy`

Todos los recursos se definen con `prevent_destroy = false` para permitir su administración dentro de las limitaciones temporales del Learner Lab.

> [!CAUTION]
> **NUNCA ejecute `terraform destroy` si existen datos de catastro, auditorías o evidencias en `/opt/campus`.**
> La destrucción de la infraestructura eliminará la instancia y el volumen EBS de datos, perdiéndose de forma irreversible la información de ambos ambientes salvo que se cuente con un respaldo previo (`pg_dump` o snapshot manual de EBS).

---

## Gestión de Secretos y Token del Runner

Por diseño de seguridad y separación de responsabilidades:
- **Ningún secreto ni token sensible pasa por Terraform ni queda almacenado en el archivo de estado (`.tfstate`).**
- Los archivos `secrets.env` de cada ambiente (`/opt/campus/develop/secrets.env` y `/opt/campus/qa/secrets.env`) se inyectan directamente vía AWS SSM con permisos `0600` o mediante variables locales, nunca por `user_data`.
- Los tokens de registro de los runners self-hosted de GitHub Actions se consumen mediante el script [`scripts/runner-instalar.sh`](../../../scripts/runner-instalar.sh) (vía variable `RUNNER_TOKEN` o AWS Systems Manager Parameter Store) y no son gestionados por Terraform.

---

## Layout de directorios en `/opt/campus`

El disco de datos EBS se monta en `/opt/campus`. Durante el arranque inicial (`bootstrap.sh.tftpl`), se inicializa la siguiente jerarquía:

```text
/opt/campus/
├── BOOTSTRAP_OK               # Marcador de finalización exitosa del cloud-init
├── develop/                   # Espacio aislado del ambiente develop (chown runner:runner)
│   ├── docker-compose.yml     # Compose de develop
│   ├── secrets.env            # Variables sensibles (modo 600)
│   └── data/
│       ├── pg/                # Directorio de PostgreSQL develop (gestión interna del contenedor)
│       └── app/               # Evidencias y archivos de la API develop (UID:GID 10001:10001)
└── qa/                        # Espacio aislado del ambiente qa (chown runner:runner)
    ├── docker-compose.yml     # Compose de qa
    ├── secrets.env            # Variables sensibles (modo 600)
    └── data/
        ├── pg/                # Directorio de PostgreSQL qa (gestión interna del contenedor)
        └── app/               # Evidencias y archivos de la API qa (UID:GID 10001:10001)
```

Adicionalmente, el script de inicialización configura:
- 2 GB de memoria swap persistente (`/swapfile`).
- Usuario de sistema `runner` perteneciente al grupo `docker`.
- Servicio `amazon-ssm-agent` activo para permitir conexión remota sin llaves SSH.
- Plugin oficial de Docker Compose v2.29.7 verificado por SHA-256 en `/usr/local/lib/docker/cli-plugins/docker-compose`.

---

## Cómo aplicar la infraestructura

El camino repetible es el workflow `aprovisionar-cuenta` (`accion=plan` y después `apply` con `confirmar=APLICAR`). Carga las credenciales en el environment `aws-lab` con `scripts/actualizar-credenciales-lab.sh`. El detalle, incluido el bootstrap del cubo de estado, está en [`docs/CAMBIO-DE-CUENTA-LAB.md`](../../../docs/CAMBIO-DE-CUENTA-LAB.md).

A mano, con la sesión del lab ya exportada y el cubo creado por `scripts/bootstrap-estado-terraform.sh`:

```bash
bash scripts/bootstrap-estado-terraform.sh "$PWD/tmp/backend.hcl"
terraform -chdir=infra/terraform/nonprod init -backend-config="$PWD/tmp/backend.hcl"
terraform -chdir=infra/terraform/nonprod apply
```

Al concluir, Terraform muestra el `instance_id`, la `public_ip` y las URLs de `develop` y `qa`. Ansible registra los runners. Si lo hace a mano, como root por SSM:

```bash
scripts/runner-instalar.sh develop
scripts/runner-instalar.sh qa
```
