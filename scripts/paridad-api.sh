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

# Mapeo de roles a usuarios semilla
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

  local intentos=0
  while [ "$intentos" -lt 3 ]; do
    local code
    code=$(curl -s -o /dev/null -w "%{http_code}" -c "$jar_file" \
      -H "Content-Type: application/json" \
      -d "{\"usuario\":\"$user\",\"clave\":\"$pass\"}" \
      "$base_url/api/v1/sesion" 2>/dev/null || echo "000")

    if [ "$code" = "429" ]; then
      echo "  [INFO] 429 rate limit alcanzado en login, esperando 10s..."
      sleep 10
      intentos=$((intentos + 1))
      continue
    fi

    if [ "$code" = "200" ] || [ "$code" = "201" ]; then
      return 0
    fi

    # Intentar en prefijo alternativo
    code=$(curl -s -o /dev/null -w "%{http_code}" -c "$jar_file" \
      -H "Content-Type: application/json" \
      -d "{\"usuario\":\"$user\",\"clave\":\"$pass\"}" \
      "$base_url/areas-verdes/v1/sesion" 2>/dev/null || echo "000")

    if [ "$code" = "429" ]; then
      echo "  [INFO] 429 rate limit alcanzado en login, esperando 10s..."
      sleep 10
      intentos=$((intentos + 1))
      continue
    fi

    if [ "$code" = "200" ] || [ "$code" = "201" ]; then
      return 0
    fi

    return 1
  done
  return 1
}

# Inicializar sesiones para cuentas clave (falla si algún login no tiene éxito)
for r in admin coordinacion jefatura norte; do
  if ! login_api "$VIEJA" "$r" "$CAMPUS_DEV_PASSWORD" "$TMPDIR/jar_vieja_${r}.txt"; then
    echo "Error: login falló en API vieja para usuario '$r'"
    exit 1
  fi
  if ! login_api "$NUEVA" "$r" "$CAMPUS_DEV_PASSWORD" "$TMPDIR/jar_nueva_${r}.txt"; then
    echo "Error: login falló en API nueva para usuario '$r'"
    exit 1
  fi
done

# Normalizar respuestas JSON eliminando campos volátiles solo a nivel superior (decisión 24)
normalize_response() {
  local in_file="$1"
  local out_file="$2"

  if [ ! -s "$in_file" ]; then
    : > "$out_file"
    return 0
  fi

  if jq -e . "$in_file" >/dev/null 2>&1; then
    jq -S 'if type == "object" then del(.timestamp, .request_id, .requestId, ."x-request-id", .duracion_ms) else . end' \
      "$in_file" > "$out_file" 2>/dev/null || cp "$in_file" "$out_file"
  else
    tr -d '\r' < "$in_file" > "$out_file"
  fi
}

# Normalizar subconjunto de cabeceras relevantes y enmascarar token de sesión (decisión 3)
normalize_headers() {
  local in_file="$1"
  local out_file="$2"

  if [ ! -s "$in_file" ]; then
    : > "$out_file"
    return 0
  fi

  awk '
    BEGIN { IGNORECASE=1 }
    /^[ \t\r]*$/ { next }
    /^HTTP\// { next }
    {
      colon = index($0, ":")
      if (colon == 0) next
      name = tolower(substr($0, 1, colon - 1))
      val = substr($0, colon + 1)
      gsub(/^[ \t\r\n]+|[ \t\r\n]+$/, "", name)
      gsub(/^[ \t\r\n]+|[ \t\r\n]+$/, "", val)

      if (name == "content-type" || name == "cache-control" || name == "content-disposition" || name ~ /^access-control-/ || name == "set-cookie") {
        if (name == "set-cookie") {
          gsub(/cv_sesion=[^; \t\r\n]+/, "cv_sesion=***MASKED***", val)
        }
        print name ": " val
      }
    }
  ' "$in_file" | sort > "$out_file"
}

PASS_COUNT=0
FAIL_COUNT=0
TOTAL_COUNT=0

# Paso de verificación de paridad de login (decisión 3)
echo "=== Verificando paridad de login (POST /sesion) ==="
LOGIN_BODY="{\"usuario\":\"admin\",\"clave\":\"$CAMPUS_DEV_PASSWORD\"}"
LOGIN_HEAD_VIEJA="$TMPDIR/login_head_vieja.txt"
LOGIN_BODY_VIEJA="$TMPDIR/login_body_vieja.raw"
STATUS_LOGIN_VIEJA=$(curl -s -X POST -H "Content-Type: application/json" -d "$LOGIN_BODY" -D "$LOGIN_HEAD_VIEJA" -o "$LOGIN_BODY_VIEJA" -w "%{http_code}" "$VIEJA/api/v1/sesion")

NORM_LOGIN_VIEJA="$TMPDIR/login_body_vieja.norm"
NORM_LOGIN_HEAD_VIEJA="$TMPDIR/login_head_vieja.norm"
normalize_response "$LOGIN_BODY_VIEJA" "$NORM_LOGIN_VIEJA"
normalize_headers "$LOGIN_HEAD_VIEJA" "$NORM_LOGIN_HEAD_VIEJA"

TOTAL_COUNT=$((TOTAL_COUNT + 1))
LOGIN_OK=true

for r_login in "/api/v1/sesion" "/areas-verdes/v1/sesion"; do
  LOGIN_HEAD_NUEVA="$TMPDIR/login_head_nueva.txt"
  LOGIN_BODY_NUEVA="$TMPDIR/login_body_nueva.raw"
  STATUS_LOGIN_NUEVA=$(curl -s -X POST -H "Content-Type: application/json" -d "$LOGIN_BODY" -D "$LOGIN_HEAD_NUEVA" -o "$LOGIN_BODY_NUEVA" -w "%{http_code}" "$NUEVA$r_login")

  NORM_LOGIN_NUEVA="$TMPDIR/login_body_nueva.norm"
  NORM_LOGIN_HEAD_NUEVA="$TMPDIR/login_head_nueva.norm"
  normalize_response "$LOGIN_BODY_NUEVA" "$NORM_LOGIN_NUEVA"
  normalize_headers "$LOGIN_HEAD_NUEVA" "$NORM_LOGIN_HEAD_NUEVA"

  if [ "$STATUS_LOGIN_VIEJA" != "$STATUS_LOGIN_NUEVA" ]; then
    echo "  [FAIL] POST /sesion -> $r_login: status code difiere (vieja=$STATUS_LOGIN_VIEJA, nueva=$STATUS_LOGIN_NUEVA)"
    LOGIN_OK=false
    continue
  fi

  if ! diff -u "$NORM_LOGIN_VIEJA" "$NORM_LOGIN_NUEVA" > "$TMPDIR/diff_login_body.patch" 2>&1; then
    echo "  [FAIL] POST /sesion -> $r_login: cuerpo difiere"
    head -n 30 "$TMPDIR/diff_login_body.patch"
    LOGIN_OK=false
    continue
  fi

  if ! diff -u "$NORM_LOGIN_HEAD_VIEJA" "$NORM_LOGIN_HEAD_NUEVA" > "$TMPDIR/diff_login_head.patch" 2>&1; then
    echo "  [FAIL] POST /sesion -> $r_login: cabeceras (Set-Cookie) difieren"
    head -n 30 "$TMPDIR/diff_login_head.patch"
    LOGIN_OK=false
    continue
  fi
done

if [ "$LOGIN_OK" = true ]; then
  echo "  [OK] POST /sesion (status: $STATUS_LOGIN_VIEJA, prefijos verificados: 2)"
  PASS_COUNT=$((PASS_COUNT + 1))
else
  FAIL_COUNT=$((FAIL_COUNT + 1))
fi

echo ""
echo "=== Verificando rutas del archivo ==="

while IFS= read -r line || [ -n "$line" ]; do
  # Ignorar comentarios y líneas en blanco
  line=$(echo "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
  if [ -z "$line" ] || [[ "$line" =~ ^# ]]; then
    continue
  fi

  read -r rol metodo ruta extra1 extra2 <<< "$line"
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

  # Parsear argumentos opcionales: status esperado y cuerpo @body
  expected_status=""
  BODY_ARGS=()
  for extra in "${extra1:-}" "${extra2:-}"; do
    if [[ "$extra" =~ ^[0-9]{3}$ ]]; then
      expected_status="$extra"
    elif [[ "$extra" =~ ^@ ]]; then
      BODY_PATH="${extra#@}"
      if [ -f "$BODY_PATH" ]; then
        BODY_ARGS=(-H "Content-Type: application/json" --data-binary "@$BODY_PATH")
      else
        echo "  [WARN] archivo de cuerpo no encontrado: $BODY_PATH"
      fi
    fi
  done

  # 1. Petición a la API vieja
  BODY_VIEJA="$TMPDIR/body_vieja.raw"
  HEAD_VIEJA="$TMPDIR/head_vieja.txt"
  STATUS_VIEJA=$(curl -s -X "$metodo" "${JAR_VIEJA_ARG[@]}" "${BODY_ARGS[@]}" -D "$HEAD_VIEJA" -o "$BODY_VIEJA" -w "%{http_code}" "$VIEJA$ruta")

  NORM_VIEJA="$TMPDIR/body_vieja.norm"
  NORM_HEAD_VIEJA="$TMPDIR/head_vieja.norm"
  normalize_response "$BODY_VIEJA" "$NORM_VIEJA"
  normalize_headers "$HEAD_VIEJA" "$NORM_HEAD_VIEJA"

  if [ -n "$expected_status" ] && [ "$STATUS_VIEJA" != "$expected_status" ]; then
    echo "  [FAIL] $metodo $ruta: API vieja retornó $STATUS_VIEJA, se esperaba $expected_status"
    FAIL_COUNT=$((FAIL_COUNT + 1))
    continue
  fi

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
    NORM_HEAD_NUEVA="$TMPDIR/head_nueva.norm"
    normalize_response "$BODY_NUEVA" "$NORM_NUEVA"
    normalize_headers "$HEAD_NUEVA" "$NORM_HEAD_NUEVA"

    if [ -n "$expected_status" ] && [ "$STATUS_NUEVA" != "$expected_status" ]; then
      echo "  [FAIL] $metodo $ruta -> $r_nueva: API nueva retornó $STATUS_NUEVA, se esperaba $expected_status"
      ROUTE_OK=false
      continue
    fi

    # Comparar status code
    if [ "$STATUS_VIEJA" != "$STATUS_NUEVA" ]; then
      echo "  [FAIL] $metodo $ruta -> $r_nueva: status code difiere (vieja=$STATUS_VIEJA, nueva=$STATUS_NUEVA)"
      ROUTE_OK=false
      continue
    fi

    # Comparar cuerpo
    if ! diff -u "$NORM_VIEJA" "$NORM_NUEVA" > "$TMPDIR/diff.patch" 2>&1; then
      echo "  [FAIL] $metodo $ruta -> $r_nueva: cuerpo difiere"
      head -n 30 "$TMPDIR/diff.patch"
      ROUTE_OK=false
      continue
    fi

    # Comparar cabeceras normalizadas
    if ! diff -u "$NORM_HEAD_VIEJA" "$NORM_HEAD_NUEVA" > "$TMPDIR/diff_headers.patch" 2>&1; then
      echo "  [FAIL] $metodo $ruta -> $r_nueva: cabeceras difieren"
      head -n 30 "$TMPDIR/diff_headers.patch"
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
