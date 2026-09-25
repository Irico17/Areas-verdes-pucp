#!/usr/bin/env bash
# Escribe /opt/campus/secrets.env en la EC2 por SSM. No pasa por Terraform ni por user data.
# Requiere AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN,
# TF_VAR_db_password, TF_VAR_dev_password y el instance_id (argumento o terraform output).
set -euo pipefail

: "${TF_VAR_db_password:?defina TF_VAR_db_password}"
: "${TF_VAR_dev_password:?defina TF_VAR_dev_password}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
region="${AWS_DEFAULT_REGION:-us-east-1}"
instance="${1:-}"
if [ -z "$instance" ]; then
  instance="$(terraform -chdir="$ROOT/infra/terraform" output -raw instance_id)"
fi

cors="${CAMPUS_CORS_ORIGINS:-}"
umask 077
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
cat >"$tmp" <<EOF
POSTGRES_PASSWORD=${TF_VAR_db_password}
DATABASE_URL=postgres://campus:${TF_VAR_db_password}@db:5432/campus_verde?sslmode=disable
CAMPUS_DEV_PASSWORD=${TF_VAR_dev_password}
CAMPUS_ENV=production
CAMPUS_COOKIE_SECURE=true
CAMPUS_CORS_ORIGINS=${cors}
EOF
b64="$(base64 -w0 "$tmp" 2>/dev/null || base64 <"$tmp" | tr -d '\n')"

aws ssm send-command \
  --region "$region" \
  --instance-ids "$instance" \
  --document-name AWS-RunShellScript \
  --comment "campus secrets.env" \
  --parameters "commands=[\"umask 077; echo $b64 | base64 -d > /opt/campus/secrets.env; chmod 600 /opt/campus/secrets.env\"]" \
  >/dev/null
echo "secrets.env escrito en $instance (el contenido no se imprime)"
