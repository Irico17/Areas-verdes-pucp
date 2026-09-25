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
# Si el operador exporta CAMPUS_COOKIE_SECURE, se respeta. Si no, la instancia
# lo deduce de los PEM en /opt/campus/certs (los mismos que nginx monta en /etc/nginx/certs).
cookie="${CAMPUS_COOKIE_SECURE:-}"
umask 077
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT
cat >"$tmp" <<EOF
POSTGRES_PASSWORD=${TF_VAR_db_password}
DATABASE_URL=postgres://campus:${TF_VAR_db_password}@db:5432/campus_verde?sslmode=disable
CAMPUS_DEV_PASSWORD=${TF_VAR_dev_password}
CAMPUS_ENV=production
CAMPUS_CORS_ORIGINS=${cors}
EOF
if [ -n "$cookie" ]; then
  printf 'CAMPUS_COOKIE_SECURE=%s\n' "$cookie" >>"$tmp"
fi
b64="$(base64 -w0 "$tmp" 2>/dev/null || base64 <"$tmp" | tr -d '\n')"

# Un solo comando en la instancia. Las comillas simples del grep no hacen falta
# para el patrón, pero el JSON de --parameters las aguanta igual que unas dobles.
remote="umask 077; echo ${b64} | base64 -d > /opt/campus/secrets.env; if ! grep -q ^CAMPUS_COOKIE_SECURE= /opt/campus/secrets.env; then if [ -f /opt/campus/certs/fullchain.pem ] && [ -f /opt/campus/certs/privkey.pem ]; then echo CAMPUS_COOKIE_SECURE=true >> /opt/campus/secrets.env; else echo CAMPUS_COOKIE_SECURE=false >> /opt/campus/secrets.env; fi; fi; chmod 600 /opt/campus/secrets.env"

# El shorthand commands=["..."] se parte si el script trae comillas. JSON escapado no.
params="$(REMOTE="$remote" python3 -c 'import json, os; print(json.dumps({"commands": [os.environ["REMOTE"]]}))')"
python3 -c 'import json, sys; json.loads(sys.argv[1])' "$params" >/dev/null

aws ssm send-command \
  --region "$region" \
  --instance-ids "$instance" \
  --document-name AWS-RunShellScript \
  --comment "campus secrets.env" \
  --parameters "$params" \
  >/dev/null
echo "secrets.env escrito en $instance (el contenido no se imprime)"
