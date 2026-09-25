#!/usr/bin/env bash
# Despliega Campus Verde en un AWS Academy Learner Lab.
# Requiere AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY y AWS_SESSION_TOKEN
# de AWS Details > AWS CLI > Show. Caducan al cerrar la sesión del lab.
set -euo pipefail

: "${AWS_ACCESS_KEY_ID:?falta AWS_ACCESS_KEY_ID}"
: "${AWS_SECRET_ACCESS_KEY:?falta AWS_SECRET_ACCESS_KEY}"
: "${AWS_SESSION_TOKEN:?falta AWS_SESSION_TOKEN}"
: "${TF_VAR_db_password:?defina TF_VAR_db_password (16+ caracteres, no la clave de laboratorio)}"
: "${TF_VAR_dev_password:?defina TF_VAR_dev_password (16+ caracteres, no la clave de laboratorio)}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-us-east-1}"
export AWS_REGION="$AWS_DEFAULT_REGION"

cd "$ROOT/infra/terraform"
terraform init -input=false -reconfigure
terraform apply -input=false -auto-approve

API_REPO="$(terraform output -raw ecr_api)"
WEB_REPO="$(terraform output -raw ecr_web)"
INSTANCE="$(terraform output -raw instance_id)"
URL="$(terraform output -raw public_url)"

if [ -z "$API_REPO" ] || [ -z "$WEB_REPO" ]; then
  echo "ECR no está creado. Revise create_ecr."
  exit 1
fi

aws ecr get-login-password --region "$AWS_DEFAULT_REGION" \
  | docker login --username AWS --password-stdin "${API_REPO%%/*}"

docker build -f "$ROOT/apps/api/Dockerfile" -t "$API_REPO:latest" "$ROOT"
docker build -f "$ROOT/apps/web/Dockerfile" -t "$WEB_REPO:latest" "$ROOT"
docker push "$API_REPO:latest"
docker push "$WEB_REPO:latest"

bash "$ROOT/scripts/poner-secretos.sh" "$INSTANCE"

echo "URL: $URL"

# Reinicio por SSM con LabInstanceProfile. El puerto 22 queda cerrado y no hay llave SSH.
if ! aws ssm send-command \
  --instance-ids "$INSTANCE" \
  --document-name AWS-RunShellScript \
  --comment "campus verde compose" \
  --parameters 'commands=["systemctl restart campus.service"]' \
  --region "$AWS_DEFAULT_REGION" \
  >/tmp/campus-ssm.json; then
  echo "SSM no aceptó el reinicio de campus.service. La URL de esta sesión es: $URL" >&2
  exit 1
fi

echo "Listo: $URL"
echo "Cuando cierre el lab la instancia se detiene. Para no gastar el saldo: terraform destroy -auto-approve"
