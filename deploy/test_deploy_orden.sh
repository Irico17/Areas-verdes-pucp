#!/usr/bin/env bash
# Arnés local para validar el orden de llamadas de deploy/deploy.sh --aws
# Sustituye terraform, aws, docker, poner-secretos y curl por stubs en un PATH temporal.
# No realiza ninguna llamada de red ni a AWS real.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

BIN_DIR="$TMPDIR/bin"
mkdir -p "$BIN_DIR"
export LOG_FILE="$TMPDIR/calls.log"
export STATE_INSTANCE_FILE="$TMPDIR/state_instance_id"
export LAST_SSM_FILE="$TMPDIR/last_ssm"
export CURRENT_ENV_FILE="$TMPDIR/current_env"

# Stub terraform
cat > "$BIN_DIR/terraform" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail
echo "TERRAFORM $*" >> "$LOG_FILE"
while [ $# -gt 0 ]; do
  case "$1" in
    init) exit 0 ;;
    workspace)
      if [ "${2:-}" = "list" ]; then
        echo "  default"
        echo "* produccion"
        echo "  develop"
        echo "  qa"
        exit 0
      elif [ "${2:-}" = "select" ]; then
        exit 0
      fi
      ;;
    apply)
      echo "ACTION_APPLY" >> "$LOG_FILE"
      exit 0
      ;;
    output)
      if [ "${2:-}" = "-raw" ]; then
        case "${3:-}" in
          instance_id)
            if [ -f "$STATE_INSTANCE_FILE" ] && [ -s "$STATE_INSTANCE_FILE" ]; then
              cat "$STATE_INSTANCE_FILE"
              exit 0
            else
              echo "None"
              exit 1
            fi
            ;;
          ecr_api)
            echo "123456789012.dkr.ecr.us-east-1.amazonaws.com/campus-verde-api"
            exit 0
            ;;
          ecr_web)
            echo "123456789012.dkr.ecr.us-east-1.amazonaws.com/campus-verde-web"
            exit 0
            ;;
          public_url)
            echo "http://198.51.100.1"
            exit 0
            ;;
        esac
      fi
      ;;
  esac
  shift
done
exit 0
EOF
chmod +x "$BIN_DIR/terraform"

# Stub aws
cat > "$BIN_DIR/aws" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail
echo "AWS $*" >> "$LOG_FILE"
case "$1" in
  ec2)
    case "$2" in
      describe-volumes)
        echo "vol-mock123456"
        exit 0
        ;;
      create-snapshot)
        echo "ACTION_SNAPSHOT_EBS" >> "$LOG_FILE"
        echo "snap-mock123456"
        exit 0
        ;;
    esac
    ;;
  ssm)
    case "$2" in
      send-command)
        comment=""
        shift 2
        while [ $# -gt 0 ]; do
          if [ "$1" = "--comment" ]; then
            comment="$2"
            shift 2
            continue
          fi
          shift
        done
        echo "$comment" > "$LAST_SSM_FILE"
        if [ "$comment" = "backup postgis" ]; then
          echo "ACTION_BACKUP_PGDUMP" >> "$LOG_FILE"
        elif [ "$comment" = "conteos" ]; then
          echo "ACTION_CONTEOS" >> "$LOG_FILE"
        elif [[ "$comment" == *"secrets.env"* ]]; then
          echo "ACTION_PONER_SECRETOS" >> "$LOG_FILE"
        elif [[ "$comment" == *"compose image"* ]]; then
          echo "ACTION_PUBLICAR_IMAGENES" >> "$LOG_FILE"
        fi
        echo "cmd-mock1234"
        exit 0
        ;;
      get-command-invocation)
        query=""
        shift 2
        while [ $# -gt 0 ]; do
          if [ "$1" = "--query" ]; then
            query="$2"
            shift 2
            continue
          fi
          shift
        done
        if [ "$query" = "Status" ]; then
          echo "Success"
          exit 0
        elif [ "$query" = "StandardOutputContent" ]; then
          last_comment="$(cat "$LAST_SSM_FILE" 2>/dev/null || true)"
          if [ "$last_comment" = "conteos" ]; then
            printf "table_name\tfilas\nareas_verdes\t521\nschema_migrations\t38\nsesiones\t1\nusuarios\t6\n"
          elif [[ "$last_comment" == *"compose image"* ]]; then
            printf "PREVIOUS_API=prev-api:123\nPREVIOUS_WEB=prev-web:123\n"
          else
            echo "mock ssm output ok"
          fi
          exit 0
        fi
        ;;
    esac
    ;;
  ecr)
    if [ "$2" = "get-login-password" ]; then
      echo "mock_ecr_password"
      exit 0
    fi
    ;;
esac
exit 0
EOF
chmod +x "$BIN_DIR/aws"

# Stub docker
cat > "$BIN_DIR/docker" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail
echo "DOCKER $*" >> "$LOG_FILE"
case "$1" in
  login|pull|tag|build) exit 0 ;;
  push)
    echo "ACTION_DOCKER_PUSH" >> "$LOG_FILE"
    exit 0
    ;;
esac
exit 0
EOF
chmod +x "$BIN_DIR/docker"

# Stub poner-secretos
cat > "$BIN_DIR/poner-secretos" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail
echo "ACTION_PONER_SECRETOS" >> "$LOG_FILE"
exit 0
EOF
chmod +x "$BIN_DIR/poner-secretos"

# Stub curl (simula health, sesion y swagger apagado en produccion)
cat > "$BIN_DIR/curl" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail
body_file=""
url=""
while [ $# -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    body_file="$2"
    shift 2
    continue
  elif [[ "$1" == http://* || "$1" == https://* ]]; then
    url="$1"
  fi
  shift
done

if [[ "$url" == *"/swagger"* ]]; then
  curr_amb="$(cat "${CURRENT_ENV_FILE:-}" 2>/dev/null || echo "produccion")"
  if [ "$curr_amb" = "produccion" ]; then
    echo "404"
  else
    echo "200"
  fi
  exit 0
fi

if [ -n "$body_file" ]; then
  if [[ "$url" == *"/health" ]]; then
    echo '{"status":"ok"}' > "$body_file"
  else
    echo '{"token":"mock"}' > "$body_file"
  fi
fi
echo "200"
exit 0
EOF
chmod +x "$BIN_DIR/curl"

export PATH="$BIN_DIR:$PATH"
export DEPLOY_AWS_CONFIRM=1
export AWS_ACCESS_KEY_ID=mock_key
export AWS_SECRET_ACCESS_KEY=mock_secret
export AWS_SESSION_TOKEN=mock_token
export TF_VAR_db_password=mock_db_password_16ch
export TF_VAR_dev_password=mock_dev_password_16ch

sha="0123456789012345678901234567890123456789"

echo "=== Prueba 1: Producción con instancia existente (orden seguro pre-apply) ==="
echo "produccion" > "$TMPDIR/current_env"
echo "i-mockinstance01" > "$STATE_INSTANCE_FILE"
: > "$LOG_FILE"

bash "$ROOT/deploy/deploy.sh" produccion --aws --sha "$sha"

# Verificar orden de acciones
snap_line="$(grep -n "ACTION_SNAPSHOT_EBS" "$LOG_FILE" | cut -d: -f1 | head -n1)"
backup_line="$(grep -n "ACTION_BACKUP_PGDUMP" "$LOG_FILE" | cut -d: -f1 | head -n1)"
conteos_antes_line="$(grep -n "ACTION_CONTEOS" "$LOG_FILE" | cut -d: -f1 | head -n1)"
apply_line="$(grep -n "ACTION_APPLY" "$LOG_FILE" | cut -d: -f1 | head -n1)"
push_line="$(grep -n "ACTION_DOCKER_PUSH" "$LOG_FILE" | cut -d: -f1 | head -n1)"
secretos_line="$(grep -n "ACTION_PONER_SECRETOS" "$LOG_FILE" | cut -d: -f1 | head -n1)"
conteos_despues_line="$(grep -n "ACTION_CONTEOS" "$LOG_FILE" | cut -d: -f1 | tail -n1)"

if [ -z "$snap_line" ] || [ -z "$backup_line" ] || [ -z "$conteos_antes_line" ] || [ -z "$apply_line" ]; then
  echo "ERROR: faltan llamadas requeridas en la Prueba 1" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi

if [ "$snap_line" -ge "$apply_line" ]; then
  echo "ERROR: Snapshot EBS ($snap_line) ocurrió después o durante apply ($apply_line)" >&2
  exit 1
fi

if [ "$backup_line" -ge "$apply_line" ]; then
  echo "ERROR: Backup pg_dump ($backup_line) ocurrió después o durante apply ($apply_line)" >&2
  exit 1
fi

if [ "$conteos_antes_line" -ge "$apply_line" ]; then
  echo "ERROR: Conteos antes ($conteos_antes_line) ocurrió después o durante apply ($apply_line)" >&2
  exit 1
fi

if [ "$apply_line" -ge "$push_line" ] || [ "$apply_line" -ge "$secretos_line" ]; then
  echo "ERROR: Push o secretos ocurrieron antes de apply" >&2
  exit 1
fi

if [ "$apply_line" -ge "$conteos_despues_line" ]; then
  echo "ERROR: Conteos después ocurrió antes de apply" >&2
  exit 1
fi

echo "Prueba 1 OK: snapshot, pg_dump y conteos antes ejecutados correctamente previo a apply."

echo "=== Prueba 2: Producción sin instancia en estado y sin DEPLOY_PRIMERA_VEZ (debe fallar) ==="
rm -f "$STATE_INSTANCE_FILE"
: > "$LOG_FILE"
out_err="$TMPDIR/err.log"

set +e
bash "$ROOT/deploy/deploy.sh" produccion --aws --sha "$sha" >"$out_err" 2>&1
rc=$?
set -e

if [ "$rc" -eq 0 ]; then
  echo "ERROR: Producción debió fallar sin instancia previa en el estado" >&2
  exit 1
fi

if ! grep -q "DEPLOY_PRIMERA_VEZ=1" "$out_err"; then
  echo "ERROR: El mensaje de error no menciona DEPLOY_PRIMERA_VEZ=1" >&2
  cat "$out_err" >&2
  exit 1
fi

if grep -q "ACTION_APPLY" "$LOG_FILE"; then
  echo "ERROR: Terraform apply no debió ejecutarse" >&2
  exit 1
fi

echo "Prueba 2 OK: falló de forma segura sin tocar apply."

echo "=== Prueba 3: Producción sin instancia en estado con DEPLOY_PRIMERA_VEZ=1 (debe pasar) ==="
echo "produccion" > "$TMPDIR/current_env"
rm -f "$STATE_INSTANCE_FILE"
: > "$LOG_FILE"

# Para la fase post-apply el estado ya tendrá instancia creada
cat > "$BIN_DIR/terraform" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail
echo "TERRAFORM $*" >> "$LOG_FILE"
while [ $# -gt 0 ]; do
  case "$1" in
    init) exit 0 ;;
    workspace)
      if [ "${2:-}" = "list" ]; then
        echo "  default"
        echo "* produccion"
        echo "  develop"
        echo "  qa"
        exit 0
      elif [ "${2:-}" = "select" ]; then
        exit 0
      fi
      ;;
    apply)
      echo "ACTION_APPLY" >> "$LOG_FILE"
      echo "i-newinstance99" > "$STATE_INSTANCE_FILE"
      exit 0
      ;;
    output)
      if [ "${2:-}" = "-raw" ]; then
        case "${3:-}" in
          instance_id)
            if [ -f "$STATE_INSTANCE_FILE" ] && [ -s "$STATE_INSTANCE_FILE" ]; then
              cat "$STATE_INSTANCE_FILE"
              exit 0
            else
              echo "None"
              exit 1
            fi
            ;;
          ecr_api)
            echo "123456789012.dkr.ecr.us-east-1.amazonaws.com/campus-verde-api"
            exit 0
            ;;
          ecr_web)
            echo "123456789012.dkr.ecr.us-east-1.amazonaws.com/campus-verde-web"
            exit 0
            ;;
          public_url)
            echo "http://198.51.100.1"
            exit 0
            ;;
        esac
      fi
      ;;
  esac
  shift
done
exit 0
EOF
chmod +x "$BIN_DIR/terraform"

DEPLOY_PRIMERA_VEZ=1 bash "$ROOT/deploy/deploy.sh" produccion --aws --sha "$sha"

if ! grep -q "ACTION_APPLY" "$LOG_FILE"; then
  echo "ERROR: Terraform apply debió ejecutarse en DEPLOY_PRIMERA_VEZ=1" >&2
  exit 1
fi

if grep -q "ACTION_SNAPSHOT_EBS" "$LOG_FILE" || grep -q "ACTION_BACKUP_PGDUMP" "$LOG_FILE"; then
  echo "ERROR: No debió ejecutarse snapshot previo en primera vez" >&2
  exit 1
fi

echo "Prueba 3 OK: primer despliegue con DEPLOY_PRIMERA_VEZ=1 funcionó sin snapshot previo."

echo "=== Prueba 4: Develop no exige snapshot previo ==="
echo "develop" > "$TMPDIR/current_env"
rm -f "$STATE_INSTANCE_FILE"
: > "$LOG_FILE"

bash "$ROOT/deploy/deploy.sh" develop --aws --sha "$sha"

if ! grep -q "ACTION_APPLY" "$LOG_FILE"; then
  echo "ERROR: Terraform apply debió ejecutarse para develop" >&2
  exit 1
fi

if grep -q "ACTION_SNAPSHOT_EBS" "$LOG_FILE"; then
  echo "ERROR: Develop no debió tomar snapshot EBS" >&2
  exit 1
fi

echo "Prueba 4 OK: develop no exige snapshot."

echo "test_deploy_orden ok"
