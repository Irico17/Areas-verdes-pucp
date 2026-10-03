# Despliegue por sesión (Learner Lab)

Hay tres ambientes (`develop`, `qa`, `produccion`): puertos, bases y el flujo develop → qa → producción están en [`AMBIENTES.md`](AMBIENTES.md). La guía del runner y el HTTPS con DuckDNS están en [`DESPLIEGUE-README.md`](DESPLIEGUE-README.md). Este archivo es la sesión del Learner Lab, que corresponde a **produccion**.

Cada sesión del laboratorio dura unas cuatro horas y las credenciales cambian. El lab no deja crear roles IAM ni un proveedor OIDC. El workflow `.github/workflows/deploy.yml` lee, del environment de GitHub (`develop`, `qa` o `produccion`), los secrets `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` y `AWS_SESSION_TOKEN`. Producción solo corre con `workflow_dispatch` y con la aprobación de ese environment. `ci.yml` prueba y publica las imágenes con el SHA; ya no despliega.

Hace falta `gh` autenticado en el repo (`gh auth login`) una sola vez.

## Cada sesión

1. En Vocareum, **Start Lab**. Espere el semáforo en verde.
2. **AWS Details** → **AWS CLI** → **Show**. Copie el bloque (no lo pegue en el chat ni en un archivo del repo).
3. Desde la raíz del repo, ejecute uno de estos:

```powershell
powershell -File scripts/set-aws-secrets.ps1
```

```bash
bash scripts/set-aws-secrets.sh
```

El script usa el bloque del portapapeles si trae las tres claves. Si no, lee el perfil `[default]` de `~/.aws/credentials`. Actualiza los tres secrets del repositorio con `gh secret set` y pregunta si lanza el workflow en la rama actual (no en `main`):

```bash
gh workflow run deploy --ref <rama> -f ambiente=produccion
```

Para dejar las claves solo en un ambiente: `gh secret set NOMBRE --env produccion`. La lista completa está en [`AMBIENTES.md`](AMBIENTES.md). No imprime los valores y no los escribe en el repositorio.

4. `ci.yml` tiene que estar en verde para ese SHA. `deploy.yml` aplica Terraform, publica el tag del SHA, en producción toma snapshot y backup, reinicia el compose por SSM y corre el smoke. Al terminar imprime `Listo: http://<ip>`. Si el environment `produccion` tiene revisores, el job espera esa aprobación.

`TF_VAR_db_password` y `TF_VAR_dev_password` se cargan una vez como secrets del repo (16 caracteres o más). No cambian con la sesión del lab. El job `deploy` los lee junto con las tres claves de AWS. El detalle del stack está en `docs/DEPLOY-AWS.md`.

## Sin llave SSH

El despliegue no usa la llave SSH del laboratorio. `ssh_cidr` y `ssh_key_name` van vacíos: el puerto 22 queda cerrado y la instancia no recibe un key pair. La instancia usa el `LabInstanceProfile` que el lab ya trae. `scripts/poner-secretos.sh` y el reinicio de `campus.service` van por `aws ssm send-command` (`AWS-RunShellScript`).

## IP pública

Una IP pública efímera cambia cuando la instancia se detiene al cerrar la sesión. Terraform reserva `aws_eip.app` y la asocia a la instancia. Esa asociación se mantiene con la máquina detenida, así que `http://<elastic-ip>` no cambia al encenderla de nuevo. El lab lo permite con el rol existente; este módulo no crea IAM.

Si en otra sesión el lab liberó esa IP y Terraform asigna otra, el job imprime la URL nueva (`URL:` y `Listo:`). La Elastic IP asociada a una instancia detenida se cobra; véase `docs/DEPLOY-AWS.md`.

## Sesión por HTTP

Sin el par PEM, el lab y los tres ambientes siguen en HTTP. Una cookie `Secure` no se guarda en ese origen y el login parece no hacer nada: la API responde 200 y el navegador descarta `cv_sesion`.

El nombre público y el HTTPS de producción (`verde-pucp.duckdns.org`, solo cuando hay PEM) están en [`DESPLIEGUE-README.md`](DESPLIEGUE-README.md). Develop y qa no usan ese dominio. La validación con la PUCP sigue abierta y la restricción de la universidad no está firmada.

`CAMPUS_COOKIE_SECURE` manda. Vale `true` solo con HTTPS. `CAMPUS_ENV=production` no la enciende: el compose de la instancia fija ese entorno también cuando nginx sigue en el puerto 80.

`scripts/poner-secretos.sh` escribe el valor así:

- Si al ejecutarlo ya está exportada `CAMPUS_COOKIE_SECURE`, usa ese valor.
- Si no, en la instancia mira `/opt/campus/certs/fullchain.pem` y `privkey.pem` (el mismo par que el contenedor web monta en `/etc/nginx/certs`). Con los dos archivos pone `true`. Sin ellos pone `false`.

Después de copiar o quitar el certificado hay que volver a correr `poner-secretos.sh` y reiniciar la API para que la cookie coincida con nginx.

## Esquema

La imagen de la API copia `db/migrations` a `/opt/campus/migrations` y el compose fija `MIGRATIONS_DIR=/opt/campus/migrations`. Postgres no monta ningún `.sql` en `docker-entrypoint-initdb.d`. `db/referencia/` no se aplica.

El corte de la base que ya está en la EC2 (backup, snapshot, conteos, tag de rollback) está en `docs/RUNBOOK-CORTE-PRODUCCION.md`. No es un paso de este flujo de sesión.

## Evidencias

`EVIDENCIAS_BUCKET` decide el almacén. Vacía: disco, como en develop, y la API lo dice una vez en el log. Con valor: los bytes van a un cubo S3 privado y Postgres guarda la clave del objeto, no una URL pública. La plantilla `deploy/env/*.env.example` deja la variable vacía. El nombre real no se commitea.

Las credenciales no van en el repositorio y son por ambiente:

| Ambiente | Cubo | Credenciales |
| --- | --- | --- |
| develop | vacío (disco) | no hacen falta |
| qa | cubo privado de qa, en el host | rol de la instancia, o las tres claves temporales en el environment `qa` o en `host.env` |
| produccion | cubo privado de producción, en el host | rol de la instancia, o las tres claves temporales en el environment `produccion` o en `host.env` |

Las tres claves, si no hay rol, son `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` y `AWS_SESSION_TOKEN`. El cubo debe bloquear el acceso público. El rol, si se usa, necesita `s3:PutObject` y `s3:GetObject` sobre ese cubo. Este flujo de sesión no crea el cubo ni escribe su nombre.
