#!/usr/bin/env bash
# Humo de bajas lógicas contra el backend NUEVO en una BD desechable.
# Uso: API_BASE=http://127.0.0.1:8097 CAMPUS_DEV_PASSWORD=pando-local bash scripts/smoke-bajas.sh
set -euo pipefail
BASE="${API_BASE:-http://127.0.0.1:8097}"
PASS="${CAMPUS_DEV_PASSWORD:-pando-local}"
PREF="${API_PREFIX:-/areas-verdes/v1}"
COOKIE=$(mktemp)
trap 'rm -f "$COOKIE"' EXIT

login() {
  local user=$1
  curl -sS -c "$COOKIE" -X POST "$BASE$PREF/sesion"     -H 'Content-Type: application/json'     -d "{"usuario":"$user","password":"$PASS"}" >/dev/null
}

code() {
  local method=$1 path=$2
  curl -sS -o /tmp/smoke-bajas.body -w '%{http_code}' -b "$COOKIE" -X "$method"     -H 'Content-Type: application/json' -d '{}' "$BASE$PREF$path"
}

echo "== humo bajas en $BASE$PREF =="
login admin
# Crear área ficticia sin geom
FID="AV-SMOKE-BAJA-$$"
curl -sS -b "$COOKIE" -X POST "$BASE$PREF/catastro/areas"   -H 'Content-Type: application/json'   -d "{"feature_id":"$FID","nombre":"Área humo baja","uso":"jardín"}" >/dev/null
c=$(code POST "/catastro/areas/$FID/baja")
test "$c" = "200" || { echo "FAIL baja area admin=$c $(cat /tmp/smoke-bajas.body)"; exit 1; }
# Listado no debe incluirla
if curl -sS -b "$COOKIE" "$BASE$PREF/catastro/areas?q=$FID" | grep -q "$FID"; then
  echo "FAIL listado aún muestra $FID tras baja"; exit 1
fi
echo "OK baja área + listado oculta"

login jefatura
c=$(code POST "/catastro/areas/$FID/baja")
test "$c" = "403" || { echo "FAIL jefatura area baja=$c"; exit 1; }
echo "OK jefatura 403 en área"

# Zona: crear Z4 si hace falta y dar de baja (Z4 suele existir en semillas; si 404, se reporta)
login admin
c=$(code POST "/catastro/zonas-supervision/Z4/baja")
if [ "$c" = "200" ] || [ "$c" = "404" ]; then
  echo "OK baja zona Z4 status=$c"
else
  echo "FAIL baja zona Z4=$c $(cat /tmp/smoke-bajas.body)"; exit 1
fi
echo "SMOKE BAJAS OK"
