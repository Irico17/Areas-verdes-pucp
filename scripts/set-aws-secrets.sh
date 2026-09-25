#!/usr/bin/env bash
# Sube las credenciales temporales del Learner Lab a los secrets del repo.
# Lee el bloque del portapapeles (AWS Details > AWS CLI > Show) o, si no trae
# las tres claves, el perfil [default] de ~/.aws/credentials.
# No imprime los valores ni los escribe en el repositorio.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! command -v gh >/dev/null 2>&1; then
  echo "Falta el comando gh. Instálelo y entre con: gh auth login" >&2
  exit 1
fi

extraer() {
  KEY_ID=""
  SECRET=""
  TOKEN=""
  en_default=0
  vio_seccion=0
  while IFS= read -r linea || [ -n "$linea" ]; do
    linea="${linea%$'\r'}"
    case "$linea" in
      \[*\])
        vio_seccion=1
        if [ "$linea" = "[default]" ]; then
          en_default=1
        else
          en_default=0
        fi
        continue
        ;;
    esac
    if [ "$vio_seccion" -eq 1 ] && [ "$en_default" -eq 0 ]; then
      continue
    fi
    limpia="${linea#export }"
    case "$limpia" in
      aws_access_key_id=*|AWS_ACCESS_KEY_ID=*)
        KEY_ID="${limpia#*=}"
        ;;
      aws_secret_access_key=*|AWS_SECRET_ACCESS_KEY=*)
        SECRET="${limpia#*=}"
        ;;
      aws_session_token=*|AWS_SESSION_TOKEN=*)
        TOKEN="${limpia#*=}"
        ;;
      *)
        continue
        ;;
    esac
  done
  KEY_ID="$(printf '%s' "$KEY_ID" | tr -d '"' | tr -d "'")"
  SECRET="$(printf '%s' "$SECRET" | tr -d '"' | tr -d "'")"
  TOKEN="$(printf '%s' "$TOKEN" | tr -d '"' | tr -d "'")"
  KEY_ID="${KEY_ID#"${KEY_ID%%[![:space:]]*}"}"
  KEY_ID="${KEY_ID%"${KEY_ID##*[![:space:]]}"}"
  SECRET="${SECRET#"${SECRET%%[![:space:]]*}"}"
  SECRET="${SECRET%"${SECRET##*[![:space:]]}"}"
  TOKEN="${TOKEN#"${TOKEN%%[![:space:]]*}"}"
  TOKEN="${TOKEN%"${TOKEN##*[![:space:]]}"}"
}

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

completas() {
  [ -n "${KEY_ID:-}" ] && [ -n "${SECRET:-}" ] && [ -n "${TOKEN:-}" ]
}

fuente=""
KEY_ID=""
SECRET=""
TOKEN=""

clip="$(portapapeles || true)"
if [ -n "$clip" ]; then
  extraer <<<"$clip"
  if completas; then
    fuente="portapapeles"
  fi
fi

if ! completas; then
  cred="${AWS_SHARED_CREDENTIALS_FILE:-$HOME/.aws/credentials}"
  if [ -f "$cred" ]; then
    extraer <"$cred"
    if completas; then
      fuente="archivo"
    fi
  fi
fi

if ! completas; then
  echo "No hay tres claves en el portapapeles ni en ~/.aws/credentials." >&2
  echo "Pegue el bloque de AWS CLI > Show y termine con una línea vacía." >&2
  if [ -t 0 ]; then
    stty -echo || true
    trap 'stty echo 2>/dev/null || true' EXIT
  fi
  pegado=""
  while IFS= read -r linea; do
    linea="${linea%$'\r'}"
    if [ -z "$linea" ] && [ -n "$pegado" ]; then
      break
    fi
    pegado+="$linea"$'\n'
  done
  if [ -t 0 ]; then
    stty echo || true
    echo >&2
  fi
  extraer <<<"$pegado"
  fuente="pegado"
fi

if ! completas; then
  echo "El bloque no trae aws_access_key_id, aws_secret_access_key y aws_session_token." >&2
  exit 1
fi

subir() {
  local nombre="$1"
  local valor="$2"
  printf '%s' "$valor" | gh secret set "$nombre"
}

subir AWS_ACCESS_KEY_ID "$KEY_ID"
subir AWS_SECRET_ACCESS_KEY "$SECRET"
subir AWS_SESSION_TOKEN "$TOKEN"
unset KEY_ID SECRET TOKEN clip pegado

echo "Secrets actualizados desde ${fuente} (los valores no se muestran)."

if [ -t 0 ]; then
  printf '¿Lanzar el despliegue (gh workflow run ci --ref main)? [s/N] '
  read -r respuesta || respuesta=""
  case "$respuesta" in
    s|S|y|Y)
      gh workflow run ci --ref main
      echo "Despliegue pedido en main. El job imprime la URL al terminar."
      ;;
    *)
      echo "Para lanzarlo después: gh workflow run ci --ref main"
      ;;
  esac
else
  echo "Para lanzarlo: gh workflow run ci --ref main"
fi
