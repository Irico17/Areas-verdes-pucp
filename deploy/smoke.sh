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

: "${CAMPUS_DEV_PASSWORD:?falta CAMPUS_DEV_PASSWORD}"
BASE="${SMOKE_BASE_URL:-http://127.0.0.1:${WEB_PORT:?falta WEB_PORT}}"
USER_NAME="${SMOKE_USER:-coordinacion}"
body="$(mktemp)"
jar="$(mktemp)"
trap 'rm -f "$body" "$jar"' EXIT

esperar_health() {
  local intento=0
  while [ "$intento" -lt 45 ]; do
    intento=$((intento + 1))
    code="$(curl -sS -o "$body" -w '%{http_code}' --max-time 5 "$BASE/health" || true)"
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
  curl -sS -o "$body" -w '%{http_code}' --max-time 20 -X "$method" "$url" "$@"
}

esperar_health

payload="$(SMOKE_USER="$USER_NAME" CAMPUS_DEV_PASSWORD="$CAMPUS_DEV_PASSWORD" python3 -c 'import json,os; print(json.dumps({"usuario": os.environ["SMOKE_USER"], "clave": os.environ["CAMPUS_DEV_PASSWORD"]}))')"

code="$(pedir POST "$BASE/api/v1/sesion" -H 'Content-Type: application/json' -c "$jar" --data "$payload" || true)"
if [ "$code" != "200" ]; then
  echo "ALERTA: login de $USER_NAME respondió $code" >&2
  cat "$body" >&2 || true
  exit 1
fi
echo "login ok"

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

echo "smoke ok ($AMBIENTE) $BASE"
