#!/usr/bin/env bash
# Despliegue por ambiente.
#   deploy/deploy.sh <develop|qa|produccion> --local
#   deploy/deploy.sh <develop|qa|produccion> --local --down
#   deploy/deploy.sh <develop|qa|produccion> --local --smoke
#   deploy/deploy.sh <develop|qa|produccion> --local --rollback
#   DEPLOY_AWS_CONFIRM=1 deploy/deploy.sh <ambiente> --aws [--sha SHA] [--rollback TAG]
#
# --aws habla con el Learner Lab (Terraform, ECR, SSM). No es el default.
# --local solo usa Docker en esta máquina. Los tres ambientes pueden convivir.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

usage() {
  echo "Uso: deploy/deploy.sh <develop|qa|produccion> --local|--aws [--down|--smoke|--rollback [TAG]] [--sha SHA]" >&2
  exit 2
}

AMBIENTE="${1:-}"
case "$AMBIENTE" in
  develop|qa|produccion) ;;
  *) usage ;;
esac
shift

mode=""
action="up"
rollback=""
sha="${GIT_SHA:-}"

while [ $# -gt 0 ]; do
  case "$1" in
    --local) mode="local" ;;
    --aws) mode="aws" ;;
    --down) action="down" ;;
    --smoke) action="smoke" ;;
    --rollback)
      action="rollback"
      if [ "${2:-}" != "" ] && [ "${2#--}" = "$2" ]; then
        rollback="$2"
        shift
      fi
      ;;
    --sha)
      shift
      sha="${1:-}"
      ;;
    -h|--help) usage ;;
    *)
      echo "opción desconocida: $1" >&2
      usage
      ;;
  esac
  shift
done

if [ -z "$mode" ]; then
  mode="local"
fi

env_file() {
  local dest="$ROOT/deploy/env/${AMBIENTE}.env"
  if [ ! -f "$dest" ]; then
    cp "$ROOT/deploy/env/${AMBIENTE}.env.example" "$dest"
    chmod 600 "$dest"
    echo "creado $dest desde el example (claves locales de demostración)" >&2
  fi
  printf '%s' "$dest"
}

load_env() {
  local file
  file="$(env_file)"
  local env_postgres_port="${POSTGRES_PORT:-}"
  local env_api_port="${API_PORT:-}"
  local env_web_port="${WEB_PORT:-}"

  set -a
  # shellcheck disable=SC1090
  source "$file"
  set +a

  if [ -n "$env_postgres_port" ]; then
    POSTGRES_PORT="$env_postgres_port"
  fi
  if [ -n "$env_api_port" ]; then
    API_PORT="$env_api_port"
  fi
  if [ -n "$env_web_port" ]; then
    WEB_PORT="$env_web_port"
  fi
  export POSTGRES_PORT API_PORT WEB_PORT

  if [ "${APP_ENV:-}" != "$AMBIENTE" ]; then
    echo "APP_ENV=${APP_ENV:-} en el env no coincide con $AMBIENTE" >&2
    exit 1
  fi
}

compose() {
  POSTGRES_PORT="$POSTGRES_PORT" API_PORT="$API_PORT" WEB_PORT="$WEB_PORT" docker compose \
    -f "$ROOT/deploy/compose.yml" \
    -f "$ROOT/deploy/compose.${AMBIENTE}.yml" \
    --env-file "$ROOT/deploy/env/${AMBIENTE}.env" \
    "$@"
}

state_dir() {
  mkdir -p "$ROOT/deploy/state"
  chmod 700 "$ROOT/deploy/state"
}

image_api() {
  printf 'campus-verde-api:%s' "$AMBIENTE"
}

image_web() {
  printf 'campus-verde-web:%s' "$AMBIENTE"
}

record_previous() {
  state_dir
  local current prev_file current_web prev_web_file
  prev_file="$ROOT/deploy/state/${AMBIENTE}.prev-image"
  prev_web_file="$ROOT/deploy/state/${AMBIENTE}.prev-image-web"
  current="$(docker image inspect "$(image_api)" --format '{{.Id}}' 2>/dev/null || true)"
  if [ -n "$current" ]; then
    printf '%s\n' "$current" >"$prev_file"
    echo "imagen anterior de api: $current"
  fi
  current_web="$(docker image inspect "$(image_web)" --format '{{.Id}}' 2>/dev/null || true)"
  if [ -n "$current_web" ]; then
    printf '%s\n' "$current_web" >"$prev_web_file"
    echo "imagen anterior de web: $current_web"
  fi
}

rollback_local() {
  state_dir
  local prev_file prev prev_web_file prev_web
  prev_file="$ROOT/deploy/state/${AMBIENTE}.prev-image"
  prev_web_file="$ROOT/deploy/state/${AMBIENTE}.prev-image-web"
  prev="${rollback:-}"
  if [ -z "$prev" ] && [ -f "$prev_file" ]; then
    prev="$(cat "$prev_file")"
  fi
  if [ -z "$prev" ]; then
    echo "no hay imagen anterior de api para rollback de $AMBIENTE" >&2
    exit 1
  fi
  if ! docker image inspect "$prev" >/dev/null 2>&1; then
    echo "no existe la imagen de api $prev" >&2
    exit 1
  fi

  prev_web=""
  if [ -f "$prev_web_file" ]; then
    prev_web="$(cat "$prev_web_file")"
  fi
  if [ -z "$prev_web" ]; then
    echo "no hay imagen anterior de web para rollback de $AMBIENTE" >&2
    exit 1
  fi
  if ! docker image inspect "$prev_web" >/dev/null 2>&1; then
    echo "no existe la imagen de web $prev_web" >&2
    exit 1
  fi

  docker tag "$prev" "$(image_api)"
  docker tag "$prev_web" "$(image_web)"
  compose up -d --no-build --remove-orphans
  echo "rollback local de $AMBIENTE a api=$prev web=$prev_web"
  SMOKE_ENV_FILE="$ROOT/deploy/env/${AMBIENTE}.env" bash "$ROOT/deploy/smoke.sh" "$AMBIENTE"
}

up_local() {
  load_env
  record_previous
  compose up -d --build --remove-orphans
  if ! SMOKE_ENV_FILE="$ROOT/deploy/env/${AMBIENTE}.env" bash "$ROOT/deploy/smoke.sh" "$AMBIENTE"; then
    echo "smoke falló. Se intenta rollback a la imagen anterior." >&2
    if [ -f "$ROOT/deploy/state/${AMBIENTE}.prev-image" ] || [ -f "$ROOT/deploy/state/${AMBIENTE}.prev-image-web" ]; then
      rollback=""
      rollback_local
      exit 1
    fi
    compose logs --tail 80 api >&2 || true
    exit 1
  fi
}

tag_valido() {
  local tag="$1"
  [[ "$tag" =~ ^[0-9a-f]{40}$ || "$tag" =~ ^rc-[A-Za-z0-9._-]+$ ]]
}

aws_deploy() {
  if [ "${DEPLOY_AWS_CONFIRM:-}" != "1" ]; then
    echo "Refusing --aws sin DEPLOY_AWS_CONFIRM=1. El workflow y deploy-learner-lab.sh lo exportan." >&2
    exit 1
  fi
  : "${AWS_ACCESS_KEY_ID:?falta AWS_ACCESS_KEY_ID}"
  : "${AWS_SECRET_ACCESS_KEY:?falta AWS_SECRET_ACCESS_KEY}"
  : "${AWS_SESSION_TOKEN:?falta AWS_SESSION_TOKEN}"
  : "${TF_VAR_db_password:?falta TF_VAR_db_password}"
  : "${TF_VAR_dev_password:?falta TF_VAR_dev_password}"

  export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-${AWS_REGION:-us-east-1}}"
  export AWS_REGION="$AWS_DEFAULT_REGION"

  if [ -z "$sha" ]; then
    sha="$(git -C "$ROOT" rev-parse HEAD)"
  fi
  if ! [[ "$sha" =~ ^[0-9a-f]{40}$ ]]; then
    echo "el SHA de despliegue debe ser 40 hex: $sha" >&2
    exit 1
  fi

  local target_tag="$sha"
  if [ "$action" = "rollback" ]; then
    if [ -z "$rollback" ]; then
      echo "rollback AWS necesita el tag (SHA o rc-*)" >&2
      exit 1
    fi
    if ! tag_valido "$rollback"; then
      echo "tag de rollback no permitido: $rollback" >&2
      exit 1
    fi
    target_tag="$rollback"
  fi

  local tf="$ROOT/infra/terraform"
  terraform -chdir="$tf" init -input=false -reconfigure
  if terraform -chdir="$tf" workspace list | grep -Eq "(^|[[:space:]])${AMBIENTE}$"; then
    terraform -chdir="$tf" workspace select "$AMBIENTE"
  elif [ "$AMBIENTE" = "produccion" ]; then
    echo "workspace produccion no existe. Se usa el workspace actual con -var ambiente=produccion."
  else
    echo "No existe el workspace $AMBIENTE. Créelo a mano cuando corresponda: terraform workspace new $AMBIENTE" >&2
    exit 1
  fi

  local antes="" despues=""

  # Orden seguro en producción (F-B):
  # 1. terraform init / workspace select
  # 2. ANTES de terraform apply (sobre la instancia existente en el estado):
  #    - Snapshot EBS del volumen de datos (campus-verde-data)
  #    - Backup lógico pg_dump en la instancia (/opt/campus/data/backups/)
  #    - Conteos «antes» de tablas
  #    (si no hay instancia previa en producción, falla salvo DEPLOY_PRIMERA_VEZ=1)
  # 3. terraform apply
  # 4. Construcción / subida de imágenes a ECR
  # 5. Inyección de secretos en la instancia (poner-secretos.sh)
  # 6. Publicación de imágenes en la instancia (patch_compose + compose pull/up)
  # 7. Smoke test (con rollback automático al tag anterior si falla)
  # 8. Conteos «después» y comparación de tablas de negocio (comparar_conteos.py)

  if [ "$AMBIENTE" = "produccion" ]; then
    local pre_instance
    pre_instance="$(terraform -chdir="$tf" output -raw instance_id 2>/dev/null || true)"
    pre_instance="$(printf '%s' "$pre_instance" | tr -d '[:space:]')"
    if [ -z "$pre_instance" ] || [ "$pre_instance" = "None" ] || [ "$pre_instance" = "null" ]; then
      if [ "${DEPLOY_PRIMERA_VEZ:-}" = "1" ]; then
        echo "Producción sin instancia previa en el estado de Terraform (DEPLOY_PRIMERA_VEZ=1). Se omite snapshot, backup y conteos antes."
      else
        echo "No hay instancia en el estado de Terraform para producción. No se puede tomar snapshot EBS ni backup antes de apply. Si es la primera creación de la infraestructura, exporte DEPLOY_PRIMERA_VEZ=1." >&2
        exit 1
      fi
    else
      echo "Tomando snapshot EBS, pg_dump y conteos antes sobre la instancia existente: $pre_instance"
      backup_produccion "$pre_instance" "$target_tag"
      antes="$(mktemp)"
      conteos_remotos "$pre_instance" >"$antes"
      echo "conteos antes: $antes"
    fi
  fi

  terraform -chdir="$tf" apply -input=false -auto-approve -var "ambiente=${AMBIENTE}"

  local api_repo web_repo instance url
  api_repo="$(terraform -chdir="$tf" output -raw ecr_api)"
  web_repo="$(terraform -chdir="$tf" output -raw ecr_web)"
  instance="$(terraform -chdir="$tf" output -raw instance_id)"
  url="$(terraform -chdir="$tf" output -raw public_url)"
  if [ -z "$api_repo" ] || [ -z "$web_repo" ] || [ -z "$instance" ]; then
    echo "Terraform no devolvió ECR o la instancia." >&2
    exit 1
  fi

  local owner_lc ghcr_api ghcr_web
  owner_lc="$(printf '%s' "${GITHUB_REPOSITORY_OWNER:-}" | tr '[:upper:]' '[:lower:]')"
  ghcr_api=""
  ghcr_web=""
  if [ -n "$owner_lc" ]; then
    ghcr_api="ghcr.io/${owner_lc}/campus-verde-api:${target_tag}"
    ghcr_web="ghcr.io/${owner_lc}/campus-verde-web:${target_tag}"
  fi

  aws ecr get-login-password --region "$AWS_DEFAULT_REGION" \
    | docker login --username AWS --password-stdin "${api_repo%%/*}"

  local pulled=0
  if [ -n "$ghcr_api" ] && docker pull "$ghcr_api" && docker pull "$ghcr_web"; then
    docker tag "$ghcr_api" "${api_repo}:${target_tag}"
    docker tag "$ghcr_web" "${web_repo}:${target_tag}"
    pulled=1
  elif docker pull "${api_repo}:${target_tag}" && docker pull "${web_repo}:${target_tag}"; then
    pulled=1
  fi
  if [ "$pulled" -eq 0 ]; then
    if [ "$action" = "rollback" ]; then
      echo "No está la imagen ${target_tag}. Un rollback no se reconstruye desde el checkout." >&2
      exit 1
    fi
    echo "No está la imagen ${target_tag} en GHCR ni en ECR. Se construye desde el checkout y se publica en ECR."
    docker build -f "$ROOT/backend/dockerfile" -t "${api_repo}:${target_tag}" "$ROOT"
    docker build -f "$ROOT/apps/web/Dockerfile" -t "${web_repo}:${target_tag}" "$ROOT"
  fi
  docker push "${api_repo}:${target_tag}"
  docker push "${web_repo}:${target_tag}"
  echo "Imagen API (tag inmutable): ${api_repo}:${target_tag}"

  if command -v poner-secretos >/dev/null 2>&1; then
    poner-secretos "$instance"
  else
    bash "$ROOT/scripts/poner-secretos.sh" "$instance"
  fi

  local previous
  previous="$(publicar_imagenes_en_instancia "$instance" "${api_repo}:${target_tag}" "${web_repo}:${target_tag}")"
  echo "$previous"

  if ! SMOKE_BASE_URL="$url" CAMPUS_DEV_PASSWORD="$TF_VAR_dev_password" \
      bash "$ROOT/deploy/smoke.sh" "$AMBIENTE"; then
    echo "smoke falló en $url. Rollback al tag anterior." >&2
    local prev_api prev_web
    prev_api="$(printf '%s\n' "$previous" | sed -n 's/^PREVIOUS_API=//p' | head -n 1)"
    prev_web="$(printf '%s\n' "$previous" | sed -n 's/^PREVIOUS_WEB=//p' | head -n 1)"
    if [ -n "$prev_api" ] && [ -n "$prev_web" ]; then
      publicar_imagenes_en_instancia "$instance" "$prev_api" "$prev_web" >/dev/null
      echo "rollback aplicado: $prev_api"
    fi
    exit 1
  fi

  if [ "$AMBIENTE" = "produccion" ]; then
    if [ -n "$antes" ] && [ -f "$antes" ]; then
      despues="$(mktemp)"
      conteos_remotos "$instance" >"$despues"
      python3 "$ROOT/deploy/comparar_conteos.py" "$antes" "$despues" --excluir "$ROOT/deploy/conteos.excluir"
      rm -f "$antes" "$despues"
    else
      echo "Primer despliegue (DEPLOY_PRIMERA_VEZ=1): se omiten conteos de comparación antes/después."
    fi
  fi

  echo "Listo: $url"
  echo "Tag desplegado: $target_tag"
  echo "Rollback: DEPLOY_AWS_CONFIRM=1 bash deploy/deploy.sh $AMBIENTE --aws --rollback <tag-anterior>"
}

backup_produccion() {
  local instance="$1"
  local tag="$2"
  local prefix="campus-verde"
  local vol snap
  vol="$(aws ec2 describe-volumes --region "$AWS_DEFAULT_REGION" \
    --filters "Name=tag:Name,Values=${prefix}-data" "Name=attachment.instance-id,Values=${instance}" \
    --query 'Volumes[0].VolumeId' --output text)"
  if [ -z "$vol" ] || [ "$vol" = "None" ]; then
    echo "No hay volumen ${prefix}-data en la instancia. Producción no se despliega sin snapshot." >&2
    exit 1
  fi
  snap="$(aws ec2 create-snapshot --region "$AWS_DEFAULT_REGION" --volume-id "$vol" \
    --description "pre-deploy ${AMBIENTE} ${tag}" \
    --tag-specifications "ResourceType=snapshot,Tags=[{Key=Name,Value=pre-deploy-${AMBIENTE}},{Key=Ambiente,Value=${AMBIENTE}},{Key=GitSha,Value=${tag}}]" \
    --query 'SnapshotId' --output text)"
  echo "snapshot EBS: $snap (volumen $vol)"

  local remoto
  remoto=$(cat <<'SH'
set -euo pipefail
mkdir -p /opt/campus/data/backups
ts=$(date -u +%Y%m%dT%H%M%SZ)
out="/opt/campus/data/backups/pre-deploy-${ts}.dump"
docker compose -f /opt/campus/docker-compose.yml exec -T db \
  pg_dump -U campus -Fc --no-owner --no-acl campus_verde > "$out"
ls -l "$out"
SH
)
  ssm_ok "$instance" "backup postgis" "$remoto"
  echo "backup pg_dump pedido en la instancia (pre-deploy)"
}

conteos_remotos() {
  local instance="$1"
  local sql
  sql="$(cat "$ROOT/deploy/conteos.sql")"
  local remoto
  remoto=$(cat <<SH
set -euo pipefail
docker compose -f /opt/campus/docker-compose.yml exec -T db \
  psql -U campus -d campus_verde -v ON_ERROR_STOP=1 -A -F \$'\\t' -P footer=off <<'SQL'
${sql}
SQL
SH
)
  ssm_stdout "$instance" "conteos" "$remoto"
}

publicar_imagenes_en_instancia() {
  local instance="$1"
  local api_image="$2"
  local web_image="$3"
  local b64
  b64="$(base64 -w0 "$ROOT/deploy/patch_compose.py" 2>/dev/null || base64 <"$ROOT/deploy/patch_compose.py" | tr -d '\n')"
  local remoto
  remoto=$(cat <<SH
set -euo pipefail
umask 077
printf '%s\n' '${b64}' | base64 -d > /tmp/patch_compose.py
python3 /tmp/patch_compose.py /opt/campus/docker-compose.yml --api '${api_image}' --web '${web_image}'
cd /opt/campus
aws ecr get-login-password --region ${AWS_DEFAULT_REGION} | docker login --username AWS --password-stdin ${api_image%%/*} || true
docker compose pull
docker compose up -d
SH
)
  ssm_stdout "$instance" "compose image ${api_image}" "$remoto"
}

ssm_send() {
  local instance="$1"
  local comment="$2"
  local script="$3"
  local params cmd_id
  params="$(python3 -c 'import json,sys; print(json.dumps({"commands":[sys.stdin.read()]}))' <<<"$script")"
  cmd_id="$(aws ssm send-command \
    --region "$AWS_DEFAULT_REGION" \
    --instance-ids "$instance" \
    --document-name AWS-RunShellScript \
    --comment "$comment" \
    --parameters "$params" \
    --query 'Command.CommandId' --output text)"
  local i status
  i=0
  while [ "$i" -lt 60 ]; do
    status="$(aws ssm get-command-invocation \
      --region "$AWS_DEFAULT_REGION" \
      --command-id "$cmd_id" \
      --instance-id "$instance" \
      --query 'Status' --output text 2>/dev/null || echo Pending)"
    case "$status" in
      Success|Failed|Cancelled|TimedOut) break ;;
    esac
    i=$((i + 1))
    sleep 5
  done
  printf '%s\n' "$cmd_id" "$status"
}

ssm_stdout() {
  local instance="$1"
  local comment="$2"
  local script="$3"
  local meta cmd_id status
  meta="$(ssm_send "$instance" "$comment" "$script")"
  cmd_id="$(printf '%s\n' "$meta" | sed -n '1p')"
  status="$(printf '%s\n' "$meta" | sed -n '2p')"
  aws ssm get-command-invocation \
    --region "$AWS_DEFAULT_REGION" \
    --command-id "$cmd_id" \
    --instance-id "$instance" \
    --query 'StandardOutputContent' --output text
  if [ "$status" != "Success" ]; then
    echo "SSM $comment terminó en $status" >&2
    aws ssm get-command-invocation \
      --region "$AWS_DEFAULT_REGION" \
      --command-id "$cmd_id" \
      --instance-id "$instance" \
      --query 'StandardErrorContent' --output text >&2 || true
    exit 1
  fi
}

ssm_ok() {
  ssm_stdout "$@" >/dev/null
}

case "$mode" in
  local)
    case "$action" in
      down)
        load_env
        compose down
        ;;
      smoke)
        load_env
        SMOKE_ENV_FILE="$ROOT/deploy/env/${AMBIENTE}.env" bash "$ROOT/deploy/smoke.sh" "$AMBIENTE"
        ;;
      rollback)
        load_env
        rollback_local
        ;;
      up)
        up_local
        ;;
    esac
    ;;
  aws)
    case "$action" in
      up|rollback) aws_deploy ;;
      *)
        echo "--aws no combina con --$action" >&2
        exit 2
        ;;
    esac
    ;;
  *)
    usage
    ;;
esac
