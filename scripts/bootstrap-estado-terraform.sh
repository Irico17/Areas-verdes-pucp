#!/usr/bin/env bash
# Crea, de forma idempotente, el cubo de estado y el cubo de transferencia SSM
# en la cuenta cuyas credenciales están en el entorno. No crea IAM.
# Escribe backend.hcl si se pasa la ruta como primer argumento.
# Por stdout solo salen asignaciones KEY=valor, sin secretos.
set -euo pipefail
set +x

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
REGION="${AWS_REGION:-${AWS_DEFAULT_REGION:-us-east-1}}"
BACKEND_HCL="${1:-}"

case "$REGION" in
  us-east-1|us-west-2) ;;
  *)
    echo "La región debe ser us-east-1 o us-west-2." >&2
    exit 2
    ;;
esac

: "${AWS_ACCESS_KEY_ID:?Faltan credenciales. Cárguelas en el environment aws-lab.}"
: "${AWS_SECRET_ACCESS_KEY:?Falta AWS_SECRET_ACCESS_KEY.}"
: "${AWS_SESSION_TOKEN:?Falta AWS_SESSION_TOKEN. El Learner Lab lo exige; si caducó, vuelva a Start Lab.}"

export AWS_DEFAULT_REGION="$REGION"
export AWS_REGION="$REGION"

tmpdir="$(mktemp -d)"
cleanup() { rm -rf "$tmpdir"; }
trap cleanup EXIT

err="$tmpdir/err"
ident="$tmpdir/ident.json"

if ! aws sts get-caller-identity --output json >"$ident" 2>"$err"; then
  echo "Las credenciales del Learner Lab no sirven: caducaron o son inválidas." >&2
  echo "Start Lab, copie el bloque AWS CLI y ejecute scripts/actualizar-credenciales-lab.sh." >&2
  if ! grep -q "$AWS_ACCESS_KEY_ID" "$err"; then
    cat "$err" >&2 || true
  fi
  exit 1
fi

ACCOUNT_ID="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["Account"])' "$ident")"
ARN="$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["Arn"])' "$ident")"
if ! [[ "$ACCOUNT_ID" =~ ^[0-9]{12}$ ]]; then
  echo "La identidad de AWS no devolvió un id de cuenta de 12 dígitos." >&2
  exit 1
fi

STATE_BUCKET="campus-verde-tfstate-${ACCOUNT_ID}"
SSM_BUCKET="campus-verde-ssm-${ACCOUNT_ID}"

crear_cubo() {
  local bucket="$1"
  if aws s3api head-bucket --bucket "$bucket" --region "$REGION" >/dev/null 2>"$err"; then
    echo "El cubo ${bucket} ya existe." >&2
    return 0
  fi
  if grep -Eq '404|Not Found|NoSuchBucket' "$err"; then
    echo "Creando el cubo ${bucket}." >&2
    if [ "$REGION" = "us-east-1" ]; then
      aws s3api create-bucket --bucket "$bucket" --region "$REGION" >/dev/null
    else
      aws s3api create-bucket --bucket "$bucket" --region "$REGION" \
        --create-bucket-configuration "LocationConstraint=${REGION}" >/dev/null
    fi
    return 0
  fi
  echo "No se pudo comprobar el cubo ${bucket}." >&2
  if ! grep -q "$AWS_ACCESS_KEY_ID" "$err"; then
    cat "$err" >&2 || true
  fi
  exit 1
}

cifrar_y_bloquear() {
  local bucket="$1"
  aws s3api put-public-access-block --bucket "$bucket" --region "$REGION" \
    --public-access-block-configuration BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true \
    >/dev/null
  aws s3api put-bucket-encryption --bucket "$bucket" --region "$REGION" \
    --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"},"BucketKeyEnabled":true}]}' \
    >/dev/null
}

crear_cubo "$STATE_BUCKET"
cifrar_y_bloquear "$STATE_BUCKET"
aws s3api put-bucket-versioning --bucket "$STATE_BUCKET" --region "$REGION" \
  --versioning-configuration Status=Enabled >/dev/null
cat >"$tmpdir/life-state.json" <<'JSON'
{
  "Rules": [
    {
      "ID": "abortar-multipart",
      "Status": "Enabled",
      "Filter": {"Prefix": ""},
      "AbortIncompleteMultipartUpload": {"DaysAfterInitiation": 7}
    }
  ]
}
JSON
aws s3api put-bucket-lifecycle-configuration --bucket "$STATE_BUCKET" --region "$REGION" \
  --lifecycle-configuration "file://${tmpdir}/life-state.json" >/dev/null

crear_cubo "$SSM_BUCKET"
cifrar_y_bloquear "$SSM_BUCKET"
# Sin versionado: el plugin SSM borra los objetos al terminar, y un versionado
# conservaría para siempre cualquier archivo que Ansible hubiera subido.
aws s3api put-bucket-versioning --bucket "$SSM_BUCKET" --region "$REGION" \
  --versioning-configuration Status=Suspended >/dev/null
cat >"$tmpdir/life-ssm.json" <<'JSON'
{
  "Rules": [
    {
      "ID": "expirar-transferencia",
      "Status": "Enabled",
      "Filter": {"Prefix": ""},
      "Expiration": {"Days": 1},
      "AbortIncompleteMultipartUpload": {"DaysAfterInitiation": 1}
    }
  ]
}
JSON
aws s3api put-bucket-lifecycle-configuration --bucket "$SSM_BUCKET" --region "$REGION" \
  --lifecycle-configuration "file://${tmpdir}/life-ssm.json" >/dev/null

if [ -n "$BACKEND_HCL" ]; then
  mkdir -p "$(dirname "$BACKEND_HCL")"
  cat >"$BACKEND_HCL" <<EOF
bucket = "${STATE_BUCKET}"
region = "${REGION}"
EOF
  chmod 600 "$BACKEND_HCL"
fi

echo "Cuenta ${ACCOUNT_ID} (${ARN})." >&2
echo "Estado: s3://${STATE_BUCKET}/learner-lab/terraform.tfstate y learner-lab-nonprod/terraform.tfstate." >&2
echo "Transferencia Ansible/SSM: s3://${SSM_BUCKET}/ (expira al día, sin versionado)." >&2

printf 'ACCOUNT_ID=%s\n' "$ACCOUNT_ID"
printf 'STATE_BUCKET=%s\n' "$STATE_BUCKET"
printf 'SSM_BUCKET=%s\n' "$SSM_BUCKET"
printf 'REGION=%s\n' "$REGION"
