#!/usr/bin/env bash
# El login por HTTPS exige Set-Cookie con el atributo Secure.
# En HTTP el smoke no pide ese atributo. No llama a DuckDNS ni a Let's Encrypt.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMP="$(mktemp -d)"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT

BIN="$TMP/bin"
mkdir -p "$BIN"
export LOG_FILE="$TMP/curl.log"

cat > "$BIN/curl" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >> "$LOG_FILE"
url=""
out_file=""
dump_file=""
args=("$@")
for ((i=0; i<${#args[@]}; i++)); do
  if [ "${args[i]}" = "-o" ]; then
    out_file="${args[i+1]}"
  fi
  if [ "${args[i]}" = "-D" ]; then
    dump_file="${args[i+1]}"
  fi
  if [[ "${args[i]}" =~ ^https?:// ]]; then
    url="${args[i]}"
  fi
done
write_body() {
  if [ -n "$out_file" ]; then
    printf '%s' "$1" > "$out_file"
  fi
}
if [[ "$url" == *"/health"* ]]; then
  write_body '{"status":"ok"}'
  printf '200'
  exit 0
fi
if [[ "$url" == *"/api/v1/sesion"* ]]; then
  write_body '{"ok":true}'
  if [ -n "$dump_file" ]; then
    case "${COOKIE_MODE:-secure}" in
      plain)
        printf 'Set-Cookie: cv_sesion=prueba; Path=/; HttpOnly\n' > "$dump_file"
        ;;
      none)
        printf 'HTTP/1.1 200 OK\n' > "$dump_file"
        ;;
      *)
        printf 'Set-Cookie: cv_sesion=prueba; Path=/; HttpOnly; Secure\n' > "$dump_file"
        ;;
    esac
  fi
  printf '200'
  exit 0
fi
if [[ "$url" == *"/areas-verdes/v1/swagger/index.html"* ]] || [[ "$url" == *"/api/v1/openapi.yaml"* ]]; then
  write_body 'swagger'
  printf '200'
  exit 0
fi
write_body '{"data":[]}'
printf '200'
exit 0
EOF
chmod +x "$BIN/curl"
export PATH="$BIN:$PATH"

cat > "$TMP/env" << 'EOF'
APP_ENV=develop
POSTGRES_USER=campus
POSTGRES_PASSWORD=campus-develop
POSTGRES_DB=campus_verde_develop
POSTGRES_PORT=5432
API_PORT=8091
WEB_PORT=8088
CAMPUS_DEV_PASSWORD=pando-local
CAMPUS_COOKIE_SECURE=false
CAMPUS_ENV=develop
EOF

echo "=== HTTPS con Set-Cookie Secure ==="
: > "$LOG_FILE"
out_ok="$(COOKIE_MODE=secure SMOKE_ENV_FILE="$TMP/env" SMOKE_BASE_URL="https://verde-pucp.duckdns.org" SMOKE_TLS_RESOLVE="verde-pucp.duckdns.org:443:127.0.0.1" bash "$ROOT/deploy/smoke.sh" develop)"
if ! grep -q "login cookie Secure" <<< "$out_ok"; then
  echo "ERROR: el smoke HTTPS no confirmó la cookie Secure." >&2
  echo "$out_ok" >&2
  exit 1
fi
if ! grep -q -- "--resolve verde-pucp.duckdns.org:443:127.0.0.1" "$LOG_FILE"; then
  echo "ERROR: no se pasó --resolve al curl del smoke HTTPS." >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
if grep -qE '(^| )-k( |$)|--insecure' "$LOG_FILE"; then
  echo "ERROR: el smoke HTTPS usó -k sin SMOKE_INSECURE_TLS=1." >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
echo "HTTPS con Secure: OK"

echo "=== HTTPS sin atributo Secure debe fallar ==="
: > "$LOG_FILE"
set +e
out_bad="$(COOKIE_MODE=plain SMOKE_ENV_FILE="$TMP/env" SMOKE_BASE_URL="https://127.0.0.1:443" SMOKE_INSECURE_TLS=1 bash "$ROOT/deploy/smoke.sh" develop 2>&1)"
rc_bad=$?
set -e
if [ "$rc_bad" -eq 0 ]; then
  echo "ERROR: el smoke HTTPS aceptó una cookie sin Secure." >&2
  exit 1
fi
if ! grep -q "Set-Cookie con Secure" <<< "$out_bad"; then
  echo "ERROR: el fallo no nombra el atributo Secure." >&2
  echo "$out_bad" >&2
  exit 1
fi
if ! grep -qE '(^| )-k( |$)' "$LOG_FILE"; then
  echo "ERROR: SMOKE_INSECURE_TLS=1 no añadió -k." >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
echo "HTTPS sin Secure: OK (falla, como debe)"

echo "=== HTTP no exige Secure ==="
: > "$LOG_FILE"
out_http="$(COOKIE_MODE=plain SMOKE_ENV_FILE="$TMP/env" SMOKE_BASE_URL="http://127.0.0.1:8088" bash "$ROOT/deploy/smoke.sh" develop)"
if ! grep -q "smoke ok" <<< "$out_http"; then
  echo "ERROR: el smoke HTTP cambió de resultado." >&2
  echo "$out_http" >&2
  exit 1
fi
if grep -q "login cookie Secure" <<< "$out_http"; then
  echo "ERROR: el smoke HTTP no debía exigir la cookie Secure." >&2
  exit 1
fi
if grep -qE '(^| )-k( |$)|--resolve' "$LOG_FILE"; then
  echo "ERROR: el smoke HTTP añadió opciones TLS." >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
echo "HTTP sin cambios: OK"

echo "TODAS LAS PRUEBAS DE test_smoke_tls.sh PASARON."
