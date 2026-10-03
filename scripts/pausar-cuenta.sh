#!/usr/bin/env bash
# Enciende o detiene las instancias etiquetadas por este Terraform.
# Detener no borra discos ni la IP elástica: esos siguen consumiendo crédito.
set -euo pipefail
set +x

ACCION="${ACCION:-stop}"
AMBIENTES="${AMBIENTES:-todos}"
CONFIRMAR="${CONFIRMAR:-}"
AWS_REGION="${AWS_REGION:-${AWS_DEFAULT_REGION:-us-east-1}}"

case "$ACCION" in
  start|stop) ;;
  *)
    echo "accion debe ser start o stop." >&2
    exit 2
    ;;
esac
case "$AMBIENTES" in
  nonprod|prod|todos) ;;
  *)
    echo "ambientes debe ser nonprod, prod o todos." >&2
    exit 2
    ;;
esac
case "$AWS_REGION" in
  us-east-1|us-west-2) ;;
  *)
    echo "AWS_REGION debe ser us-east-1 o us-west-2." >&2
    exit 2
    ;;
esac
case "$ACCION" in
  stop)
    if [ "$CONFIRMAR" != "DETENER" ]; then
      echo "Para detener hay que escribir confirmar=DETENER. No se detuvo nada." >&2
      exit 2
    fi
    ;;
  start)
    if [ "$CONFIRMAR" != "ENCENDER" ]; then
      echo "Para encender hay que escribir confirmar=ENCENDER. No se encendió nada." >&2
      exit 2
    fi
    ;;
esac

if [ "${PAUSAR_SOLO_VALIDAR:-}" = "1" ]; then
  echo "Entradas válidas: accion=${ACCION} ambientes=${AMBIENTES}."
  exit 0
fi

: "${AWS_ACCESS_KEY_ID:?Faltan credenciales del environment aws-lab.}"
: "${AWS_SECRET_ACCESS_KEY:?Falta AWS_SECRET_ACCESS_KEY.}"
: "${AWS_SESSION_TOKEN:?Falta AWS_SESSION_TOKEN. Si caducó la sesión, actualice los secretos.}"
export AWS_REGION AWS_DEFAULT_REGION="$AWS_REGION"

if ! aws sts get-caller-identity --output text >/dev/null; then
  echo "Las credenciales del Learner Lab no sirven: caducaron o son inválidas." >&2
  echo "Ejecute scripts/actualizar-credenciales-lab.sh con un bloque nuevo." >&2
  exit 1
fi

stacks=()
case "$AMBIENTES" in
  nonprod) stacks=(nonprod) ;;
  prod) stacks=(learner-lab) ;;
  todos) stacks=(nonprod learner-lab) ;;
esac

# AWS trata varios valores del mismo filtro como OR.
stack_values="$(IFS=,; echo "${stacks[*]}")"
estado_buscado="running"
if [ "$ACCION" = "start" ]; then
  estado_buscado="stopped"
fi

json="$(aws ec2 describe-instances --region "$AWS_REGION" \
  --filters \
    "Name=tag:Project,Values=campus-verde" \
    "Name=tag:CampusGestion,Values=cuenta-lab" \
    "Name=tag:Stack,Values=${stack_values}" \
    "Name=instance-state-name,Values=${estado_buscado}" \
  --output json)"

mapfile -t ids < <(python3 -c 'import json,sys; d=json.loads(sys.argv[1]);
ids=[]
for r in d.get("Reservations", []):
  for i in r.get("Instances", []):
    ids.append(i["InstanceId"])
print("\n".join(ids))' "$json")

if [ "${#ids[@]}" -eq 0 ]; then
  echo "No hay instancias ${estado_buscado} con Project=campus-verde, CampusGestion=cuenta-lab y Stack=${stack_values}."
  echo "Si la máquina es anterior a esta etiqueta, un apply de aprovisionar-cuenta la añade sin reemplazarla."
  exit 0
fi

echo "Instancias: ${ids[*]}"
if [ "$ACCION" = "stop" ]; then
  aws ec2 stop-instances --region "$AWS_REGION" --instance-ids "${ids[@]}" --output text >/dev/null
  echo "Detenidas. El disco y la IP elástica siguen en la factura. Para soltarlos, use destroy."
else
  aws ec2 start-instances --region "$AWS_REGION" --instance-ids "${ids[@]}" --output text >/dev/null
  echo "Encendidas. El runner systemd debería volver solo."
fi
