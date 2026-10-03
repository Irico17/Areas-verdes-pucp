#!/usr/bin/env bash
# Arma infra/ansible/inventory/hosts.yml desde las salidas JSON de Terraform.
# Uso:
#   generar-inventario-ansible.sh --region us-east-1 --ssm-bucket NOMBRE \
#     --nonprod salida-nonprod.json --prod salida-prod.json --out hosts.yml
# Cada archivo JSON es el de `terraform output -json`. Puede omitirse un stack.
set -euo pipefail

REGION=""
SSM_BUCKET=""
NONPROD=""
PROD=""
OUT=""

while [ $# -gt 0 ]; do
  case "$1" in
    --region) REGION="$2"; shift 2 ;;
    --ssm-bucket) SSM_BUCKET="$2"; shift 2 ;;
    --nonprod) NONPROD="$2"; shift 2 ;;
    --prod) PROD="$2"; shift 2 ;;
    --out) OUT="$2"; shift 2 ;;
    *)
      echo "Argumento no reconocido: $1" >&2
      exit 2
      ;;
  esac
done

if [ -z "$REGION" ] || [ -z "$SSM_BUCKET" ] || [ -z "$OUT" ]; then
  echo "Hacen falta --region, --ssm-bucket y --out." >&2
  exit 2
fi
case "$REGION" in
  us-east-1|us-west-2) ;;
  *)
    echo "Región no permitida." >&2
    exit 2
    ;;
esac
if ! [[ "$SSM_BUCKET" =~ ^campus-verde-ssm-[0-9]{12}$ ]]; then
  echo "El cubo SSM no tiene la forma campus-verde-ssm-<cuenta>." >&2
  exit 2
fi

python3 - "$REGION" "$SSM_BUCKET" "$NONPROD" "$PROD" "$OUT" <<'PY'
import json
import re
import sys

region, bucket, nonprod_path, prod_path, out_path = sys.argv[1:]
instancia = re.compile(r"^i-[0-9a-f]{8,17}$")
ipv4 = re.compile(r"^(?:(?:25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d\d?)$")

def cargar(path):
    if not path:
        return None
    with open(path, encoding="utf-8") as fh:
        data = json.load(fh)
    def val(key):
        item = data[key]
        return item["value"] if isinstance(item, dict) and "value" in item else item
    return val

hosts = []
if nonprod_path:
    datos = cargar(nonprod_path)
    iid = datos("instance_id")
    ip = datos("public_ip")
    if not instancia.match(str(iid)) or not ipv4.match(str(ip)):
        sys.exit("Salida nonprod inválida")
    hosts.append(("nonprod", iid, ip, "nonprod"))
if prod_path:
    datos = cargar(prod_path)
    iid = datos("instance_id")
    ip = datos("public_ip")
    if not instancia.match(str(iid)) or not ipv4.match(str(ip)):
        sys.exit("Salida prod inválida")
    hosts.append(("prod", iid, ip, "produccion"))
if not hosts:
    sys.exit("No hay salidas de Terraform para el inventario")

lineas = [
    "---",
    "all:",
    "  vars:",
    "    ansible_connection: amazon.aws.aws_ssm",
    f"    ansible_aws_ssm_region: {region}",
    f"    ansible_aws_ssm_bucket_name: {bucket}",
    "    ansible_aws_ssm_bucket_sse_mode: AES256",
    "    ansible_aws_ssm_timeout: 180",
    "    ansible_python_interpreter: /usr/bin/python3",
    "  children:",
    "    campus:",
    "      hosts:",
]
for nombre, iid, ip, perfil in hosts:
    lineas.extend([
        f"        {nombre}:",
        f"          ansible_host: {iid}",
        f"          ansible_aws_ssm_instance_id: {iid}",
        f"          campus_perfil: {perfil}",
        f"          public_ip: {ip}",
    ])
with open(out_path, "w", encoding="utf-8") as fh:
    fh.write("\n".join(lineas) + "\n")
PY

echo "Inventario escrito en ${OUT}."
