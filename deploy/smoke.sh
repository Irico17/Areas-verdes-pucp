#!/usr/bin/env bash
# Smoke post-despliegue: /health, login de una cuenta ficticia y dos lecturas.
# Uso: deploy/smoke.sh <develop|qa|produccion>
# SMOKE_BASE_URL fuerza la URL (en AWS es el Elastic IP). No imprime la clave.
set -euo pipefail

AMBIENTE="${1:?falta el ambiente: develop, qa o produccion}"
case "$AMBIENTE" in
  develop|qa|produccion) ;;
  *)
    echo "ambiente desconocido: $AMBIENTE" >&2
    exit 2
    ;;
esac

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${SMOKE_ENV_FILE:-$ROOT/deploy/env/${AMBIENTE}.env}"
if [ ! -f "$ENV_FILE" ]; then
  example="$ROOT/deploy/env/${AMBIENTE}.env.example"
  if [ ! -f "$example" ]; then
    echo "no está $ENV_FILE ni el example" >&2
    exit 1
  fi
  cp "$example" "$ENV_FILE"
  chmod 600 "$ENV_FILE"
fi

preset_password="${CAMPUS_DEV_PASSWORD:-}"
preset_base="${SMOKE_BASE_URL:-}"
preset_postgres_port="${POSTGRES_PORT:-}"
preset_api_port="${API_PORT:-}"
preset_web_port="${WEB_PORT:-}"

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a
if [ -n "$preset_password" ]; then
  CAMPUS_DEV_PASSWORD="$preset_password"
fi
if [ -n "$preset_base" ]; then
  SMOKE_BASE_URL="$preset_base"
fi
if [ -n "$preset_postgres_port" ]; then
  POSTGRES_PORT="$preset_postgres_port"
fi
if [ -n "$preset_api_port" ]; then
  API_PORT="$preset_api_port"
fi
if [ -n "$preset_web_port" ]; then
  WEB_PORT="$preset_web_port"
fi
export POSTGRES_PORT API_PORT WEB_PORT

: "${CAMPUS_DEV_PASSWORD:?falta CAMPUS_DEV_PASSWORD}"
if [ "$AMBIENTE" = "produccion" ]; then
  trimmed_dev_pass="$(printf "%s" "$CAMPUS_DEV_PASSWORD" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
  if [ "$trimmed_dev_pass" = "pando-local" ] || [ "$trimmed_dev_pass" = "campus-lab" ]; then
    echo "ERROR: en producción, CAMPUS_DEV_PASSWORD no puede ser una clave de laboratorio" >&2
    exit 1
  fi
  if [ "${#trimmed_dev_pass}" -lt 16 ]; then
    echo "ERROR: en producción, CAMPUS_DEV_PASSWORD debe tener al menos 16 caracteres (longitud actual: ${#trimmed_dev_pass})" >&2
    exit 1
  fi
fi
BASE="${SMOKE_BASE_URL:-http://127.0.0.1:${WEB_PORT:?falta WEB_PORT}}"
USER_NAME="${SMOKE_USER:-coordinacion}"
body="$(mktemp)"
jar="$(mktemp)"
hdr="$(mktemp)"
trap 'rm -f "$body" "$jar" "$hdr"' EXIT

# HTTPS: -k solo si SMOKE_INSECURE_TLS=1 (nunca por defecto).
# SMOKE_TLS_RESOLVE, si viene, fija el nombre al 127.0.0.1 del host.
curl_tls_args=()
case "$BASE" in
  https://*)
    if [ "${SMOKE_INSECURE_TLS:-0}" = "1" ]; then
      curl_tls_args+=(-k)
    fi
    if [ -n "${SMOKE_TLS_RESOLVE:-}" ]; then
      curl_tls_args+=(--resolve "$SMOKE_TLS_RESOLVE")
    fi
    if [ -n "${SMOKE_CACERT:-}" ]; then
      curl_tls_args+=(--cacert "$SMOKE_CACERT")
    fi
    ;;
esac

esperar_health() {
  local intento=0
  while [ "$intento" -lt 45 ]; do
    intento=$((intento + 1))
    code="$(curl -sS "${curl_tls_args[@]}" -o "$body" -w '%{http_code}' --max-time 5 "$BASE/health" || true)"
    if [ "$code" = "200" ] && grep -q '"status":"ok"' "$body"; then
      echo "health ok $BASE/health"
      return 0
    fi
    sleep 2
  done
  echo "ALERTA: /health no respondió status=ok en $BASE" >&2
  echo "último código: ${code:-sin respuesta}" >&2
  cat "$body" >&2 || true
  return 1
}

pedir() {
  local method="$1"
  local url="$2"
  shift 2
  curl -sS "${curl_tls_args[@]}" -o "$body" -w '%{http_code}' --max-time 20 -X "$method" "$url" "$@"
}

cookie_https_lleva_secure() {
  SMOKE_HDR="$hdr" python3 - <<'PY'
import os, sys
text = open(os.environ["SMOKE_HDR"], encoding="utf-8", errors="replace").read()
lines = [ln for ln in text.splitlines() if ln.lower().startswith("set-cookie:")]
if not lines:
    sys.exit(1)
for ln in lines:
    attrs = [p.strip().lower() for p in ln.split(":", 1)[1].split(";")][1:]
    if "secure" in attrs:
        sys.exit(0)
sys.exit(1)
PY
}

esperar_health

payload="$(SMOKE_USER="$USER_NAME" CAMPUS_DEV_PASSWORD="$CAMPUS_DEV_PASSWORD" python3 -c 'import json,os; print(json.dumps({"usuario": os.environ["SMOKE_USER"], "clave": os.environ["CAMPUS_DEV_PASSWORD"]}))')"

login_args=(-H 'Content-Type: application/json' -c "$jar" --data "$payload")
if [[ "$BASE" == https://* ]]; then
  login_args+=(-D "$hdr")
fi
code="$(pedir POST "$BASE/api/v1/sesion" "${login_args[@]}" || true)"
if [ "$code" != "200" ]; then
  echo "ALERTA: login de $USER_NAME respondió $code" >&2
  cat "$body" >&2 || true
  exit 1
fi
echo "login ok"
if [[ "$BASE" == https://* ]]; then
  if ! cookie_https_lleva_secure; then
    echo "ALERTA: el login por HTTPS no devolvió Set-Cookie con Secure" >&2
    exit 1
  fi
  echo "login cookie Secure"
fi

for ruta in /api/v1/geo/resumen /api/v1/catalogos; do
  code="$(pedir GET "$BASE$ruta" -b "$jar" || true)"
  if [ "$code" != "200" ]; then
    echo "ALERTA: GET $ruta respondió $code" >&2
    cat "$body" >&2 || true
    exit 1
  fi
  echo "lectura ok $ruta"
done

code="$(pedir GET "$BASE/areas-verdes/v1/swagger/index.html" || true)"
case "$AMBIENTE" in
  develop|qa)
    if [ "$code" = "404" ]; then
      echo "ALERTA: Swagger debería estar en $AMBIENTE y respondió 404" >&2
      exit 1
    fi
    echo "swagger visible ($code)"
    ;;
  produccion)
    if [ "$code" != "404" ]; then
      echo "ALERTA: Swagger en producción respondió $code" >&2
      exit 1
    fi
    echo "swagger apagado en produccion"
    ;;
esac

code_yaml="$(pedir GET "$BASE/api/v1/openapi.yaml" || true)"
case "$AMBIENTE" in
  develop|qa)
    if [ "$code_yaml" = "404" ]; then
      echo "ALERTA: openapi.yaml debería estar en $AMBIENTE y respondió 404" >&2
      exit 1
    fi
    echo "openapi.yaml visible ($code_yaml)"
    ;;
  produccion)
    if [ "$code_yaml" != "404" ]; then
      echo "ALERTA: openapi.yaml en producción respondió $code_yaml" >&2
      exit 1
    fi
    echo "openapi.yaml apagado en produccion"
    ;;
esac

echo "smoke ok ($AMBIENTE) $BASE"
