# Despliegue por sesión (Learner Lab)

Cada sesión del laboratorio dura unas cuatro horas y las credenciales cambian. El lab no deja crear roles IAM ni un proveedor OIDC. El job `deploy` de `.github/workflows/ci.yml` lee los secrets `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` y `AWS_SESSION_TOKEN`, y solo corre con `workflow_dispatch`.

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

El script usa el bloque del portapapeles si trae las tres claves. Si no, lee el perfil `[default]` de `~/.aws/credentials`. Actualiza los tres secrets con `gh secret set` y pregunta si lanza:

```bash
gh workflow run ci --ref main
```

No imprime los valores y no los escribe en el repositorio.

4. El job `test` corre primero. Si pasa, `deploy` aplica Terraform, sube las imágenes y reinicia `campus.service` por SSM. Al terminar imprime `Listo: http://<ip>`.

`TF_VAR_db_password` y `TF_VAR_dev_password` se cargan una vez como secrets del repo (16 caracteres o más). No cambian con la sesión del lab. El job `deploy` los lee junto con las tres claves de AWS. El detalle del stack está en `docs/DEPLOY-AWS.md`.

## Sin llave SSH

El despliegue no usa la llave SSH del laboratorio. `ssh_cidr` y `ssh_key_name` van vacíos: el puerto 22 queda cerrado y la instancia no recibe un key pair. La instancia usa el `LabInstanceProfile` que el lab ya trae. `scripts/poner-secretos.sh` y el reinicio de `campus.service` van por `aws ssm send-command` (`AWS-RunShellScript`).

## IP pública

Una IP pública efímera cambia cuando la instancia se detiene al cerrar la sesión. Terraform reserva `aws_eip.app` y la asocia a la instancia. Esa asociación se mantiene con la máquina detenida, así que `http://<elastic-ip>` no cambia al encenderla de nuevo. El lab lo permite con el rol existente; este módulo no crea IAM.

Si en otra sesión el lab liberó esa IP y Terraform asigna otra, el job imprime la URL nueva (`URL:` y `Listo:`). La Elastic IP asociada a una instancia detenida se cobra; véase `docs/DEPLOY-AWS.md`.

## Sesión por HTTP

El lab publica `http://<elastic-ip>` sin TLS. Una cookie `Secure` no se guarda en ese origen y el login parece no hacer nada: la API responde 200 y el navegador descarta `cv_sesion`.

`CAMPUS_COOKIE_SECURE` manda. Vale `true` solo con HTTPS. `CAMPUS_ENV=production` no la enciende: el compose de la instancia fija ese entorno también cuando nginx sigue en el puerto 80.

`scripts/poner-secretos.sh` escribe el valor así:

- Si al ejecutarlo ya está exportada `CAMPUS_COOKIE_SECURE`, usa ese valor.
- Si no, en la instancia mira `/opt/campus/certs/fullchain.pem` y `privkey.pem` (el mismo par que el contenedor web monta en `/etc/nginx/certs`). Con los dos archivos pone `true`. Sin ellos pone `false`.

Después de copiar o quitar el certificado hay que volver a correr `poner-secretos.sh` y reiniciar la API para que la cookie coincida con nginx.
