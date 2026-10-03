#!/usr/bin/env bash
# Plan, apply o destroy de la cuenta de Learner Lab que esté en el entorno.
# No imprime secretos. El plan binario no se sube; solo el texto ya revisado.
set -euo pipefail
set +x
umask 077

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ACCION="${ACCION:-plan}"
AMBIENTES="${AMBIENTES:-nonprod}"
CONFIRMAR="${CONFIRMAR:-}"
ACEPTO_REEMPLAZO="${ACEPTO_REEMPLAZO:-}"
DISCO_RAIZ_GB="${DISCO_RAIZ_GB:-20}"
DISCO_DATOS_GB="${DISCO_DATOS_GB:-20}"
TIPO_NONPROD="${TIPO_NONPROD:-t3.small}"
TIPO_PROD="${TIPO_PROD:-t3.micro}"
CONFIGURAR_HOSTS="${CONFIGURAR_HOSTS:-true}"
ANSIBLE_CHECK="${ANSIBLE_CHECK:-false}"
REGISTRAR_RUNNERS="${REGISTRAR_RUNNERS:-true}"
AWS_REGION="${AWS_REGION:-${AWS_DEFAULT_REGION:-us-east-1}}"
PLAN_DIR="${PLAN_DIR:-$ROOT/tmp/planes}"
RESUMEN="${RESUMEN:-$ROOT/tmp/resumen-aprovision.md}"

validar_entradas() {
  case "$ACCION" in
    plan|apply|destroy) ;;
    *)
      echo "accion debe ser plan, apply o destroy." >&2
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
  case "$TIPO_NONPROD" in
    t3.micro|t3.small|t3.medium) ;;
    *)
      echo "tipo nonprod debe ser t3.micro, t3.small o t3.medium." >&2
      exit 2
      ;;
  esac
  case "$TIPO_PROD" in
    t3.micro|t3.small) ;;
    *)
      echo "tipo prod debe ser t3.micro o t3.small." >&2
      exit 2
      ;;
  esac
  if ! [[ "$DISCO_RAIZ_GB" =~ ^[0-9]+$ ]] || [ "$DISCO_RAIZ_GB" -lt 20 ] || [ "$DISCO_RAIZ_GB" -gt 40 ]; then
    echo "disco raíz debe ser un entero entre 20 y 40." >&2
    exit 2
  fi
  if ! [[ "$DISCO_DATOS_GB" =~ ^[0-9]+$ ]] || [ "$DISCO_DATOS_GB" -lt 20 ] || [ "$DISCO_DATOS_GB" -gt 40 ]; then
    echo "disco de datos debe ser un entero entre 20 y 40." >&2
    exit 2
  fi
  case "$ACCION" in
    apply)
      if [ "$CONFIRMAR" != "APLICAR" ]; then
        echo "Para apply hay que escribir confirmar=APLICAR. No se cambió la cuenta." >&2
        exit 2
      fi
      ;;
    destroy)
      if [ "$CONFIRMAR" != "DESTRUIR" ]; then
        echo "Para destroy hay que escribir confirmar=DESTRUIR. No se destruyó nada." >&2
        exit 2
      fi
      ;;
  esac
}

validar_entradas
if [ "${APROVISIONAR_SOLO_VALIDAR:-}" = "1" ]; then
  echo "Entradas válidas: accion=${ACCION} ambientes=${AMBIENTES} región=${AWS_REGION}."
  exit 0
fi

export AWS_REGION AWS_DEFAULT_REGION="$AWS_REGION"
unset TF_LOG TF_LOG_PATH || true

if ! command -v terraform >/dev/null 2>&1; then
  echo "No está terraform en PATH. El workflow lo instala con scripts/instalar-terraform.sh." >&2
  exit 1
fi
if ! command -v aws >/dev/null 2>&1; then
  echo "No está aws en PATH." >&2
  exit 1
fi

: "${AWS_ACCESS_KEY_ID:?Faltan credenciales del environment aws-lab.}"
: "${AWS_SECRET_ACCESS_KEY:?Falta AWS_SECRET_ACCESS_KEY.}"
: "${AWS_SESSION_TOKEN:?Falta AWS_SESSION_TOKEN. Si la sesión del lab caducó, actualice los secretos.}"

echo "Comprobando terraform fmt."
terraform fmt -check -recursive "$ROOT/infra/terraform"

mkdir -p "$PLAN_DIR" "$(dirname "$RESUMEN")"
BACKEND_HCL="$ROOT/tmp/backend.hcl"
info="$(bash "$ROOT/scripts/bootstrap-estado-terraform.sh" "$BACKEND_HCL")"
ACCOUNT_ID="$(printf '%s\n' "$info" | sed -n 's/^ACCOUNT_ID=//p')"
STATE_BUCKET="$(printf '%s\n' "$info" | sed -n 's/^STATE_BUCKET=//p')"
SSM_BUCKET="$(printf '%s\n' "$info" | sed -n 's/^SSM_BUCKET=//p')"
if ! [[ "$ACCOUNT_ID" =~ ^[0-9]{12}$ ]]; then
  echo "Bootstrap no devolvió un id de cuenta válido." >&2
  exit 1
fi
if ! [[ "$STATE_BUCKET" =~ ^campus-verde-tfstate-[0-9]{12}$ ]]; then
  echo "Bootstrap no devolvió un cubo de estado válido." >&2
  exit 1
fi
if ! [[ "$SSM_BUCKET" =~ ^campus-verde-ssm-[0-9]{12}$ ]]; then
  echo "Bootstrap no devolvió un cubo SSM válido." >&2
  exit 1
fi

token_file=""
limpiar_secretos_remotos() {
  if [ -n "${token_file:-}" ]; then
    rm -f "$token_file"
  fi
  if [ "${TOKENS_SSM:-}" = "1" ]; then
    local amb
    for amb in develop qa produccion; do
      aws ssm delete-parameter --region "$AWS_REGION" --name "/campus/runner-token/${amb}" >/dev/null 2>&1 || true
    done
  fi
  rm -f "$PLAN_DIR"/*.tfplan
}
trap limpiar_secretos_remotos EXIT

quiere() {
  case "$AMBIENTES" in
    todos) return 0 ;;
    "$1") return 0 ;;
    *) return 1 ;;
  esac
}

ejecutar_stack() {
  local nombre="$1" dir="$2" tipo="$3"
  local planbin="$PLAN_DIR/${nombre}.tfplan"
  local plantxt="$PLAN_DIR/${nombre}.txt"
  local -a vars
  vars=(
    -var "aws_region=${AWS_REGION}"
    -var "instance_type=${tipo}"
    -var "root_volume_gb=${DISCO_RAIZ_GB}"
    -var "data_volume_gb=${DISCO_DATOS_GB}"
  )
  if [ "$nombre" = "prod" ]; then
    vars+=(-var "ambiente=produccion")
    if [ -z "${TF_VAR_db_password:-}" ] || [ -z "${TF_VAR_dev_password:-}" ]; then
      echo "Faltan TF_VAR_db_password y TF_VAR_dev_password en el environment aws-lab." >&2
      echo "Terraform de producción los exige y no los escribe en la instancia. No se imprimen." >&2
      exit 1
    fi
  fi
  local -a destroy_flag=()
  if [ "$ACCION" = "destroy" ]; then
    destroy_flag=(-destroy)
  fi

  echo "Terraform init de ${nombre}."
  terraform -chdir="$dir" init -input=false -reconfigure -backend-config="$BACKEND_HCL"
  terraform -chdir="$dir" validate
  echo "Terraform plan de ${nombre}."
  terraform -chdir="$dir" plan -input=false -lock-timeout=5m -no-color \
    "${vars[@]}" "${destroy_flag[@]}" -out="$planbin"
  terraform -chdir="$dir" show -no-color "$planbin" >"$plantxt"
  if ! bash "$ROOT/scripts/assert-plan-sin-secretos.sh" "$plantxt"; then
    rm -f "$plantxt" "$planbin"
    exit 1
  fi
  chmod 644 "$plantxt"

  if [ "$ACCION" = "plan" ]; then
    rm -f "$planbin"
    return 0
  fi
  if [ "$ACCION" = "apply" ] && grep -q 'forces replacement' "$plantxt"; then
    if [ "$ACEPTO_REEMPLAZO" != "SI" ]; then
      echo "El plan de ${nombre} reemplaza recursos ya creados. No se aplicó." >&2
      echo "Si esa máquina tiene datos, cancele y repita con el tamaño actual." >&2
      echo "Si acepta perderlos, vuelva a lanzar con acepto_reemplazo=SI." >&2
      exit 2
    fi
    echo "AVISO: ${nombre} reemplaza recursos y acepto_reemplazo=SI." >&2
  fi
  echo "Terraform apply del plan exacto de ${nombre}."
  terraform -chdir="$dir" apply -input=false -lock-timeout=5m -no-color "$planbin"
  rm -f "$planbin"
}

if quiere nonprod; then
  ejecutar_stack nonprod "$ROOT/infra/terraform/nonprod" "$TIPO_NONPROD"
fi
if quiere prod; then
  ejecutar_stack prod "$ROOT/infra/terraform" "$TIPO_PROD"
fi

salida_json() {
  local dir="$1" dest="$2"
  if ! terraform -chdir="$dir" output -json >"$dest" 2>/dev/null; then
    rm -f "$dest"
    return 1
  fi
  return 0
}

NONPROD_JSON=""
PROD_JSON=""
if quiere nonprod && salida_json "$ROOT/infra/terraform/nonprod" "$PLAN_DIR/nonprod-output.json"; then
  NONPROD_JSON="$PLAN_DIR/nonprod-output.json"
fi
if quiere prod && salida_json "$ROOT/infra/terraform" "$PLAN_DIR/prod-output.json"; then
  PROD_JSON="$PLAN_DIR/prod-output.json"
fi

configurar() {
  local check="$1"
  if [ -z "$NONPROD_JSON" ] && [ -z "$PROD_JSON" ]; then
    echo "No hay salidas de Terraform: no se configura el host." >&2
    return 0
  fi
  if ! command -v ansible-playbook >/dev/null 2>&1; then
    echo "Falta ansible-playbook." >&2
    return 1
  fi
  local inv="$ROOT/infra/ansible/inventory/hosts.yml"
  local -a args=(--region "$AWS_REGION" --ssm-bucket "$SSM_BUCKET" --out "$inv")
  if [ -n "$NONPROD_JSON" ]; then
    args+=(--nonprod "$NONPROD_JSON")
  fi
  if [ -n "$PROD_JSON" ]; then
    args+=(--prod "$PROD_JSON")
  fi
  bash "$ROOT/scripts/generar-inventario-ansible.sh" "${args[@]}"
  local -a play=(
    ansible-playbook "$ROOT/infra/ansible/site.yml"
    -i "$inv"
  )
  if [ "$REGISTRAR_RUNNERS" = "true" ] && [ "$check" != "check" ]; then
    play+=(-e "campus_registrar_runner=true")
  else
    play+=(-e "campus_registrar_runner=false")
  fi
  if [ "$check" = "check" ]; then
    play+=(--check)
  fi
  ANSIBLE_CONFIG="$ROOT/infra/ansible/ansible.cfg" "${play[@]}"
}

esperar_ssm() {
  local id="$1" i status
  echo "Esperando a que SSM vea ${id}."
  for i in $(seq 1 40); do
    status="$(aws ssm describe-instance-information --region "$AWS_REGION" \
      --filters "Key=InstanceIds,Values=${id}" \
      --query 'InstanceInformationList[0].PingStatus' --output text 2>/dev/null || true)"
    if [ "$status" = "Online" ]; then
      echo "SSM Online para ${id}."
      return 0
    fi
    sleep 15
  done
  echo "SSM no llegó a Online para ${id}. Revise LabInstanceProfile y el agente." >&2
  return 1
}

if [ "$ACCION" = "apply" ] && [ "$CONFIGURAR_HOSTS" = "true" ]; then
  if [ "$REGISTRAR_RUNNERS" = "true" ]; then
    if [ -z "${RUNNER_REG_TOKEN_PAT:-}" ]; then
      echo "Falta el secreto RUNNER_REG_TOKEN_PAT en aws-lab. La infraestructura quedó creada; los runners no se registraron." >&2
      echo "Cree un PAT de grano fino, solo este repo, permiso Administration lectura y escritura. Vea docs/CAMBIO-DE-CUENTA-LAB.md." >&2
      exit 1
    fi
    if [ -z "${GITHUB_REPOSITORY:-}" ]; then
      echo "Falta GITHUB_REPOSITORY para pedir el token de registro." >&2
      exit 1
    fi
    token_file="$(mktemp)"
    json_token="$(mktemp)"
    if ! GH_TOKEN="$RUNNER_REG_TOKEN_PAT" gh api -X POST \
      "repos/${GITHUB_REPOSITORY}/actions/runners/registration-token" >"$json_token"; then
      echo "GitHub rechazó el PAT al crear el token de registro del runner." >&2
      rm -f "$json_token"
      exit 1
    fi
    python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["token"], end="")' "$json_token" >"$token_file"
    rm -f "$json_token"
    chmod 600 "$token_file"
    if [ ! -s "$token_file" ]; then
      echo "El token de registro llegó vacío." >&2
      exit 1
    fi
    poner_token() {
      aws ssm put-parameter --region "$AWS_REGION" \
        --name "/campus/runner-token/${1}" \
        --type SecureString \
        --value "file://${token_file}" \
        --overwrite --output text >/dev/null
    }
    if [ -n "$NONPROD_JSON" ]; then
      poner_token develop
      poner_token qa
    fi
    if [ -n "$PROD_JSON" ]; then
      poner_token produccion
    fi
    TOKENS_SSM=1
  fi
  python3 - "$NONPROD_JSON" "$PROD_JSON" <<'PY' | while IFS= read -r iid; do
import json, sys
for path in sys.argv[1:]:
    if not path:
        continue
    data = json.load(open(path, encoding="utf-8"))
    item = data.get("instance_id", {})
    print(item.get("value", item))
PY
    esperar_ssm "$iid"
  done
  configurar apply
fi

if [ "$ACCION" = "plan" ] && [ "$ANSIBLE_CHECK" = "true" ]; then
  REGISTRAR_RUNNERS=false
  configurar check || echo "Ansible --check no pudo correr (¿la instancia aún no existe?)." >&2
fi

publicar_urls() {
  if [ -z "${RUNNER_REG_TOKEN_PAT:-}" ]; then
    echo "Sin RUNNER_REG_TOKEN_PAT no se actualiza PUBLIC_URL. El resumen deja las URL para gh variable set." >&2
    return 1
  fi
  GH_TOKEN="$RUNNER_REG_TOKEN_PAT" python3 - "$NONPROD_JSON" "$PROD_JSON" <<'PY'
import json, os, subprocess, sys
nonprod, prod = sys.argv[1:]

def ip_de(path):
    if not path:
        return ""
    data = json.load(open(path, encoding="utf-8"))
    item = data.get("public_ip", {})
    return item.get("value", item)

def publicar(ambiente, url):
    subprocess.run(
        ["gh", "variable", "set", "PUBLIC_URL", "--env", ambiente, "--body", url],
        check=True,
    )
    print(f"PUBLIC_URL de {ambiente} actualizada.")

ip_np = ip_de(nonprod)
ip_pr = ip_de(prod)
if ip_np:
    publicar("develop", f"http://{ip_np}:8088")
    publicar("qa", f"http://{ip_np}:8188")
if ip_pr:
    publicar("produccion", f"http://{ip_pr}")
PY
}

URLS_OK="omitido"
if [ "$ACCION" = "apply" ]; then
  if publicar_urls; then
    URLS_OK="si"
  else
    URLS_OK="no"
  fi
fi
unset RUNNER_REG_TOKEN_PAT || true

{
  echo "### Aprovisionamiento ${ACCION}"
  echo "- Cuenta: \`${ACCOUNT_ID}\`"
  echo "- Región: \`${AWS_REGION}\`"
  echo "- Estado: \`${STATE_BUCKET}\` (claves learner-lab y learner-lab-nonprod)"
  echo "- Ambientes pedidos: \`${AMBIENTES}\`"
  echo "- PUBLIC_URL actualizada: \`${URLS_OK:-omitido}\`"
  if [ -f "$PLAN_DIR/nonprod.txt" ]; then
    echo "- Plan nonprod: artefacto de texto, sin secretos."
  fi
  if [ -f "$PLAN_DIR/prod.txt" ]; then
    echo "- Plan prod: artefacto de texto, sin secretos."
  fi
  if [ -n "$NONPROD_JSON" ]; then
    echo "- Nonprod: revise las salidas de Terraform en el log (IP e id de instancia)."
  fi
} >"$RESUMEN"

echo "Listo: accion=${ACCION} cuenta=${ACCOUNT_ID}."
