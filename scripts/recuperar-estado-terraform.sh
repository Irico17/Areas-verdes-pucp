#!/usr/bin/env bash
# Imprime comandos `terraform import` para una cuenta cuyo estado se perdió
# pero cuyos recursos siguen vivos. No ejecuta import ni apply.
#
#   recuperar-estado-terraform.sh --stack nonprod \
#     --instancias instances.json --volumenes volumes.json \
#     --direcciones addresses.json --grupos groups.json
#
# Sin esos archivos llama a AWS (hace falta una sesión viva de la MISMA cuenta).
set -euo pipefail
set +x

STACK=""
INSTANCIAS=""
VOLUMENES=""
DIRECCIONES=""
GRUPOS=""
AWS_REGION="${AWS_REGION:-${AWS_DEFAULT_REGION:-us-east-1}}"

while [ $# -gt 0 ]; do
  case "$1" in
    --stack) STACK="$2"; shift 2 ;;
    --instancias) INSTANCIAS="$2"; shift 2 ;;
    --volumenes) VOLUMENES="$2"; shift 2 ;;
    --direcciones) DIRECCIONES="$2"; shift 2 ;;
    --grupos) GRUPOS="$2"; shift 2 ;;
    *)
      echo "Argumento no reconocido: $1" >&2
      exit 2
      ;;
  esac
done

case "$STACK" in
  nonprod|prod) ;;
  *)
    echo "Use --stack nonprod o --stack prod." >&2
    exit 2
    ;;
esac

if [ -z "$INSTANCIAS" ]; then
  tag_stack="nonprod"
  if [ "$STACK" = "prod" ]; then
    tag_stack="learner-lab"
  fi
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT
  aws ec2 describe-instances --region "$AWS_REGION" \
    --filters "Name=tag:Project,Values=campus-verde" "Name=tag:Stack,Values=${tag_stack}" \
    --output json >"$tmp/instances.json"
  aws ec2 describe-volumes --region "$AWS_REGION" \
    --filters "Name=tag:Project,Values=campus-verde" "Name=tag:Stack,Values=${tag_stack}" \
    --output json >"$tmp/volumes.json"
  aws ec2 describe-addresses --region "$AWS_REGION" \
    --filters "Name=tag:Project,Values=campus-verde" "Name=tag:Stack,Values=${tag_stack}" \
    --output json >"$tmp/addresses.json"
  aws ec2 describe-security-groups --region "$AWS_REGION" \
    --filters "Name=tag:Project,Values=campus-verde" "Name=tag:Stack,Values=${tag_stack}" \
    --output json >"$tmp/groups.json"
  INSTANCIAS="$tmp/instances.json"
  VOLUMENES="$tmp/volumes.json"
  DIRECCIONES="$tmp/addresses.json"
  GRUPOS="$tmp/groups.json"
fi

python3 - "$STACK" "$INSTANCIAS" "${VOLUMENES:-}" "${DIRECCIONES:-}" "${GRUPOS:-}" <<'PY'
import json, sys

stack, inst_path, vol_path, addr_path, sg_path = sys.argv[1:]

def load(path):
    if not path:
        return {}
    with open(path, encoding="utf-8") as fh:
        return json.load(fh)

if stack == "nonprod":
    sg_addr = "aws_security_group.nonprod"
    inst_addr = "aws_instance.nonprod"
    eip_addr = "aws_eip.nonprod"
    assoc_addr = "aws_eip_association.nonprod"
    chdir = "infra/terraform/nonprod"
else:
    sg_addr = "aws_security_group.app"
    inst_addr = "aws_instance.app"
    eip_addr = "aws_eip.app"
    assoc_addr = "aws_eip_association.app"
    chdir = "infra/terraform"

print("# Revise cada línea antes de ejecutarla. Esto no importa ni aplica nada.")
print(f"# Directorio: {chdir}")
print("# No copie el estado de otra cuenta: esos id no existen aquí.")

for res in load(sg_path).get("SecurityGroups", []):
    gid = res.get("GroupId", "")
    if gid.startswith("sg-"):
        print(f"terraform -chdir={chdir} import {sg_addr} {gid}")

for res in load(inst_path).get("Reservations", []):
    for inst in res.get("Instances", []):
        iid = inst.get("InstanceId", "")
        if iid.startswith("i-"):
            print(f"terraform -chdir={chdir} import {inst_addr} {iid}")

for vol in load(vol_path).get("Volumes", []):
    vid = vol.get("VolumeId", "")
    if vid.startswith("vol-"):
        print(f"terraform -chdir={chdir} import aws_ebs_volume.data {vid}")
    for att in vol.get("Attachments", []):
        dev = att.get("Device", "")
        iid = att.get("InstanceId", "")
        if vid.startswith("vol-") and iid.startswith("i-") and dev:
            print(f"terraform -chdir={chdir} import aws_volume_attachment.data {dev}:{vid}:{iid}")

for addr in load(addr_path).get("Addresses", []):
    alloc = addr.get("AllocationId", "")
    assoc = addr.get("AssociationId", "")
    if alloc.startswith("eipalloc-"):
        print(f"terraform -chdir={chdir} import {eip_addr} {alloc}")
    if assoc.startswith("eipassoc-"):
        print(f"terraform -chdir={chdir} import {assoc_addr} {assoc}")

if stack == "prod":
    print("# Si create_ecr sigue en true, importe también los repositorios por nombre:")
    print("# terraform -chdir=infra/terraform import 'aws_ecr_repository.api[0]' campus-verde-api")
    print("# terraform -chdir=infra/terraform import 'aws_ecr_repository.web[0]' campus-verde-web")
PY
