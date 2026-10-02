#!/usr/bin/env bash
# Despliega el ambiente produccion en un AWS Academy Learner Lab.
# Requiere AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY y AWS_SESSION_TOKEN
# de AWS Details > AWS CLI > Show. Caducan al cerrar la sesión del lab.
# El procedimiento (backup, snapshot, conteos, smoke y rollback) está en deploy/deploy.sh.
set -euo pipefail

: "${AWS_ACCESS_KEY_ID:?falta AWS_ACCESS_KEY_ID}"
: "${AWS_SECRET_ACCESS_KEY:?falta AWS_SECRET_ACCESS_KEY}"
: "${AWS_SESSION_TOKEN:?falta AWS_SESSION_TOKEN}"
: "${TF_VAR_db_password:?defina TF_VAR_db_password (16+ caracteres, no la clave de laboratorio)}"
: "${TF_VAR_dev_password:?defina TF_VAR_dev_password (16+ caracteres, no la clave de laboratorio)}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-us-east-1}"
export AWS_REGION="$AWS_DEFAULT_REGION"
export DEPLOY_AWS_CONFIRM=1

exec "$ROOT/deploy/deploy.sh" produccion --aws "$@"
