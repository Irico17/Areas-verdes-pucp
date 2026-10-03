# Despliegue de VerdePUCP: guía rápida

Este documento explica cómo se despliega VerdePUCP en AWS (develop, qa y producción). **No contiene contraseñas.** Las claves viven en los *environments* de GitHub y en archivos `.env` locales que nunca se suben al repositorio.

Detalle técnico complementario: `docs/DEPLOY-RUNNER.md` (runners), `docs/AMBIENTES.md` (ambientes), `docs/RUNBOOK-CORTE-PRODUCCION.md` (corte de producción), [`docs/CAMBIO-DE-CUENTA-LAB.md`](CAMBIO-DE-CUENTA-LAB.md) (otra cuenta de Learner Lab).

## 1. Resumen en una pantalla

```
push a develop ─► CI (tests, escaneo, build de imágenes por SHA → ghcr.io)
                    └─► deploy develop (runner campus-develop, EC2 no-prod)
                          └─► (mismo SHA) deploy qa  (runner campus-qa, misma EC2 no-prod)
                                └─► (mismo SHA) deploy produccion (runner campus-prod, EC2 de producción)
                                      └─ requiere aprobación manual del revisor del environment
```

- GitHub construye y publica las imágenes. Los servidores **no usan credenciales de AWS para desplegar**: cada EC2 ejecuta un runner self-hosted de GitHub Actions que descarga la imagen de ghcr y levanta Docker Compose en el propio host.
- Las credenciales de AWS (Learner Lab) solo se necesitan para **aprovisionar** (Terraform) o administrar por SSM. No hacen falta para desplegar.

## 2. Ramas

| Rama | Uso |
|---|---|
| `develop` | Integración. Cada push dispara CI y despliega a develop automáticamente. |
| `backend/arquitectura-equipo` | Rama de trabajo; se mantiene alineada con `develop`. |
| `main` | Rama por defecto del repo; no despliega. |
| Tags `rc-*` | Un tag `rc-…` dispara CI y despliega a **qa** (también deja las imágenes etiquetadas con ese tag, útil para rollback). |

Los cambios que solo tocan `*.md` o `docs/**` no disparan CI ni despliegue (`paths-ignore`).

## 3. CI (`.github/workflows/ci.yml`)

- Tests Go (api y backend con PostGIS), lint/tests/build del frontend, pruebas de los scripts de despliegue, deriva de Swagger y revisiones de seguridad de migraciones (sin `DROP`/`TRUNCATE`/`DELETE FROM` desde la 047, sin `AutoMigrate`, sin `initdb`).
- Escaneo **no bloqueante**: gitleaks y Trivy (fs e imágenes, con la imagen de Trivy fijada por digest).
- Permisos mínimos (`contents: read`; solo el job de imágenes tiene `packages: write`). Todas las acciones de terceros están fijadas por SHA.
- Job `imagenes`: solo en push a `develop`, `backend/arquitectura-equipo` o tags `rc-*`. Publica `ghcr.io/irico17/campus-verde-api:<sha>` y `ghcr.io/irico17/campus-verde-web:<sha>` (SHA completo de 40 caracteres). Caché de capas en el registro (`:buildcache`).
- Los PR de forks necesitan aprobación para ejecutar workflows. Los workflows de despliegue nunca se disparan con `pull_request`.

## 4. Despliegue (`.github/workflows/deploy.yml`)

Jobs: `resolver` → `esperar-ci` → `promocion` → `deploy` (en el runner del ambiente) → verificación pública.

1. **resolver**: decide ambiente y SHA exacto; valida entradas.
2. **esperar-ci**: espera a que el CI del mismo SHA termine en verde. Si falla, no se despliega.
3. **promocion**: qa exige un despliegue exitoso de **ese mismo SHA** en develop; producción exige uno exitoso en qa. En un rollback explícito se omite.
4. **deploy**: corre en `runs-on: [self-hosted, campus-<ambiente>]`, entra a ghcr con el `GITHUB_TOKEN` del run y ejecuta `deploy/host-deploy.sh`.
5. **Concurrencia** por ambiente: no hay dos despliegues del mismo ambiente a la vez.

### Qué hace `deploy/host-deploy.sh`

1. Lee `host.env` del servidor (claves del ambiente; modo 600; nunca en el repo).
2. `docker pull` de las imágenes del SHA.
3. **Solo producción, antes de tocar nada:** conteos de tablas, `pg_dump -Fc` en `/opt/campus/data/backups/` y snapshot EBS del volumen de datos (usa el rol de la instancia; si no se puede, aborta salvo `DEPLOY_SIN_SNAPSHOT=1`).
4. Guarda las imágenes en uso (para rollback). En el primer despliegue de producción detiene —con `stop`, sin `down -v`— el stack antiguo `campus` y conserva su compose.
5. `docker compose up -d` con los datos en *bind mounts* (`pg/` y `app/`). Nunca se borran volúmenes ni datos.
6. Espera a que Postgres esté sano y ejecuta el **smoke** (`deploy/smoke.sh`): `/health`, login de una cuenta ficticia y dos lecturas.
7. Si el smoke falla: **rollback automático** a las imágenes previas (o reinicio del stack antiguo en la primera migración de producción).
8. Solo producción: conteos «después» y comparación con los «antes» (`deploy/comparar_conteos.py`, excluye tablas volátiles como sesiones). Solo una baja de conteo (o una tabla que desaparece) en tablas de negocio hace fallar el despliegue; los aumentos y las tablas nuevas se reportan en el log.

Las migraciones son **aditivas** y las aplica la API al arrancar. El ETL inicial (`data/raw`) solo carga si el catastro está vacío; develop y qa además reciben una semilla ficticia aditiva (`ON CONFLICT DO NOTHING`).

## 5. Cómo disparar cada despliegue

### develop
- Automático: `git push origin develop` (con CI en verde).
- Manual: `gh workflow run deploy.yml --repo Irico17/Areas-verdes-pucp --ref develop -f ambiente=develop -f ref=<sha-o-rama>`

### qa
- Con el mismo SHA que ya está en develop:
  `gh workflow run deploy.yml --repo Irico17/Areas-verdes-pucp --ref develop -f ambiente=qa -f ref=<sha40>`
- O con un tag: `git tag rc-2026-10-02-1 <sha> && git push origin rc-2026-10-02-1`.

### producción
1. Dispara: `gh workflow run deploy.yml --repo Irico17/Areas-verdes-pucp --ref develop -f ambiente=produccion -f ref=<sha40>`
2. El job `deploy` queda **esperando aprobación** (environment `produccion` con revisor requerido).
3. Apruébalo: en la web, *Actions → el run → Review deployments → produccion → Approve and deploy*; o con `gh`:
   `gh api repos/Irico17/Areas-verdes-pucp/actions/runs/<run_id>/pending_deployments` (obtén el `environment.id`) y luego
   `gh api -X POST repos/Irico17/Areas-verdes-pucp/actions/runs/<run_id>/pending_deployments -f state=approved -f comment="ok" -F 'environment_ids[]=<id>'`
4. Producción solo se despliega a mano (nunca por push) y solo desde la rama `develop`.

### Desde la web de GitHub
*Actions → deploy → Run workflow* (rama `develop`): elige `ambiente`, y opcionalmente `ref` y `rollback`.

## 6. Rollback

Vuelve a una versión anterior sin reconstruir nada, usando imágenes ya publicadas:

```
gh workflow run deploy.yml --repo Irico17/Areas-verdes-pucp --ref develop \
  -f ambiente=<develop|qa|produccion> -f ref=<sha-actual> -f rollback=<sha40-anterior | rc-...>
```

- `rollback` acepta un SHA de 40 hex o un tag `rc-*`. En producción también pasa por aprobación.
- Las migraciones son aditivas, por lo que la versión anterior sigue funcionando con la base actual.
- Si necesitas recuperar datos: `pg_dump` en `/opt/campus/data/backups/` y el snapshot EBS tomado antes del despliegue (ver `docs/RUNBOOK-CORTE-PRODUCCION.md`).

## 7. Ambientes, URLs y puertos

| Ambiente | Web | API (vía la web) | Swagger | Servidor |
|---|---|---|---|---|
| develop | variable `PUBLIC_URL` del environment `develop` (puerto 8088) | la misma URL + `/areas-verdes/v1` (también `/api/v1`) | `/swagger/index.html` | EC2 no-prod |
| qa | variable `PUBLIC_URL` del environment `qa` (puerto 8188) | la misma URL + `/areas-verdes/v1` | `/swagger/index.html` | misma EC2 no-prod |
| producción | variable `PUBLIC_URL` del environment `produccion` (`https://verde-pucp.duckdns.org` cuando hay PEM) | la misma URL + `/areas-verdes/v1` | desactivado (404) | EC2 de producción |

La IP no está en el workflow. La escribe `aprovisionar-cuenta` en `PUBLIC_URL` al terminar el apply, o se actualiza con `gh variable set`. Cómo cambiar de cuenta: [`CAMBIO-DE-CUENTA-LAB.md`](CAMBIO-DE-CUENTA-LAB.md).

Puertos (en el servidor):

| Ambiente | Web (público) | API (solo 127.0.0.1) | Postgres (solo 127.0.0.1) |
|---|---|---|---|
| develop | 8088 | 8091 | 5432 |
| qa | 8188 | 8191 | 5433 |
| producción | 80 | 8091 | 5432 |

Solo los puertos web están abiertos en el Security Group. **Postgres y la API no son públicos.** Las IP son Elastic IP, no cambian al reiniciar.

Estructura en cada servidor: `/opt/campus/<develop|qa>/{host.env,data/,state/}` en la EC2 no-prod y `/opt/campus/{host.env,data/,state/}` en producción; los runners están en `/opt/actions-runner-<ambiente>` como servicios systemd.

## 8. Dónde viven las claves

- **Servidor:** `host.env` (modo 600, dueño `runner`) con las claves de Postgres y de las cuentas de cada ambiente.
- **GitHub:** *Settings → Environments → develop / qa / produccion*. Cada environment tiene la variable `PUBLIC_URL`; `produccion` tiene revisor requerido y solo acepta la rama `develop`. El despliegue no necesita secretos de AWS.
- **PC del administrador:** archivos `.env` ignorados por git (`deploy/env/<ambiente>.env`, `apps/web/.env.<ambiente>`) y `CREDENCIALES-LOCAL.md` (ignorado). Plantillas sin secretos: `deploy/env/*.env.example`.
- **Nunca** se suben claves al repositorio. Las claves de producción deben tener ≥ 16 caracteres y no ser `pando-local` ni `campus-lab` (la API y el script lo rechazan).
- Cuentas ficticias de cada ambiente: `admin`, `coordinacion`, `jefatura`, `norte`, `sur`, `riego`. La clave es distinta por ambiente.
- Los secretos antiguos `AWS_*` y `TF_VAR_*` del repositorio ya no se usan y pueden borrarse.
- **Evidencias por ambiente.** `EVIDENCIAS_BUCKET` vacío guarda en disco (develop y la simulación local). Qa y producción documentan un cubo privado: el nombre se escribe en el `.env` del host, no en git. La base guarda la clave del objeto, no una URL pública. Las credenciales de ese cubo son del ambiente: el rol de la instancia (`s3:PutObject` y `s3:GetObject`) o, si no hay rol, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` y `AWS_SESSION_TOKEN` en el environment de GitHub (`qa` o `produccion`) o en `host.env`. El despliegue de las imágenes sigue sin necesitar esas claves. No se sube ninguna al repositorio.

## 9. Si algo falla

### El runner está offline (el job queda en «Queued»)
1. Revisa *Settings → Actions → Runners* (deben verse `campus-develop`, `campus-qa` y `campus-prod` en línea).
2. Entra al servidor por **SSM** (consola AWS → Systems Manager → Session Manager) y mira el servicio: `systemctl status 'actions.runner.*'`; reinicia con `sudo systemctl restart 'actions.runner.*'`.
3. Si el registro caducó: genera un token (`gh api -X POST repos/Irico17/Areas-verdes-pucp/actions/runners/registration-token`), guárdalo como SecureString en SSM `/campus/runner-token/<ambiente>` y ejecuta `scripts/runner-instalar.sh <develop|qa|produccion>` en el servidor. Borra el parámetro después.
4. Si el Learner Lab reinició las instancias, los runners vuelven solos (servicio systemd) y las IP se conservan.

### Caducaron las credenciales de AWS
El despliegue **no se ve afectado** (no usa AWS). Solo dejan de funcionar Terraform, SSM desde fuera y el acceso a la consola. Renueva las credenciales del Learner Lab (*AWS Details → Show*) con `scripts/actualizar-credenciales-lab.sh` cuando necesites aprovisionar o entrar por SSM. El snapshot EBS previo a producción lo toma la propia instancia con su rol, no tus credenciales. Si hay que pasar a la cuenta de otro estudiante, sigue [`CAMBIO-DE-CUENTA-LAB.md`](CAMBIO-DE-CUENTA-LAB.md): el estado vive en la cuenta nueva y los deploys no se editan.

### Troubleshooting

| Síntoma | Causa probable y acción |
|---|---|
| `esperar-ci` falla | El CI del SHA está en rojo o aún no existe. Corrige y vuelve a empujar. |
| `promocion` rechazada | El SHA no se desplegó antes en el ambiente previo. Despliega develop → qa → producción en orden. |
| No aparece la aprobación de producción | Verifica que el run se lanzó desde la rama `develop` y que `produccion` tiene al revisor configurado. |
| `docker pull` falla / 401 / «not found» | La imagen del SHA no existe (el CI no llegó al job `imagenes`) o el paquete de ghcr no está enlazado al repo. Revisa *Packages*. |
| Smoke falla y se hace rollback | Mira el log del paso «Desplegar en el host». En el servidor: `docker logs campus-<amb>-api`. Si la API reinicia en bucle, suele ser migración o base de datos. |
| Mapa sin áreas | La base está vacía o incompleta. La API debe reportar 521 áreas/534 zonas (528/540 en develop y qa con la semilla). Revisa el log de la API al arrancar. |
| 502 en la web | La API no está arriba (se está reiniciando o falló). `docker ps` y `docker logs` en el servidor. |
| Producción: «No se pudo crear el snapshot EBS» | El rol de la instancia no puede crear snapshots. Toma uno a mano o usa `DEPLOY_SIN_SNAPSHOT=1` solo con el `pg_dump` ya verificado. |
| Disco o memoria | La EC2 de producción es pequeña (1 GB RAM, con swap de 2 GB). Limpia imágenes viejas con `docker image prune` (nunca volúmenes ni `data/`). |

### Reglas que no se rompen
- No hay `docker compose down -v`, `TRUNCATE`, `DROP` ni borrado de datos cargados.
- No se hace force push ni se sube nada al repositorio del grupo; solo al repositorio personal.

## 10. Pendientes conocidos
- El nombre del cubo privado de qa y de producción se configura en el host (`EVIDENCIAS_BUCKET`). La plantilla del repositorio lo deja vacío.
- `etl-lote` no completa un catastro parcial en una base ya cargada (el arranque continúa sin él).
- La validación con la PUCP sigue abierta y la restricción de la universidad no está firmada. DuckDNS no es el dominio definitivo.

## 11. HTTPS con DuckDNS

Cubre el tramo público de DEC-09 y de RNF-03, y la parte de tránsito de RNF-04: el nombre `verde-pucp.duckdns.org`, el puerto 443 y la cookie `Secure` cuando hay certificado. No cierra la validación institucional. La validación con la PUCP sigue abierta; la restricción institucional no está firmada; DuckDNS es un dominio gratuito de laboratorio, no el dominio definitivo (pendiente de decisión del cliente).

### A qué ambiente apunta

A **producción** (`100.51.112.92`). Es el único ambiente que publica la web en el puerto 80 y cuyo grupo de seguridad ya abre el 443. Develop (`:8088`) y qa (`:8188`) comparten la otra EC2, siguen en HTTP y con `CAMPUS_COOKIE_SECURE=false`. Llevar el dominio ahí exigiría publicar el 443, abrirlo en ese grupo de seguridad y un nombre por ambiente: fuera de alcance.

Hasta que una persona mueva el DNS, `verde-pucp.duckdns.org` sigue apuntando a la IP doméstica. Este repositorio no llama a DuckDNS ni a Let's Encrypt ni cambia el registro.

### Por qué DNS-01 y por qué lego

El dominio hoy no apunta a la EC2, así que HTTP-01 no puede validar hasta después del cambio de DNS. DNS-01 emite el certificado **antes** de mover la IP y no necesita el puerto 80 libre. HTTP-01 queda como alternativa solo en producción, cuando el nombre ya apunte a `100.51.112.92` y el 80 siga expuesto.

El cliente es `lego` (imagen `goacme/lego` fijada por versión y digest en `deploy/tls/comun.sh`). Trae el proveedor `duckdns`, lee `DUCKDNS_TOKEN` del entorno y corre en el host. No se añade Caddy ni un segundo proxy: el nginx de `apps/web` ya copia `nginx.tls.conf` si al arrancar ve `/etc/nginx/certs/fullchain.pem` y `privkey.pem`.

### Dónde va el token

No entra al repositorio. Dos sitios, nunca un valor de ejemplo real:

1. En el host, `/opt/campus/duckdns.env` con modo 600, fuera de cualquier checkout y fuera de `/opt/campus/certs`. Plantilla: `deploy/tls/duckdns.env.example` (`DUCKDNS_TOKEN=` vacío y `DUCKDNS_DOMAIN=verde-pucp`). Otra ruta se indica con `DUCKDNS_ENV_FILE`.
2. Si se usa el workflow, el secreto de Actions `DUCKDNS_TOKEN`. El workflow lo enmascara con `::add-mask::`.

Si el token no está, los scripts salen con un error y no lo imprimen. El token de la cuenta DuckDNS controla todos los subdominios de esa cuenta: si se filtra, hay que rotarlo en DuckDNS y reescribir el archivo o el secreto.

### Ensayo en staging y emisión real

Los límites de tasa de Let's Encrypt castigan los reintentos del directorio real. Primero staging (`ACME_STAGING=1`, que es el valor por defecto). El certificado de staging no es de confianza en el navegador.

En el host de producción, con Docker y el archivo 600 ya creado:

```bash
sudo install -d -m 755 /opt/campus/bin
sudo install -m 755 deploy/tls/comun.sh deploy/tls/emitir-certificado.sh \
  deploy/tls/renovar-certificado.sh deploy/tls/actualizar-ip-duckdns.sh \
  /opt/campus/bin/
export DUCKDNS_ENV_FILE=/opt/campus/duckdns.env
export ACME_EMAIL=operador.campus@example.com
export ACME_STAGING=1
bash deploy/tls/emitir-certificado.sh
```

`ACME_EMAIL` es el contacto de Let's Encrypt, no un secreto. Cuando el ensayo haya dejado un PEM y `journalctl` no muestre el token, repita con `ACME_STAGING=0`. Eso pide el certificado real, todavía con el DNS antiguo: DNS-01 no necesita que la IP pública ya sea la de la EC2.

La cuenta ACME y los archivos de lego quedan en `/opt/campus/letsencrypt` (modo 700). Nginx solo monta `/opt/campus/certs`.

El workflow `.github/workflows/duckdns-cert.yml` es solo `workflow_dispatch`, corre en `[self-hosted, campus-prod]` y no se dispara solo. El input `confirmar` tiene que ser la palabra `EMITIR`. `staging` vale `1` por defecto. No actualiza la IP.

### Instalar el PEM y encender TLS

`emitir-certificado.sh` escribe, de forma atómica, `/opt/campus/certs/fullchain.pem` (644) y `privkey.pem` (600, dueño de quien corre Docker). Si el contenedor web **ya** estaba sirviendo TLS, hace `docker exec campus-produccion-web nginx -s reload`. No recrea contenedores.

La primera vez el entrypoint ya eligió la config HTTP, porque mira los PEM al arrancar. Orden, a mano, no desde este repo:

1. Dejar los dos PEM en `/opt/campus/certs`.
2. En `/opt/campus/host.env` (modo 600), solo con el par instalado: `CAMPUS_COOKIE_SECURE=true`, `PUBLIC_URL=https://verde-pucp.duckdns.org`, `CAMPUS_CORS_ORIGINS=https://verde-pucp.duckdns.org` y `CAMPUS_CERTS_DIR=/opt/campus/certs`. Sin PEM, deje HTTP y la cookie en false. El ejemplo comentado está en `deploy/env/host.env.example.produccion`.
3. Volver a desplegar (o `docker compose` con `deploy/compose.tls.yml`). Eso recrea la web: el entrypoint copia `nginx.tls.conf`. La API arranca con `CAMPUS_COOKIE_SECURE=true` porque ese override lo fija cuando el fichero TLS entra en el compose.
4. Comprobar. `host-deploy.sh` incluye `compose.tls.yml` solo si los dos PEM existen. Sin ellos no publica el 443 ni monta certificados, y db/api siguen en `127.0.0.1`.

`compose.tls.yml` publica `0.0.0.0:${WEB_PORT}:80` y `0.0.0.0:443:443` (o `WEB_TLS_PORT` si hace falta otro puerto en un ensayo). Nginx escucha 443 con TLS 1.2 y 1.3, redirige el resto del 80 a HTTPS y deja `GET /health` en el 80 para el timer. HSTS usa `max-age=31536000` (un año): un navegador que haya visto el encabezado seguirá pidiendo HTTPS durante ese plazo aunque se quite el certificado.

### Verificación

En la EC2, con el certificado real (el nombre tiene que coincidir; la conexión va a la máquina local):

```bash
curl -sS -o /dev/null -w '%{http_code}\n' http://127.0.0.1/health
curl -sSI http://127.0.0.1/ | head -n 5
curl -sS --resolve verde-pucp.duckdns.org:443:127.0.0.1 \
  -o /dev/null -w '%{http_code}\n' https://verde-pucp.duckdns.org/health
curl -sSI --resolve verde-pucp.duckdns.org:443:127.0.0.1 \
  https://verde-pucp.duckdns.org/health | grep -i strict-transport-security
```

El login de una cuenta ficticia, por ese mismo `--resolve`, debe devolver `Set-Cookie` con `Secure` y `HttpOnly`. `deploy/smoke.sh` lo exige cuando `SMOKE_BASE_URL` empieza por `https://`. `SMOKE_INSECURE_TLS=1` (curl `-k`) es solo para un certificado de prueba y no se activa solo. Con el PEM real, `host-deploy.sh` usa `https://verde-pucp.duckdns.org` y `--resolve verde-pucp.duckdns.org:443:127.0.0.1`.

Sintaxis de nginx, si hace falta repetirla a mano: `bash deploy/tls/test_nginx_tls.sh`.

### Renovación

`deploy/tls/renovar-certificado.sh` no llama a Let's Encrypt si el certificado sigue vigente más de 30 días. Si falla, el código de salida no es 0.

```bash
sudo cp deploy/tls/systemd/campus-cert-renew.service /etc/systemd/system/
sudo cp deploy/tls/systemd/campus-cert-renew.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now campus-cert-renew.timer
journalctl -u campus-cert-renew.service
```

El timer corre a las 03:00 y a las 15:00. El log no lleva el token. Habilítelo después de la primera emisión correcta. En staging deje `ACME_STAGING=1` en la unidad; en el certificado real, `0`.

### Actualizar la IP, al final

`deploy/tls/actualizar-ip-duckdns.sh` no hace nada si no le pasan `--ip <IPv4>` o `--auto`. `--dry-run` es el modo seguro: exige el token, no lo imprime y no llama a la red. El script rechaza `CI=true` y `GITHUB_ACTIONS=true`. No lo corra en un runner: `--auto` registraría la IP del runner.

Cuando ya exista el PEM real, la web recreada y la cookie Secure comprobada, una persona puede fijar la IP de producción:

```bash
bash deploy/tls/actualizar-ip-duckdns.sh --dry-run --ip 100.51.112.92
bash deploy/tls/actualizar-ip-duckdns.sh --ip 100.51.112.92
```

`--auto` es para un timer en el propio host de producción (`deploy/tls/systemd/campus-duckdns-ip.timer`, cada 5 minutos). No lo habilite antes de decidir el cambio. Mientras tanto el dominio sigue en la IP doméstica y no se ejecuta nada.

### Riesgos

- Hasta el cambio de DNS el nombre responde en otro sitio. No asuma que `https://verde-pucp.duckdns.org` llega a la EC2.
- Let's Encrypt limita el número de emisiones. Use staging antes del directorio real.
- El token DuckDNS administra todos los subdominios de la cuenta. Rótelo si se filtra.
- El Learner Lab apaga la EC2. La IP elástica se conserva; al volver hay que mirar que el contenedor web y el timer sigan activos.
- HSTS con un año de `max-age` hace que los navegadores que ya visitaron el sitio no vuelvan a HTTP con facilidad.
- Postgres no se publica: db y api siguen en `127.0.0.1`.

### Lo que este repositorio no hace

No despliega, no llama a la API de DuckDNS, no llama a Let's Encrypt, no ejecuta Terraform y no mueve el DNS. La emisión real con el token, en staging y luego en el directorio de producción, la corre el operador con los comandos de arriba.
