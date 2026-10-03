#!/usr/bin/env bash
# Sube el bloque «AWS CLI» del Learner Lab al environment de GitHub aws-lab.
# Lee stdin (si no es una terminal) o el portapapeles. No imprime los valores
# y no los escribe en el repositorio.
#
#   pbpaste | bash scripts/actualizar-credenciales-lab.sh
#   bash scripts/actualizar-credenciales-lab.sh < bloque.txt
set -euo pipefail
set +x

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=scripts/lib/credenciales-lab.sh
source "$ROOT/scripts/lib/credenciales-lab.sh"

ENV_NAME="${ENV_NAME:-aws-lab}"
cd "$ROOT"

if ! command -v gh >/dev/null 2>&1; then
  echo "Falta el comando gh. Instálelo y entre con: gh auth login" >&2
  exit 1
fi

portapapeles() {
  if command -v wl-paste >/dev/null 2>&1; then
    wl-paste -n 2>/dev/null || true
  elif command -v xclip >/dev/null 2>&1; then
    xclip -selection clipboard -o 2>/dev/null || true
  elif command -v xsel >/dev/null 2>&1; then
    xsel --clipboard --output 2>/dev/null || true
  elif command -v pbpaste >/dev/null 2>&1; then
    pbpaste 2>/dev/null || true
  elif command -v powershell.exe >/dev/null 2>&1; then
    powershell.exe -NoProfile -Command "Get-Clipboard -Raw" 2>/dev/null || true
  fi
}

leer_bloque() {
  if [ ! -t 0 ]; then
    cat
    return
  fi
  local clip
  clip="$(portapapeles || true)"
  if [ -n "$clip" ]; then
    printf '%s\n' "$clip"
    return
  fi
  echo "Pegue el bloque de AWS Details → AWS CLI → Show y termine con una línea vacía." >&2
  stty -echo || true
  trap 'stty echo 2>/dev/null || true' EXIT
  local linea pegado=""
  while IFS= read -r linea; do
    linea="${linea%$'\r'}"
    if [ -z "$linea" ] && [ -n "$pegado" ]; then
      break
    fi
    pegado+="$linea"$'\n'
  done
  stty echo 2>/dev/null || true
  echo >&2
  printf '%s' "$pegado"
}

bloque="$(leer_bloque)"
parsear_bloque_aws <<<"$bloque"
unset bloque
validar_forma_credenciales

subir() {
  local nombre="$1"
  local valor="$2"
  printf '%s' "$valor" | gh secret set "$nombre" --env "$ENV_NAME" >/dev/null
}

subir AWS_ACCESS_KEY_ID "$PARSE_KEY"
subir AWS_SECRET_ACCESS_KEY "$PARSE_SECRET"
subir AWS_SESSION_TOKEN "$PARSE_TOKEN"
gh variable set AWS_REGION --env "$ENV_NAME" --body "$PARSE_REGION" >/dev/null
unset PARSE_KEY PARSE_SECRET PARSE_TOKEN

echo "Credenciales cargadas en el environment ${ENV_NAME} (región ${PARSE_REGION}). Los valores no se muestran."
echo "Caducan con la sesión del Learner Lab. Cuando Start Lab entregue otras, vuelva a ejecutar este script."
