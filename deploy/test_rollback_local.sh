#!/usr/bin/env bash
# Arnés de prueba para verificar el rollback local de imagen api y web.
# Sin dependencias externas, sin red, sin daemon de Docker.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMPDIR="$(mktemp -d)"
LOG_FILE="$TMPDIR/docker.log"

cleanup() {
  rm -rf "$TMPDIR"
  rm -f "$ROOT/deploy/state/develop.prev-image"
  rm -f "$ROOT/deploy/state/develop.prev-image-web"
  rm -f "$ROOT/deploy/state/develop.prev-image"*
}
trap cleanup EXIT

BIN_DIR="$TMPDIR/bin"
mkdir -p "$BIN_DIR"

# Stub docker
cat > "$BIN_DIR/docker" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail

case "${1:-}" in
  image)
    if [ "${2:-}" = "inspect" ]; then
      img="${3:-}"
      format=""
      for arg in "$@"; do
        if [ "$arg" = "{{.Id}}" ]; then
          format="id"
        fi
      done
      case "$img" in
        backend-campus-verde:develop)
          if [ "$format" = "id" ]; then
            echo "sha256:api_old_111111111111"
          fi
          exit 0
          ;;
        frontend-campus-verde:develop)
          if [ "$format" = "id" ]; then
            echo "sha256:web_old_222222222222"
          fi
          exit 0
          ;;
        sha256:api_old_111111111111|sha256:web_old_222222222222|tag-api-custom)
          exit 0
          ;;
        *)
          echo "Error: No such image: $img" >&2
          exit 1
          ;;
      esac
    fi
    ;;
  tag)
    echo "TAG $2 $3" >> "$LOG_FILE"
    exit 0
    ;;
  compose)
    echo "COMPOSE $*" >> "$LOG_FILE"
    exit 0
    ;;
esac
exit 0
EOF
chmod +x "$BIN_DIR/docker"

# Stub curl para smoke test
cat > "$BIN_DIR/curl" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail

done_file="${FAIL_DONE_FILE:-/tmp/fail_done_fallback}"
if [ "${FAIL_SMOKE_ALWAYS:-0}" = "1" ]; then
  for arg in "$@"; do
    if [[ "$arg" == *"/sesion"* ]]; then
      printf '500'
      exit 0
    fi
  done
fi
if [ "${FAIL_LOGIN_ONCE:-0}" = "1" ] && [ ! -f "$done_file" ]; then
  for arg in "$@"; do
    if [[ "$arg" == *"/sesion"* ]]; then
      touch "$done_file"
      printf '500'
      exit 0
    fi
  done
fi

out_file=""
while [ $# -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    out_file="$2"
    shift 2
    continue
  fi
  shift
done

if [ -n "$out_file" ]; then
  echo '{"status":"ok","usuario":"coordinacion"}' > "$out_file"
fi
printf '200'
exit 0
EOF
chmod +x "$BIN_DIR/curl"

export PATH="$BIN_DIR:$PATH"
export LOG_FILE
export FAIL_DONE_FILE="$TMPDIR/fail_done"

echo "=== Prueba 1: record_previous guarda api y web ==="
: > "$LOG_FILE"
bash "$ROOT/deploy/deploy.sh" develop --local

if [ ! -f "$ROOT/deploy/state/develop.prev-image" ]; then
  echo "ERROR: deploy/state/develop.prev-image no existe" >&2
  exit 1
fi
if [ ! -f "$ROOT/deploy/state/develop.prev-image-web" ]; then
  echo "ERROR: deploy/state/develop.prev-image-web no existe" >&2
  exit 1
fi

prev_api="$(cat "$ROOT/deploy/state/develop.prev-image")"
prev_web="$(cat "$ROOT/deploy/state/develop.prev-image-web")"

if [ "$prev_api" != "sha256:api_old_111111111111" ]; then
  echo "ERROR: prev_api inesperado: $prev_api" >&2
  exit 1
fi
if [ "$prev_web" != "sha256:web_old_222222222222" ]; then
  echo "ERROR: prev_web inesperado: $prev_web" >&2
  exit 1
fi
echo "Prueba 1 OK: prev-image y prev-image-web guardados correctamente."

echo "=== Prueba 2: rollback local sin tag retaguea api y web ==="
: > "$LOG_FILE"
bash "$ROOT/deploy/deploy.sh" develop --local --rollback

if ! grep -q "TAG sha256:api_old_111111111111 backend-campus-verde:develop" "$LOG_FILE"; then
  echo "ERROR: No se retagueó la imagen de api" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
if ! grep -q "TAG sha256:web_old_222222222222 frontend-campus-verde:develop" "$LOG_FILE"; then
  echo "ERROR: No se retagueó la imagen de web" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
if ! grep -q "COMPOSE .* up -d --no-build --remove-orphans" "$LOG_FILE"; then
  echo "ERROR: No se ejecutó compose up --no-build" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
echo "Prueba 2 OK: rollback sin tag retagueó api y web y levantó sin rebuild."

echo "=== Prueba 3: rollback local con TAG explícito retaguea api con ese tag y web del estado ==="
: > "$LOG_FILE"
bash "$ROOT/deploy/deploy.sh" develop --local --rollback tag-api-custom

if ! grep -q "TAG tag-api-custom backend-campus-verde:develop" "$LOG_FILE"; then
  echo "ERROR: No se retagueó la imagen de api con el tag explícito" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
if ! grep -q "TAG sha256:web_old_222222222222 frontend-campus-verde:develop" "$LOG_FILE"; then
  echo "ERROR: No se retagueó la imagen de web desde el estado" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
echo "Prueba 3 OK: rollback con tag explícito aplicó a api y tomó web del estado."

echo "=== Prueba 4: rollback falla si falta prev-image-web ==="
rm -f "$ROOT/deploy/state/develop.prev-image-web"
set +e
err_out="$(bash "$ROOT/deploy/deploy.sh" develop --local --rollback tag-api-custom 2>&1)"
rc=$?
set -e
if [ "$rc" -eq 0 ]; then
  echo "ERROR: El rollback debió fallar al faltar prev-image-web" >&2
  exit 1
fi
if ! [[ "$err_out" =~ "no hay imagen anterior de web" ]]; then
  echo "ERROR: Mensaje de error no menciona imagen anterior de web: $err_out" >&2
  exit 1
fi
echo "Prueba 4 OK: falló limpiamente al faltar prev-image-web."

echo "=== Prueba 5: rollback automático en up_local cuando falla smoke ==="
echo "sha256:api_old_111111111111" > "$ROOT/deploy/state/develop.prev-image"
echo "sha256:web_old_222222222222" > "$ROOT/deploy/state/develop.prev-image-web"
: > "$LOG_FILE"
rm -f "$FAIL_DONE_FILE"

set +e
FAIL_LOGIN_ONCE=1 bash "$ROOT/deploy/deploy.sh" develop --local
rc_up=$?
set -e

if [ "$rc_up" -eq 0 ]; then
  echo "ERROR: up_local debió terminar con error tras fallo del smoke inicial" >&2
  exit 1
fi
if ! grep -q "TAG sha256:api_old_111111111111 backend-campus-verde:develop" "$LOG_FILE"; then
  echo "ERROR: Rollback automático no retagueó api" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
if ! grep -q "TAG sha256:web_old_222222222222 frontend-campus-verde:develop" "$LOG_FILE"; then
  echo "ERROR: Rollback automático no retagueó web" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
echo "Prueba 5 OK: rollback automático retagueó api y web tras fallo de smoke."

echo "=== Prueba 6: rollback local directo con fallo de smoke debe terminar con exit code != 0 ==="
echo "sha256:api_old_111111111111" > "$ROOT/deploy/state/develop.prev-image"
echo "sha256:web_old_222222222222" > "$ROOT/deploy/state/develop.prev-image-web"
: > "$LOG_FILE"

set +e
FAIL_SMOKE_ALWAYS=1 bash "$ROOT/deploy/deploy.sh" develop --local --rollback
rc_rb=$?
set -e

if [ "$rc_rb" -eq 0 ]; then
  echo "ERROR: rollback local directo debió terminar con error al fallar smoke" >&2
  exit 1
fi
echo "Prueba 6 OK: rollback directo con fallo de smoke terminó con código de error ($rc_rb != 0)."

echo "=== Prueba 7: up_local con fallo en smoke y fallo en smoke de rollback debe terminar con exit code != 0 ==="
echo "sha256:api_old_111111111111" > "$ROOT/deploy/state/develop.prev-image"
echo "sha256:web_old_222222222222" > "$ROOT/deploy/state/develop.prev-image-web"
: > "$LOG_FILE"

set +e
FAIL_SMOKE_ALWAYS=1 bash "$ROOT/deploy/deploy.sh" develop --local
rc_double_fail=$?
set -e

if [ "$rc_double_fail" -eq 0 ]; then
  echo "ERROR: up_local debió terminar con error cuando fallan ambos smokes" >&2
  exit 1
fi
echo "Prueba 7 OK: up_local con doble fallo de smoke terminó con código de error ($rc_double_fail != 0)."

echo "test_rollback_local ok"
