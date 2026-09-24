#!/usr/bin/env bash
# Despliega Campus Verde en un AWS Academy Learner Lab.
# Requiere AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY y AWS_SESSION_TOKEN
# de AWS Details > AWS CLI > Show. Caducan al cerrar la sesión del lab.
set -euo pipefail

: "${AWS_ACCESS_KEY_ID:?falta AWS_ACCESS_KEY_ID}"
: "${AWS_SECRET_ACCESS_KEY:?falta AWS_SECRET_ACCESS_KEY}"
: "${AWS_SESSION_TOKEN:?falta AWS_SESSION_TOKEN}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-us-east-1}"
export AWS_REGION="$AWS_DEFAULT_REGION"

cd "$ROOT/infra/terraform"
terraform init -input=false
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

aws ssm send-command \
  --instance-ids "$INSTANCE" \
  --document-name AWS-RunShellScript \
  --comment "campus verde compose" \
  --parameters 'commands=["systemctl restart campus.service"]' \
  >/tmp/campus-ssm.json || {
    echo "SSM no respondió. Si tiene la clave SSH: ssh ec2-user@$(terraform output -raw public_ip) 'systemctl restart campus.service'"
  }

echo "Listo: $URL"
echo "Cuando cierre el lab la instancia se detiene. Para no gastar el saldo: terraform destroy -auto-approve"
