# Operación de Campus Verde

Runbook del piloto en AWS Academy y de una cuenta propia. No despliega nada por sí solo. El Learner Lab no es producción: las credenciales duran unas cuatro horas y la EC2 se detiene al cerrar la sesión.

## Encender

1. En una cuenta propia, o con el lab en verde, exporte credenciales y la región `us-east-1`.
2. Defina las claves fuera del repo. No use valores de laboratorio.

```bash
export TF_VAR_db_password='(16 caracteres o más)'
export TF_VAR_dev_password='(16 caracteres o más)'
```

3. Desde el raíz: `bash scripts/deploy-learner-lab.sh` solo si va a usar el lab. El script aplica Terraform, sube las imágenes y escribe `/opt/campus/secrets.env` por SSM (`scripts/poner-secretos.sh`). Esas claves no van en el user data ni en el estado.
4. Si la instancia ya existe, no hace falta reemplazarla. Suba las imágenes a ECR y ejecute `systemctl restart campus.service`. Ese unit hace `docker compose pull` y `up`. `user_data_replace_on_change` está en false y el AMI se ignora en cambios posteriores: un apply que solo cambia el tag de la imagen no recrea la EC2.
5. Entre con la clave de `TF_VAR_dev_password`. La cookie sale `Secure` porque el compose de la instancia fija `CAMPUS_ENV=production`.

SSH queda cerrado salvo que pase `ssh_cidr` con una IP (`x.x.x.x/32`). Postgres no publica puerto.

## Certificado TLS

Copie `fullchain.pem` y `privkey.pem` a `/opt/campus/certs/` en la instancia y reinicie el servicio `web`. Nginx escucha 443, redirige el 80 a HTTPS y añade HSTS. Sin esos dos archivos sigue en HTTP, que es el caso del lab sin dominio.

El origen de la PWA, si no es el mismo host, va en `CAMPUS_CORS_ORIGINS` antes de `poner-secretos.sh`. No use `*`.

## Backup de PostGIS

Cada día, en la instancia o desde un host que llegue a la base:

```bash
DATABASE_URL='postgres://campus:CLAVE@127.0.0.1:5432/campus_verde' \
  bash scripts/backup-postgis.sh /opt/campus/backups/campus.dump
```

El formato es custom (`pg_dump -Fc`) e incluye la extensión PostGIS. Conserve el archivo fuera del disco de la instancia (S3 del cubo de estado, u otro cubo con versionado). El cubo de evidencias, si `create_evidence_bucket` está en true, queda con versionado y con el acceso público bloqueado.

Un timer de ejemplo, en la EC2, a las 06:10 UTC:

```bash
10 6 * * * DATABASE_URL='postgres://campus:CLAVE@127.0.0.1:5432/campus_verde' /opt/campus/backup-postgis.sh /opt/campus/backups/diario.dump
```

Copie el script del repo a `/opt/campus/` la primera vez. La base del compose no publica el puerto: en la instancia use `docker compose exec` (el script lo hace si no hay `DATABASE_URL` y el servicio `db` está arriba) o el socket local.

## Restaurar

La base destino tiene que existir y estar vacía. No restaure encima de `campus_verde` de producción sin haberla renombrado.

```bash
psql "$ADMIN_URL" -c 'CREATE DATABASE campus_restore'
TARGET_DATABASE_URL='postgres://campus:CLAVE@127.0.0.1:5432/campus_restore' \
  bash scripts/restore-postgis.sh /opt/campus/backups/campus.dump
```

Prueba hecha en este frente: se crea una fila con geometría, se vuelca, se restaura en otra base vacía y el conteo coincide. Comando: `bash scripts/probar-restauracion.sh`.

RPO, RTO y retención no están acordados con la universidad. Hasta ese escrito, trate el volcado diario como medida mínima y no como compromiso.

## Rotar una clave

1. Genere la nueva fuera del repo.
2. Exporte `TF_VAR_db_password` o `TF_VAR_dev_password`.
3. `bash scripts/poner-secretos.sh <instance-id>`.
4. Si rotó Postgres, cámbiela también dentro de la base (`ALTER USER campus PASSWORD ...`) antes de reiniciar, o el compose no podrá entrar.
5. `systemctl restart campus.service`.
6. La clave de las cuentas locales solo afecta a usuarios que todavía no existen: `Ensure` no reescribe hashes ya guardados. Para rotar una cuenta existente haga `UPDATE` del `password_hash` con bcrypt, no un apply.

Terraform no recibe la clave en un recurso. No hace falta `apply` para rotar.

## Healthcheck

`scripts/healthcheck.sh` (y `/usr/local/bin/campus-healthcheck` en la instancia) pide `GET http://127.0.0.1/health` sin seguir redirecciones. Si no responde o `status` no es `ok`, escribe `ALERTA` y sale con código 1. En la EC2 el timer `campus-health.timer` lo corre cada cinco minutos y manda el fallo al journal (`journalctl -t campus-health`).

Con certificado, el resto del puerto 80 redirige a HTTPS. `/health` no entra en esa redirección: el timer sigue recibiendo 200 y el cuerpo con `"status":"ok"`.

```bash
HEALTH_URL=http://127.0.0.1/health bash scripts/healthcheck.sh
```

En una cuenta propia, una alarma de CloudWatch sobre ese fallo es el paso siguiente. El lab no trae esa alarma.

## Logs

La API escribe una línea por petición: `request_id`, método, ruta, estado, latencia y bytes. No escribe el cuerpo, la query, la cookie ni `Authorization`. El id también va en la cabecera `X-Request-ID`.

En la instancia: `docker compose -f /opt/campus/docker-compose.yml logs api`. No hay agregador. No busque claves en ese log; si aparecen, es un fallo y hay que rotarlas.

## Parar

En el lab, al cerrar la sesión la instancia se detiene y el Elastic IP detenido se cobra. Para soltar el saldo: `terraform destroy` desde `infra/terraform`, con credenciales vivas. El cubo de estado no se borra con el destroy.

## Costo aproximado fuera del Learner Lab

Cifras de lista us-east-1, un mes, orden de magnitud. No es una cotización. No incluye NAT ni ALB: no los encienda (solo el balanceador ya ronda 16–20 USD).

| Pieza | Aprox. |
| --- | --- |
| t3.small encendida | 15 USD |
| RDS Postgres `db.t3.micro` con 20 GB, o la EC2 actual con backup | 15–20 USD si se separa la base; 0 extra si sigue en el volumen |
| EBS gp3 20 GB + snapshots diarios de 20 GB | 4–8 USD |
| S3 evidencias, pocos GB, más versionado | 1–3 USD |
| Route 53 + certificado ACM | menos de 1 USD |
| IP elástica en uso | 0 |
| **Orden de magnitud** | **25–45 USD al mes** |

El piloto del Learner Lab sigue en unos 12 USD si se dejara encendido el mes entero (t3.micro, dos discos, sin ALB) y cabe en el saldo del curso solo si se destruye al cerrar. Producción de verdad pide una cuenta con presupuesto propio.
