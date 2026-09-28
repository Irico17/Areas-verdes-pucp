#!/usr/bin/env bash
# scripts/paridad-api.sh: arnés de verificación de paridad entre la API vieja y la nueva.
# Uso: scripts/paridad-api.sh <VIEJA_URL> <NUEVA_URL> <RUTAS_FILE>
# Ejemplo: scripts/paridad-api.sh http://127.0.0.1:8091 http://127.0.0.1:8092 scripts/paridad-rutas.txt

set -euo pipefail

if [ "$#" -lt 3 ]; then
  echo "Uso: $0 <VIEJA_URL> <NUEVA_URL> <RUTAS_FILE>"
  exit 1
fi

VIEJA="${1%/}"
NUEVA="${2%/}"
RUTAS_FILE="$3"

if [ ! -f "$RUTAS_FILE" ]; then
  echo "Error: archivo de rutas '$RUTAS_FILE' no encontrado"
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "Error: jq es requerido para normalizar respuestas JSON"
  exit 1
fi

CAMPUS_DEV_PASSWORD="${CAMPUS_DEV_PASSWORD:-pando-local}"
TMPDIR=$(mktemp -d /tmp/paridad-api-XXXXXX)
trap 'rm -rf "$TMPDIR"' EXIT

echo "=== Arnés de Paridad de API ==="
echo "Vieja: $VIEJA"
echo "Nueva: $NUEVA"
echo "Rutas: $RUTAS_FILE"
echo ""

# Intentar login por rol para generar cookie jars separados
declare -A ROLES_MAP=(
  ["norte"]="norte"
  ["sur"]="sur"
  ["riego"]="riego"
  ["capataz"]="norte"
  ["coordinacion"]="coordinacion"
  ["jefatura"]="jefatura"
  ["admin"]="admin"
)

login_api() {
  local base_url="$1"
  local user="$2"
  local pass="$3"
  local jar_file="$4"

  local code
  code=$(curl -s -o /dev/null -w "%{http_code}" -c "$jar_file" \
    -H "Content-Type: application/json" \
    -d "{\"usuario\":\"$user\",\"clave\":\"$pass\"}" \
    "$base_url/api/v1/sesion" 2>/dev/null || echo "000")

  if [ "$code" != "200" ] && [ "$code" != "201" ]; then
    code=$(curl -s -o /dev/null -w "%{http_code}" -c "$jar_file" \
      -H "Content-Type: application/json" \
      -d "{\"usuario\":\"$user\",\"clave\":\"$pass\"}" \
      "$base_url/areas-verdes/v1/sesion" 2>/dev/null || echo "000")
  fi

  if [ "$code" = "200" ] || [ "$code" = "201" ]; then
    return 0
  fi
  return 1
}

# Inicializar sesiones para cuentas clave
for r in admin coordinacion jefatura norte; do
  login_api "$VIEJA" "$r" "$CAMPUS_DEV_PASSWORD" "$TMPDIR/jar_vieja_${r}.txt" || true
  login_api "$NUEVA" "$r" "$CAMPUS_DEV_PASSWORD" "$TMPDIR/jar_nueva_${r}.txt" || true
done

normalize_response() {
  local in_file="$1"
  local out_file="$2"

  if [ ! -s "$in_file" ]; then
    : > "$out_file"
    return 0
  fi

  if jq -e . "$in_file" >/dev/null 2>&1; then
    # Normalizar JSON con jq -S excluyendo campos volátiles
    jq -S 'walk(if type == "object" then del(.timestamp, .request_id, .requestId, ."x-request-id", .duracion_ms) else . end)' \
      "$in_file" > "$out_file" 2>/dev/null || cp "$in_file" "$out_file"
  else
    tr -d '\r' < "$in_file" > "$out_file"
  fi
}

PASS_COUNT=0
FAIL_COUNT=0
TOTAL_COUNT=0

while IFS= read -r line || [ -n "$line" ]; do
  # Ignorar comentarios y líneas en blanco
  line=$(echo "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
  if [ -z "$line" ] || [[ "$line" =~ ^# ]]; then
    continue
  fi

  read -r rol metodo ruta extra <<< "$line"
  TOTAL_COUNT=$((TOTAL_COUNT + 1))

  # Cookie jar para rol
  JAR_VIEJA_ARG=()
  JAR_NUEVA_ARG=()
  user="${ROLES_MAP[$rol]:-}"
  if [ -n "$user" ]; then
    if [ -f "$TMPDIR/jar_vieja_${user}.txt" ]; then
      JAR_VIEJA_ARG=(-b "$TMPDIR/jar_vieja_${user}.txt")
    fi
    if [ -f "$TMPDIR/jar_nueva_${user}.txt" ]; then
      JAR_NUEVA_ARG=(-b "$TMPDIR/jar_nueva_${user}.txt")
    fi
  fi

  # Cuerpo opcional con token @path/to/body.json
  BODY_ARGS=()
  if [ -n "${extra:-}" ]; then
    if [[ "$extra" =~ ^@ ]]; then
      BODY_PATH="${extra#@}"
      if [ -f "$BODY_PATH" ]; then
        BODY_ARGS=(-H "Content-Type: application/json" --data-binary "@$BODY_PATH")
      else
        echo "  [WARN] archivo de cuerpo no encontrado: $BODY_PATH"
      fi
    fi
  fi

  # 1. Petición a la API vieja
  BODY_VIEJA="$TMPDIR/body_vieja.raw"
  HEAD_VIEJA="$TMPDIR/head_vieja.txt"
  STATUS_VIEJA=$(curl -s -X "$metodo" "${JAR_VIEJA_ARG[@]}" "${BODY_ARGS[@]}" -D "$HEAD_VIEJA" -o "$BODY_VIEJA" -w "%{http_code}" "$VIEJA$ruta")

  NORM_VIEJA="$TMPDIR/body_vieja.norm"
  normalize_response "$BODY_VIEJA" "$NORM_VIEJA"

  # 2. Rutas a verificar en la nueva (para no-GET solo se llama una vez en /areas-verdes/v1)
  RUTAS_NUEVA=()
  if [ "$metodo" != "GET" ]; then
    if [[ "$ruta" =~ ^/api/v1 ]]; then
      RUTAS_NUEVA+=("${ruta/\/api\/v1/\/areas-verdes\/v1}")
    else
      RUTAS_NUEVA+=("$ruta")
    fi
  else
    if [[ "$ruta" =~ ^/api/v1 ]]; then
      RUTAS_NUEVA+=("$ruta")
      RUTAS_NUEVA+=("${ruta/\/api\/v1/\/areas-verdes\/v1}")
    else
      RUTAS_NUEVA+=("$ruta")
    fi
  fi

  ROUTE_OK=true

  for r_nueva in "${RUTAS_NUEVA[@]}"; do
    BODY_NUEVA="$TMPDIR/body_nueva.raw"
    HEAD_NUEVA="$TMPDIR/head_nueva.txt"
    STATUS_NUEVA=$(curl -s -X "$metodo" "${JAR_NUEVA_ARG[@]}" "${BODY_ARGS[@]}" -D "$HEAD_NUEVA" -o "$BODY_NUEVA" -w "%{http_code}" "$NUEVA$r_nueva")

    NORM_NUEVA="$TMPDIR/body_nueva.norm"
    normalize_response "$BODY_NUEVA" "$NORM_NUEVA"

    # Comparar status code
    if [ "$STATUS_VIEJA" != "$STATUS_NUEVA" ]; then
      echo "  [FAIL] $metodo $ruta -> $r_nueva: status code difiere (vieja=$STATUS_VIEJA, nueva=$STATUS_NUEVA)"
      ROUTE_OK=false
      continue
    fi

    # Comparar cuerpo
    if ! diff -u "$NORM_VIEJA" "$NORM_NUEVA" > "$TMPDIR/diff.patch" 2>&1; then
      echo "  [FAIL] $metodo $ruta -> $r_nueva: cuerpo difiere"
      cat "$TMPDIR/diff.patch" | head -n 30
      ROUTE_OK=false
      continue
    fi
  done

  if [ "$ROUTE_OK" = true ]; then
    echo "  [OK] $metodo $ruta (status: $STATUS_VIEJA, prefijos verificados: ${#RUTAS_NUEVA[@]})"
    PASS_COUNT=$((PASS_COUNT + 1))
  else
    FAIL_COUNT=$((FAIL_COUNT + 1))
  fi

done < "$RUTAS_FILE"

echo ""
echo "=== Resumen de Paridad ==="
echo "Total rutas verificadas: $TOTAL_COUNT"
echo "Rutas exitosas:          $PASS_COUNT"
echo "Rutas con diferencias:   $FAIL_COUNT"

if [ "$FAIL_COUNT" -gt 0 ]; then
  echo "Estado: FALLO (se detectaron diferencias de paridad)"
  exit 1
else
  echo "Estado: EXCELENTE (100% paridad)"
  exit 0
fi
