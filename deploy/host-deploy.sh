#!/usr/bin/env bash
# Despliegue en el propio host para runners self-hosted (campus-develop, campus-qa, campus-prod).
# Uso: deploy/host-deploy.sh <develop|qa|produccion> --sha <40hex> [--rollback <sha40|rc-*>] [--dry-run]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

usage() {
  echo "Uso: deploy/host-deploy.sh <develop|qa|produccion> --sha <40hex> [--rollback <sha40|rc-*>] [--dry-run]" >&2
  exit 2
}

AMBIENTE="${1:-}"
case "$AMBIENTE" in
  develop|qa|produccion) ;;
  *) usage ;;
esac
shift

sha=""
rollback=""
dry_run=0

while [ $# -gt 0 ]; do
  case "$1" in
    --sha)
      shift
      sha="${1:-}"
      ;;
    --rollback)
      shift
      rollback="${1:-}"
      ;;
    --dry-run)
      dry_run=1
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo "Opción desconocida: $1" >&2
      usage
      ;;
  esac
  shift
done

if [ -z "$sha" ]; then
  echo "ERROR: Debe especificar --sha con el commit de 40 caracteres hexadecimales." >&2
  usage
fi

if ! [[ "$sha" =~ ^[0-9a-f]{40}$ ]]; then
  echo "ERROR: El parámetro --sha debe ser exactamente de 40 dígitos hexadecimales (valor: $sha)." >&2
  exit 1
fi

if [ -n "$rollback" ]; then
  if ! [[ "$rollback" =~ ^[0-9a-f]{40}$ || "$rollback" =~ ^rc-[A-Za-z0-9._-]+$ ]]; then
    echo "ERROR: El parámetro --rollback debe ser un SHA de 40 dígitos hexadecimales o un tag rc-* (valor: $rollback)." >&2
    exit 1
  fi
fi

target_tag="${rollback:-$sha}"

# 1. Layout por ambiente (variable CAMPUS_HOME)
if [ -z "${CAMPUS_HOME:-}" ]; then
  case "$AMBIENTE" in
    develop) CAMPUS_HOME="/opt/campus/develop" ;;
    qa) CAMPUS_HOME="/opt/campus/qa" ;;
    produccion) CAMPUS_HOME="/opt/campus" ;;
  esac
fi

# Un solo despliegue a la vez por ambiente en este servidor. Pueden coexistir runners de más de
# un repositorio (cada uno con su propio concurrency); el candado por archivo los serializa.
if [ "$dry_run" -ne 1 ]; then
  mkdir -p "$CAMPUS_HOME/state"
  exec 9>"$CAMPUS_HOME/state/deploy.lock"
  echo "Esperando el candado de despliegue de ${AMBIENTE} (${CAMPUS_HOME}/state/deploy.lock)..."
  if ! flock -w 1800 9; then
    echo "ERROR: Otro despliegue de ${AMBIENTE} sigue en curso tras 30 minutos." >&2
    exit 1
  fi
fi

HOST_ENV="$CAMPUS_HOME/host.env"
if [ ! -f "$HOST_ENV" ]; then
  echo "ERROR: No existe el archivo de entorno en el host: $HOST_ENV" >&2
  echo "El operador debe crear este archivo con permisos 600 (ver deploy/env/host.env.example.${AMBIENTE})." >&2
  exit 1
fi

# Cargar variables del ambiente desde host.env
set -a
# shellcheck disable=SC1090
source "$HOST_ENV"
set +a

if [ "${APP_ENV:-}" != "$AMBIENTE" ]; then
  echo "ERROR: APP_ENV en $HOST_ENV (${APP_ENV:-}) no coincide con el ambiente solicitado ($AMBIENTE)." >&2
  exit 1
fi

CAMPUS_DATA_DIR="${CAMPUS_DATA_DIR:-$CAMPUS_HOME/data}"
POSTGRES_USER="${POSTGRES_USER:-campus}"
POSTGRES_DB="${POSTGRES_DB:-campus_verde}"
POSTGRES_PORT="${POSTGRES_PORT:-}"
API_PORT="${API_PORT:-}"
WEB_PORT="${WEB_PORT:-}"
CAMPUS_DEV_PASSWORD="${CAMPUS_DEV_PASSWORD:-}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-}"
CAMPUS_COOKIE_SECURE="${CAMPUS_COOKIE_SECURE:-false}"
CAMPUS_ENV="${CAMPUS_ENV:-$AMBIENTE}"
PUBLIC_URL="${PUBLIC_URL:-http://127.0.0.1:${WEB_PORT}}"
CAMPUS_CERTS_DIR="${CAMPUS_CERTS_DIR:-/opt/campus/certs}"
WEB_TLS_PORT="${WEB_TLS_PORT:-443}"
LEGACY_PROJECT="${LEGACY_PROJECT:-campus}"
LEGACY_COMPOSE="${LEGACY_COMPOSE:-$CAMPUS_HOME/docker-compose.yml}"

for req_var in POSTGRES_PORT API_PORT WEB_PORT CAMPUS_DEV_PASSWORD POSTGRES_PASSWORD; do
  if [ -z "${!req_var:-}" ]; then
    echo "ERROR: Falta definir la variable requerida $req_var en $HOST_ENV." >&2
    exit 1
  fi
done

# Reglas de claves en producción
if [ "$AMBIENTE" = "produccion" ]; then
  for pass_var in CAMPUS_DEV_PASSWORD POSTGRES_PASSWORD; do
    val="${!pass_var:-}"
    trimmed="$(printf '%s' "$val" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
    if [ "$trimmed" = "pando-local" ] || [ "$trimmed" = "campus-lab" ]; then
      echo "ERROR: En producción, $pass_var no puede ser una clave de laboratorio ($trimmed)." >&2
      exit 1
    fi
    if [ "${#trimmed}" -lt 16 ]; then
      echo "ERROR: En producción, $pass_var debe tener al menos 16 caracteres (longitud actual: ${#trimmed})." >&2
      exit 1
    fi
  done
fi

# Resolución de imágenes GHCR
OWNER="${GHCR_OWNER:-${GITHUB_REPOSITORY_OWNER:-}}"
if [ -z "$OWNER" ]; then
  remote_url="$(git -C "$ROOT" remote get-url origin 2>/dev/null || true)"
  if [[ "$remote_url" =~ github\.com[:/]([^/]+)/ ]]; then
    OWNER="${BASH_REMATCH[1]}"
  fi
fi
if [ -z "$OWNER" ]; then
  OWNER="campus-verde"
fi
OWNER_LC="$(printf '%s' "$OWNER" | tr '[:upper:]' '[:lower:]')"

API_IMAGE="ghcr.io/${OWNER_LC}/backend-campus-verde:${target_tag}"
WEB_IMAGE="ghcr.io/${OWNER_LC}/frontend-campus-verde:${target_tag}"
export API_IMAGE WEB_IMAGE CAMPUS_DATA_DIR APP_ENV POSTGRES_PORT API_PORT WEB_PORT
export CAMPUS_CERTS_DIR WEB_TLS_PORT CAMPUS_COOKIE_SECURE

echo "=== Despliegue en host: $AMBIENTE ==="
echo "Target Tag: $target_tag"
echo "API Image:  $API_IMAGE"
echo "Web Image:  $WEB_IMAGE"
echo "Home:       $CAMPUS_HOME"
echo "Data Dir:   $CAMPUS_DATA_DIR"
if [ "$dry_run" -eq 1 ]; then
  echo "Modo:       DRY-RUN (sin efectos reales)"
fi

# Pull de imágenes
if [ "$dry_run" -eq 1 ]; then
  echo "[dry-run] docker pull $API_IMAGE"
  echo "[dry-run] docker pull $WEB_IMAGE"
else
  echo "Descargando imágenes de GHCR..."
  docker pull "$API_IMAGE"
  docker pull "$WEB_IMAGE"
fi

# Pasos de seguridad previos en producción (TODO antes de tocar nada)
conteos_antes=""
if [ "$AMBIENTE" = "produccion" ]; then
  echo "--- Verificaciones y respaldos previos de producción ---"
  mkdir -p "$CAMPUS_HOME/state"
  conteos_antes="$CAMPUS_HOME/state/conteos-antes.tsv"

  # (i) Conteos «antes»
  if [ "$dry_run" -eq 1 ]; then
    echo "[dry-run] Tomar conteos «antes» con $ROOT/deploy/conteos.sql en contenedor db"
  else
    db_container="campus-produccion-db"
    if ! docker ps --format '{{.Names}}' 2>/dev/null | grep -Eq "^campus-produccion-db$"; then
      legacy_db="$(docker compose -p "$LEGACY_PROJECT" -f "$LEGACY_COMPOSE" ps -q db 2>/dev/null || true)"
      if [ -n "$legacy_db" ]; then
        db_container="$legacy_db"
      fi
    fi
    echo "Tomando conteos «antes» en base de datos ($db_container)..."
    docker exec -i "$db_container" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -A -F $'\t' -P footer=off < "$ROOT/deploy/conteos.sql" > "$conteos_antes"
    echo "Conteos «antes» guardados en $conteos_antes"
  fi

  # (ii) pg_dump -Fc
  mkdir -p "${CAMPUS_DATA_DIR}/backups"
  ts="$(date -u +%Y%m%dT%H%M%SZ)"
  dump_prefix="pre-deploy"
  if [ -n "$rollback" ]; then
    dump_prefix="pre-rollback"
  fi
  dump_file="${CAMPUS_DATA_DIR}/backups/${dump_prefix}-${ts}.dump"

  if [ "$dry_run" -eq 1 ]; then
    echo "[dry-run] pg_dump -Fc a $dump_file"
  else
    echo "Generando backup lógico con pg_dump en $dump_file..."
    docker exec -i "$db_container" pg_dump -U "$POSTGRES_USER" -Fc --no-owner --no-acl "$POSTGRES_DB" > "$dump_file"
    echo "Backup pg_dump generado exitosamente."
  fi

  # (iii) Snapshot EBS mediante IMDSv2
  if [ "$dry_run" -eq 1 ]; then
    echo "[dry-run] Obtener metadatos por IMDSv2 y crear snapshot EBS del volumen campus-verde-data"
  else
    echo "Obteniendo metadatos de instancia por IMDSv2..."
    imds_token="$(curl -sS --max-time 3 -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600" 2>/dev/null || true)"
    instance_id=""
    region=""
    if [ -n "$imds_token" ]; then
      instance_id="$(curl -sS --max-time 3 -H "X-aws-ec2-metadata-token: $imds_token" "http://169.254.169.254/latest/meta-data/instance-id" 2>/dev/null || true)"
      region="$(curl -sS --max-time 3 -H "X-aws-ec2-metadata-token: $imds_token" "http://169.254.169.254/latest/meta-data/placement/region" 2>/dev/null || true)"
    fi
    if [ -z "$region" ]; then
      region="${AWS_DEFAULT_REGION:-${AWS_REGION:-us-east-1}}"
    fi

    snap_ok=0
    if [ -n "$instance_id" ] && [ -n "$region" ]; then
      echo "Consultando volumen campus-verde-data adjunto a $instance_id en $region..."
      vol_id="$(aws ec2 describe-volumes --region "$region" \
        --filters "Name=attachment.instance-id,Values=${instance_id}" "Name=tag:Name,Values=campus-verde-data" \
        --query 'Volumes[0].VolumeId' --output text 2>/dev/null || true)"
      vol_id="$(printf '%s' "$vol_id" | tr -d '[:space:]')"

      if [ -n "$vol_id" ] && [ "$vol_id" != "None" ] && [ "$vol_id" != "null" ]; then
        snap_desc="${dump_prefix} ${AMBIENTE} ${target_tag}"
        snap_name="${dump_prefix}-${AMBIENTE}"
        echo "Creando snapshot EBS para el volumen $vol_id..."
        snap_id="$(aws ec2 create-snapshot --region "$region" --volume-id "$vol_id" \
          --description "$snap_desc" \
          --tag-specifications "ResourceType=snapshot,Tags=[{Key=Name,Value=${snap_name}},{Key=Ambiente,Value=${AMBIENTE}},{Key=GitSha,Value=${target_tag}}]" \
          --query 'SnapshotId' --output text 2>/dev/null || true)"
        snap_id="$(printf '%s' "$snap_id" | tr -d '[:space:]')"
        if [ -n "$snap_id" ] && [ "$snap_id" != "None" ] && [ "$snap_id" != "null" ]; then
          snap_ok=1
          echo "Snapshot EBS creado exitosamente: $snap_id (volumen $vol_id)"
        fi
      fi
    fi

    if [ "$snap_ok" -eq 0 ]; then
      if [ "${DEPLOY_SIN_SNAPSHOT:-0}" = "1" ]; then
        echo "ADVERTENCIA: No se pudo crear snapshot EBS pero DEPLOY_SIN_SNAPSHOT=1 está activo. Continuando."
      else
        echo "ERROR: No se pudo crear el snapshot EBS del volumen campus-verde-data en producción." >&2
        echo "El despliegue en producción requiere snapshot previo por seguridad. Si desea omitir, declare DEPLOY_SIN_SNAPSHOT=1." >&2
        exit 1
      fi
    fi
  fi
fi

# Guardar imágenes en uso para rollback
mkdir -p "$CAMPUS_HOME/state"
prev_api=""
prev_web=""
if [ "$dry_run" -eq 1 ]; then
  echo "[dry-run] Guardar imágenes en uso en $CAMPUS_HOME/state/prev-images.env"
else
  prev_api="$(docker inspect --format '{{.Config.Image}}' "campus-${AMBIENTE}-api" 2>/dev/null || true)"
  prev_web="$(docker inspect --format '{{.Config.Image}}' "campus-${AMBIENTE}-web" 2>/dev/null || true)"
  if [ -z "$prev_api" ] && [ "$AMBIENTE" = "produccion" ]; then
    prev_api="$(docker inspect --format '{{.Config.Image}}' campus-api-1 2>/dev/null || docker inspect --format '{{.Config.Image}}' campus-api 2>/dev/null || true)"
    prev_web="$(docker inspect --format '{{.Config.Image}}' campus-web-1 2>/dev/null || docker inspect --format '{{.Config.Image}}' campus-web 2>/dev/null || true)"
  fi
  cat > "$CAMPUS_HOME/state/prev-images.env" <<EOF
PREV_API_IMAGE=${prev_api}
PREV_WEB_IMAGE=${prev_web}
EOF
  echo "Imágenes previas registradas en $CAMPUS_HOME/state/prev-images.env"
fi

# Migración de stack legado (solo producción, una vez)
migrating_from_legacy=0
if [ "$AMBIENTE" = "produccion" ]; then
  legacy_containers=""
  if command -v docker >/dev/null 2>&1; then
    legacy_containers="$(docker ps -q --filter "label=com.docker.compose.project=${LEGACY_PROJECT}" 2>/dev/null || true)"
  fi
  if [ -n "$legacy_containers" ] && [ -f "$LEGACY_COMPOSE" ]; then
    migrating_from_legacy=1
    echo "Detectado stack legado en ejecución (proyecto $LEGACY_PROJECT). Deteniendo con 'docker compose stop'..."
    if [ "$dry_run" -eq 1 ]; then
      echo "[dry-run] docker compose -p $LEGACY_PROJECT -f $LEGACY_COMPOSE stop"
    else
      docker compose -p "$LEGACY_PROJECT" -f "$LEGACY_COMPOSE" stop
      if [ ! -f "$CAMPUS_HOME/docker-compose.legacy.yml" ]; then
        cp "$LEGACY_COMPOSE" "$CAMPUS_HOME/docker-compose.legacy.yml"
        echo "Copia de respaldo del compose legado conservada en $CAMPUS_HOME/docker-compose.legacy.yml"
      fi
    fi
  fi
fi

# Asegurar permisos de directorio de datos app (UID 10001)
if [ "$dry_run" -eq 1 ]; then
  echo "[dry-run] mkdir -p ${CAMPUS_DATA_DIR}/app ${CAMPUS_DATA_DIR}/pg"
  echo "[dry-run] chown -R 10001:10001 ${CAMPUS_DATA_DIR}/app"
else
  mkdir -p "${CAMPUS_DATA_DIR}/app" "${CAMPUS_DATA_DIR}/pg"
  if ! chown -R 10001:10001 "${CAMPUS_DATA_DIR}/app" 2>/dev/null; then
    if command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
      sudo chown -R 10001:10001 "${CAMPUS_DATA_DIR}/app"
    else
      docker run --rm -v "${CAMPUS_DATA_DIR}/app:/data" alpine:latest chown -R 10001:10001 /data 2>/dev/null || \
        chown -R 10001:10001 "${CAMPUS_DATA_DIR}/app" || true
    fi
  fi
fi

# TLS solo si los dos PEM ya están en el host. Sin ellos el compose es el de siempre.
tls_activo=0
if [ -f "${CAMPUS_CERTS_DIR}/fullchain.pem" ] && [ -f "${CAMPUS_CERTS_DIR}/privkey.pem" ]; then
  tls_activo=1
  echo "TLS: PEM presentes en ${CAMPUS_CERTS_DIR}; se incluye deploy/compose.tls.yml"
  if [ "$CAMPUS_COOKIE_SECURE" != "true" ]; then
    echo "AVISO: hay PEM pero CAMPUS_COOKIE_SECURE no es true en host.env. El compose de TLS igual envía true al contenedor de la API; deje el valor en true para que coincida con el archivo." >&2
  fi
  if [[ "$PUBLIC_URL" != https://* ]]; then
    echo "AVISO: hay PEM pero PUBLIC_URL no empieza por https (${PUBLIC_URL})." >&2
  fi
else
  echo "TLS: sin par PEM en ${CAMPUS_CERTS_DIR}; el despliegue sigue en HTTP."
fi

smoke_base="http://127.0.0.1:${WEB_PORT}"
smoke_resolve=""
smoke_insecure=0
if [ "$tls_activo" -eq 1 ]; then
  if [ "${SMOKE_INSECURE_TLS:-0}" = "1" ]; then
    smoke_base="https://127.0.0.1:${WEB_TLS_PORT}"
    smoke_insecure=1
    echo "Smoke HTTPS con SMOKE_INSECURE_TLS=1 (certificado de prueba) en ${smoke_base}"
  else
    if [ "$WEB_TLS_PORT" = "443" ]; then
      smoke_base="https://verde-pucp.duckdns.org"
    else
      smoke_base="https://verde-pucp.duckdns.org:${WEB_TLS_PORT}"
    fi
    smoke_resolve="verde-pucp.duckdns.org:${WEB_TLS_PORT}:127.0.0.1"
    echo "Smoke HTTPS de ${smoke_base} resuelto a 127.0.0.1:${WEB_TLS_PORT}"
  fi
fi

# Despliegue con docker compose
compose_cmd=(
  docker compose
  -f "$ROOT/deploy/compose.yml"
  -f "$ROOT/deploy/compose.${AMBIENTE}.yml"
  -f "$ROOT/deploy/compose.host.yml"
)
if [ "$tls_activo" -eq 1 ]; then
  compose_cmd+=(-f "$ROOT/deploy/compose.tls.yml")
fi
compose_cmd+=(
  -p "campus-${AMBIENTE}"
  --env-file "$HOST_ENV"
)

if [ "$dry_run" -eq 1 ]; then
  echo "[dry-run] ${compose_cmd[*]} up -d --no-build --remove-orphans"
  echo "[dry-run] Esperar healthy en campus-${AMBIENTE}-db"
  echo "[dry-run] Ejecutar smoke test deploy/smoke.sh $AMBIENTE en ${smoke_base}"
  if [ "$AMBIENTE" = "produccion" ]; then
    echo "[dry-run] Conteos después con deploy/conteos.sql y comparar_conteos.py"
  fi
  echo "[dry-run] Registrar estado en $CAMPUS_HOME/state/deployed.env"
  echo "Dry-run completado sin errores ni modificaciones."
  exit 0
fi

echo "Levantando servicios del nuevo stack..."
"${compose_cmd[@]}" up -d --no-build --remove-orphans

echo "Esperando a que campus-${AMBIENTE}-db esté en estado healthy..."
healthy=0
intento=0
max_intentos=45
while [ "$intento" -lt "$max_intentos" ]; do
  intento=$((intento + 1))
  status="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "campus-${AMBIENTE}-db" 2>/dev/null || echo "starting")"
  if [ "$status" = "healthy" ]; then
    healthy=1
    break
  fi
  sleep 2
done

if [ "$healthy" -eq 0 ]; then
  echo "ERROR: El contenedor campus-${AMBIENTE}-db no alcanzó estado healthy (último estado: $status)." >&2
  "${compose_cmd[@]}" logs --tail 80 db >&2 || true
  exit 1
fi
echo "Servicio de base de datos healthy."

# Smoke test
echo "Ejecutando smoke test post-despliegue..."
smoke_passed=1
smoke_env=(
  SMOKE_ENV_FILE="$HOST_ENV"
  SMOKE_BASE_URL="$smoke_base"
  CAMPUS_DEV_PASSWORD="$CAMPUS_DEV_PASSWORD"
)
if [ -n "$smoke_resolve" ]; then
  smoke_env+=(SMOKE_TLS_RESOLVE="$smoke_resolve")
fi
if [ "$smoke_insecure" -eq 1 ]; then
  smoke_env+=(SMOKE_INSECURE_TLS=1)
fi
if ! env "${smoke_env[@]}" bash "$ROOT/deploy/smoke.sh" "$AMBIENTE"; then
  smoke_passed=0
fi

if [ "$smoke_passed" -eq 0 ]; then
  echo "ERROR: El smoke test falló tras el despliegue. Iniciando rollback automático..." >&2
  if [ "$migrating_from_legacy" -eq 1 ]; then
    echo "Revirtiendo migración inicial: deteniendo campus-${AMBIENTE} y reiniciando stack legado $LEGACY_PROJECT..." >&2
    "${compose_cmd[@]}" stop || true
    docker compose -p "$LEGACY_PROJECT" -f "$LEGACY_COMPOSE" start || docker compose -p "$LEGACY_PROJECT" -f "$LEGACY_COMPOSE" up -d || true
    echo "Stack legado reiniciado tras fallo." >&2
  elif [ -n "$prev_api" ] && [ -n "$prev_web" ]; then
    echo "Restaurando imágenes previas: api=$prev_api web=$prev_web..." >&2
    API_IMAGE="$prev_api" WEB_IMAGE="$prev_web" "${compose_cmd[@]}" up -d --no-build --remove-orphans
    echo "Rollback automático a imágenes previas completado." >&2
  else
    echo "No hay imágenes previas registradas para realizar rollback automático." >&2
    "${compose_cmd[@]}" logs --tail 80 api >&2 || true
  fi
  exit 1
fi
echo "Smoke test verificado con éxito."

# Conteos después y comparación en producción
if [ "$AMBIENTE" = "produccion" ]; then
  conteos_despues="$CAMPUS_HOME/state/conteos-despues.tsv"
  echo "Tomando conteos «después» en campus-produccion-db..."
  docker exec -i "campus-produccion-db" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -A -F $'\t' -P footer=off < "$ROOT/deploy/conteos.sql" > "$conteos_despues"
  echo "Comparando conteos antes vs después con comparar_conteos.py..."
  if ! python3 "$ROOT/deploy/comparar_conteos.py" "$conteos_antes" "$conteos_despues" --excluir "$ROOT/deploy/conteos.excluir"; then
    echo "ERROR: Discrepancia detectada en tablas de negocio entre antes y después del despliegue." >&2
    exit 1
  fi
  echo "Verificación de integridad de datos de negocio completada exitosamente."
fi

# Registrar estado del despliegue
now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
deployed_file="$CAMPUS_HOME/state/deployed.env"
cat > "$deployed_file" <<EOF
DEPLOYED_SHA=${target_tag}
DEPLOYED_AT=${now}
DEPLOYED_ENV=${AMBIENTE}
DEPLOYED_API_IMAGE=${API_IMAGE}
DEPLOYED_WEB_IMAGE=${WEB_IMAGE}
PUBLIC_URL=${PUBLIC_URL}
EOF
echo "Estado guardado en $deployed_file"

if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
  {
    echo "### Despliegue en $AMBIENTE: EXITOSO"
    echo ""
    echo "- **SHA / Tag**: \`$target_tag\`"
    echo "- **Fecha**: \`$now\`"
    echo "- **Ambiente**: \`$AMBIENTE\`"
    echo "- **API Image**: \`$API_IMAGE\`"
    echo "- **Web Image**: \`$WEB_IMAGE\`"
    echo "- **URL**: \`$PUBLIC_URL\`"
  } >> "$GITHUB_STEP_SUMMARY"
fi

echo "=== Despliegue de $AMBIENTE finalizado exitosamente ==="
echo "Tag desplegado: $target_tag"
echo "URL disponible: $PUBLIC_URL"
