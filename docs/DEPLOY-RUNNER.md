# Despliegue en el propio host con runners self-hosted

Este documento describe la arquitectura y operación del despliegue directo en el host mediante runners self-hosted de GitHub Actions y el script [`deploy/host-deploy.sh`](../deploy/host-deploy.sh).

Este diseño desacopla los despliegues de credenciales de AWS desde GitHub: las máquinas ejecutan el despliegue localmente descargando imágenes inmutables de GitHub Packages (GHCR) y gestionando Docker Compose directamente en la instancia.

---

## 1. Arquitectura

### Runners y Ambientes

Cada máquina EC2 aloja uno o más runners de GitHub Actions configurados con etiquetas (*labels*) específicas:

| Host | Ambientes alojados | Labels del runner | Puertos en el host | Base de datos |
| --- | --- | --- | --- | --- |
| **EC2 Compartida** | `develop` | `campus-develop` | Web: `8088`, API: `8091`, PG: `5432` | `campus_verde_develop` |
| **EC2 Compartida** | `qa` | `campus-qa` | Web: `8188`, API: `8191`, PG: `5433` | `campus_verde_qa` |
| **EC2 Producción** | `produccion` | `campus-prod` | Web: `80`, API: `8091` (local), PG: `5432` (local) | `campus_verde` |

Los runners se ejecutan bajo el usuario de sistema `runner`, miembro del grupo `docker`, sin privilegios de superusuario root para la operación diaria.

### Layout en el host (`CAMPUS_HOME`)

Cada ambiente mantiene un directorio raíz aislado:

- `develop`: `/opt/campus/develop`
- `qa`: `/opt/campus/qa`
- `produccion`: `/opt/campus` (stack legado y vigente)

Estructura dentro de cada `CAMPUS_HOME`:

```
$CAMPUS_HOME/
├── host.env              # Variables y secretos locales (modo 600, NO versionado)
├── docker-compose.legacy.yml # Respaldo del compose previo (solo producción)
├── data/
│   ├── pg/               # Bind mount de datos de PostgreSQL/PostGIS
│   ├── app/              # Bind mount de archivos y evidencias (UID 10001)
│   └── backups/          # Volcados lógicos pg_dump de pre-despliegue
└── state/
    ├── prev-images.env   # Imágenes previas para rollback automático
    ├── deployed.env      # SHA, fecha e imágenes actualmente en servicio
    ├── conteos-antes.tsv # Conteos de tablas previos al despliegue (producción)
    └── conteos-despues.tsv # Conteos de tablas posteriores (producción)
```

### Gestión de Claves y Secretos

Las claves de base de datos y de aplicación **no se almacenan en GitHub Secrets** ni provienen de AWS Parameter Store o SSM. Se definen directamente en `$CAMPUS_HOME/host.env` (permisos `600`, propiedad del operador o runner).

Ejemplos documentados sin secretos se encuentran en:
- [`deploy/env/host.env.example.develop`](../deploy/env/host.env.example.develop)
- [`deploy/env/host.env.example.qa`](../deploy/env/host.env.example.qa)
- [`deploy/env/host.env.example.produccion`](../deploy/env/host.env.example.produccion)

**Reglas de seguridad en producción:**
En `produccion`, `CAMPUS_DEV_PASSWORD` y `POSTGRES_PASSWORD` deben tener $\ge 16$ caracteres y no pueden ser claves de laboratorio (`pando-local` ni `campus-lab`). El script de despliegue rechaza el inicio si detecta claves inválidas.

---

## 2. Compose Override para el Host (`compose.host.yml`)

El despliegue en el host combina tres archivos:
1. `deploy/compose.yml` (definición base)
2. `deploy/compose.<ambiente>.yml` (perfil del ambiente)
3. `deploy/compose.host.yml` (override específico para el host)

Características de [`deploy/compose.host.yml`](../deploy/compose.host.yml):
- **Imágenes GHCR:** usa `${API_IMAGE}` y `${WEB_IMAGE}` apuntando a referencias inmutables `ghcr.io/<owner>/campus-verde-api:<sha>`.
- **Sin build local:** elimina la sección de construcción (`build: !reset null`, requiere Compose $\ge 2.24$).
- **Aislamiento de red:** publica `db` y `api` estrictamente en `127.0.0.1` (`ports: !override`), evitando exposición a redes externas. Publica `web` en `0.0.0.0:${WEB_PORT}`.
- **Persistencia en disco:** monta bind mounts `${CAMPUS_DATA_DIR}/pg` y `${CAMPUS_DATA_DIR}/app` (`volumes: !override`), anulando los volúmenes nombrados para impedir que se cree una base de datos vacía y se dispare el ETL accidentalmente.

---

## 3. Flujo de Despliegue con `deploy/host-deploy.sh`

El script [`deploy/host-deploy.sh`](../deploy/host-deploy.sh) ejecuta el siguiente ciclo:

1. **Validación:** valida el ambiente (`develop`, `qa`, `produccion`), la longitud del SHA (40 hex), la presencia de `$CAMPUS_HOME/host.env` y las reglas de seguridad de contraseñas.
2. **Descarga de imágenes:** ejecuta `docker pull` de las imágenes de API y Web correspondientes al SHA desde GHCR.
3. **Salvaguardas de producción (ANTES de cualquier cambio en contenedores):**
   - **Conteos «antes»:** ejecuta [`deploy/conteos.sql`](../deploy/conteos.sql) en el contenedor de base de datos para registrar el número de filas en cada tabla.
   - **Backup lógico:** genera un volcado `pg_dump -Fc` en `${CAMPUS_DATA_DIR}/backups/pre-deploy-<UTC>.dump`.
   - **Snapshot EBS:** obtiene `instance-id` y región vía IMDSv2 (con token seguro de sesión), busca el volumen de datos con etiqueta `Name=campus-verde-data` y crea un snapshot EBS con tags de auditoría. Si el snapshot falla, el despliegue se detiene de inmediato (salvo declaración explícita de `DEPLOY_SIN_SNAPSHOT=1`).
4. **Registro de estado previo:** guarda las imágenes actualmente en servicio en `$CAMPUS_HOME/state/prev-images.env`.
5. **Migración de stack legado (producción, primera vez):** si se detectan contenedores del proyecto `campus`, los detiene con `docker compose -p campus stop` (nunca `down -v`) y respalda `docker-compose.legacy.yml`.
6. **Permisos:** asegura que `${CAMPUS_DATA_DIR}/app` pertenezca al UID `10001` (usuario de la API en el contenedor).
7. **Arranque:** ejecuta `docker compose up -d --no-build --remove-orphans`.
8. **Espera de salud:** comprueba hasta 45 intentos que `campus-<ambiente>-db` alcance estado `healthy`.
9. **Smoke test:** ejecuta [`deploy/smoke.sh`](../deploy/smoke.sh) contra `http://127.0.0.1:${WEB_PORT}`.
   - Si el smoke falla: ejecuta **rollback automático** a las imágenes previas (o reinicia el stack legado si era la primera migración) y sale con error `1`.
10. **Conteos «después» (producción):** ejecuta nuevamente [`deploy/conteos.sql`](../deploy/conteos.sql) y compara con [`deploy/comparar_conteos.py`](../deploy/comparar_conteos.py) usando las exclusiones de [`deploy/conteos.excluir`](../deploy/conteos.excluir). Si alguna tabla de negocio varía en filas, el despliegue aborta con error.
11. **Registro:** escribe `$CAMPUS_HOME/state/deployed.env` con el SHA y fecha, y reporta en `$GITHUB_STEP_SUMMARY` si está disponible.

---

## 4. Cómo Desplegar de Ahora en Adelante

### Despliegue Automático

- **Develop:** Cada `push` a `develop` o `backend/arquitectura-equipo` compila, prueba y publica imágenes en GHCR. El workflow de deploy envía el trabajo al runner con label `campus-develop`.
- **QA:** Cada tag `rc-*` publica imágenes con ese tag y despliega en el runner con label `campus-qa`.

### Despliegue Manual (Producción y Contingencias)

Para desplegar producción mediante GitHub Actions:

```bash
gh workflow run deploy.yml \
  --ref backend/arquitectura-equipo \
  -f ambiente=produccion \
  -f ref=<sha-40-hex>
```

Para ejecutar el despliegue directamente en la máquina EC2 (como usuario `runner`):

```bash
# Simulación previa sin cambios
deploy/host-deploy.sh produccion --sha <sha-40-hex> --dry-run

# Despliegue real
deploy/host-deploy.sh produccion --sha <sha-40-hex>
```

---

## 5. Cómo Hacer Rollback

### Vía GitHub Actions

```bash
gh workflow run deploy.yml \
  --ref backend/arquitectura-equipo \
  -f ambiente=produccion \
  -f rollback=<sha-o-tag-anterior>
```

### Directamente en el Host

```bash
deploy/host-deploy.sh produccion \
  --sha <sha-actual> \
  --rollback <sha-o-tag-anterior>
```

### Rollback Automático Integrado

Si tras levantar los nuevos contenedores el smoke test falla (`/health`, login ficticio, OpenAPI/Swagger), el script revierte automáticamente la imagen de la API y de la web a las registradas en `$CAMPUS_HOME/state/prev-images.env`.

> [!IMPORTANT]
> **Las migraciones son aditivas y solo hacia adelante:** el rollback revierte el código de la aplicación (contenedores de API y Web). No se revierten las migraciones de base de datos ni se eliminan columnas creadas (todas las migraciones en `db/migrations` son aditivas y compatibles con versiones anteriores).

---

## 6. Registro de un Runner Self-Hosted

La instalación y registro del runner en las instancias EC2 la realiza el script de aprovisionamiento del operador o agente de infraestructura:

- Referencia: `scripts/runner-instalar.sh`
- Parámetros requeridos: URL del repositorio, token de registro de GitHub Actions y etiquetas (`campus-develop`, `campus-qa` o `campus-prod`).
- El servicio se configura en systemd para iniciar en el arranque (`systemctl enable --now actions.runner.*`).

---

## 7. Rotación de Claves

1. Acceder al host vía SSM o consola de administración.
2. Editar `$CAMPUS_HOME/host.env` (permisos `600`).
3. Si se rota la contraseña de base de datos (`POSTGRES_PASSWORD`):
   ```bash
   docker exec -i campus-<ambiente>-db psql -U campus -d campus_verde -c "ALTER USER campus WITH PASSWORD '<nueva-clave>';"
   ```
   Actualizar `POSTGRES_PASSWORD` en `$CAMPUS_HOME/host.env`.
4. Si se rota la clave del sistema (`CAMPUS_DEV_PASSWORD`):
   Actualizar `CAMPUS_DEV_PASSWORD` en `$CAMPUS_HOME/host.env`.
5. Reiniciar los contenedores aplicando el archivo de entorno:
   ```bash
   deploy/host-deploy.sh <ambiente> --sha <sha-actual>
   ```

---

## 8. Seguridad de Runners Self-Hosted en Repositorios Públicos

> [!CAUTION]
> **Riesgo crítico de seguridad:** En repositorios públicos, ejecutar runners self-hosted en eventos `pull_request` permitiría que cualquier usuario externo envíe código malicioso en un PR y lo ejecute con acceso directo a la red privada y al daemon de Docker de la EC2.

**Reglas de mitigación obligatorias:**
1. **Nunca habilitar el runner en `pull_request`:** los workflows con `runs-on: [self-hosted, ...]` solo deben activarse en eventos `push` de ramas protegidas (`develop`, `backend/arquitectura-equipo`), tags de versión (`rc-*`) o ejecuciones manuales (`workflow_dispatch`).
2. **Aprobación requerida en producción:** el ambiente `produccion` en GitHub requiere revisión manual obligatoria (*Required reviewers*).
3. **Usuario sin privilegios root:** el proceso del runner corre bajo el usuario `runner`, restringido al grupo `docker`.
4. **Puertos de base de datos y API locales:** los puertos de backend y BD están expuestos únicamente en `127.0.0.1`.

---

## 9. Limitaciones Actuales

- **TLS / HTTPS pendiente:** El frontend expone tráfico HTTP directo (puerto 80 en producción, 8088 en develop, 8188 en qa). La terminación TLS mediante reverse proxy (Caddy / Nginx / Certbot / ALB) se configurará en un hito posterior. Por esta razón, `CAMPUS_COOKIE_SECURE` permanece en `false` hasta activar certificados SSL/TLS.
- **Almacenamiento S3 pendiente:** Los archivos y evidencias se almacenan localmente en el bind mount `/opt/campus/data/app` (`CAMPUS_DATA_DIR/app`). La migración a bucket S3 privado se implementará en la siguiente fase de infraestructura.
