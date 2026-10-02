#!/usr/bin/env bash
# Arnés de prueba para verificar que los puertos POSTGRES_PORT, API_PORT y WEB_PORT
# se pueden sobrescribir desde el entorno del llamador con la precedencia adecuada:
# Entorno del llamador > archivo .env > example.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMPDIR="$(mktemp -d)"
LOG_FILE="$TMPDIR/calls.log"
HTTP_LOG="$TMPDIR/http.log"
PY_SERVER_PID=""

cleanup() {
  if [ -n "$PY_SERVER_PID" ]; then
    kill "$PY_SERVER_PID" 2>/dev/null || true
    wait "$PY_SERVER_PID" 2>/dev/null || true
  fi
  rm -rf "$TMPDIR"
}
trap cleanup EXIT

BIN_DIR="$TMPDIR/bin"
mkdir -p "$BIN_DIR"

# Stub docker para registrar llamadas y variables de entorno de puertos
cat > "$BIN_DIR/docker" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [ "${1:-}" = "compose" ]; then
  echo "COMPOSE_PG=${POSTGRES_PORT:-default_unset} COMPOSE_API=${API_PORT:-default_unset} COMPOSE_WEB=${WEB_PORT:-default_unset} args=$*" >> "$LOG_FILE"
  exit 0
elif [ "${1:-}" = "exec" ]; then
  echo "EXEC_PG=${POSTGRES_PORT:-default_unset} EXEC_API=${API_PORT:-default_unset} EXEC_WEB=${WEB_PORT:-default_unset} args=$*" >> "$LOG_FILE"
  exit 0
elif [ "${1:-}" = "image" ]; then
  exit 0
fi
exit 0
EOF
chmod +x "$BIN_DIR/docker"

export PATH="$BIN_DIR:$PATH"
export LOG_FILE

echo "=== Prueba 1: Puertos por defecto en deploy.sh (develop, qa, produccion) ==="
: > "$LOG_FILE"
bash "$ROOT/deploy/deploy.sh" develop --local --down
if ! grep -q "COMPOSE_PG=5432 COMPOSE_API=8091 COMPOSE_WEB=8088" "$LOG_FILE"; then
  echo "ERROR: Puertos por defecto de develop incorrectos" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi

: > "$LOG_FILE"
bash "$ROOT/deploy/deploy.sh" qa --local --down
if ! grep -q "COMPOSE_PG=5433 COMPOSE_API=8191 COMPOSE_WEB=8188" "$LOG_FILE"; then
  echo "ERROR: Puertos por defecto de qa incorrectos" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi

: > "$LOG_FILE"
bash "$ROOT/deploy/deploy.sh" produccion --local --down
if ! grep -q "COMPOSE_PG=5434 COMPOSE_API=8291 COMPOSE_WEB=8288" "$LOG_FILE"; then
  echo "ERROR: Puertos por defecto de produccion incorrectos" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
echo "Prueba 1 OK: Puertos por defecto correctos para los 3 ambientes."

echo "=== Prueba 2: Sobrescritura desde el entorno en deploy.sh ==="
: > "$LOG_FILE"
POSTGRES_PORT=5442 API_PORT=8093 WEB_PORT=8089 bash "$ROOT/deploy/deploy.sh" develop --local --down
if ! grep -q "COMPOSE_PG=5442 COMPOSE_API=8093 COMPOSE_WEB=8089" "$LOG_FILE"; then
  echo "ERROR: deploy.sh no respetó puertos del entorno" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
echo "Prueba 2 OK: deploy.sh usó los puertos del entorno (5442/8093/8089)."

echo "=== Prueba 3: deploy.sh develop --local --smoke contra servidor HTTP en puerto alterno ==="
PORT_FILE="$TMPDIR/port.txt"
python3 -c "
import http.server
import socketserver
import json
import sys

class MockHandler(http.server.BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        pass
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.end_headers()
        if self.path == '/health':
            self.wfile.write(b'{\"status\":\"ok\"}')
        else:
            self.wfile.write(b'{\"ok\":true}')
    def do_POST(self):
        self.send_response(200)
        self.send_header('Content-Type', 'application/json')
        self.send_header('Set-Cookie', 'cv_sesion=mock_session; Path=/')
        self.end_headers()
        self.wfile.write(b'{\"usuario\":\"coordinacion\"}')

socketserver.TCPServer.allow_reuse_address = True
server = socketserver.TCPServer(('127.0.0.1', 0), MockHandler)
port = server.server_address[1]
with open('$PORT_FILE', 'w') as f:
    f.write(str(port))
sys.stdout.flush()
server.serve_forever()
" > "$HTTP_LOG" 2>&1 &
PY_SERVER_PID=$!

# Esperar a que el servidor escriba el puerto y esté listo
ALT_PORT=""
for _ in $(seq 1 30); do
  if [ -f "$PORT_FILE" ]; then
    ALT_PORT="$(cat "$PORT_FILE")"
    if [ -n "$ALT_PORT" ] && curl -s "http://127.0.0.1:$ALT_PORT/health" | grep -q '"status":"ok"'; then
      break
    fi
  fi
  sleep 0.1
done

if [ -z "$ALT_PORT" ]; then
  echo "ERROR: Servidor Python no arrancó" >&2
  cat "$HTTP_LOG" >&2
  exit 1
fi

# Ejecutar smoke vía deploy.sh con WEB_PORT alterno
out_smoke="$(WEB_PORT=$ALT_PORT bash "$ROOT/deploy/deploy.sh" develop --local --smoke)"
if ! [[ "$out_smoke" =~ "smoke ok (develop) http://127.0.0.1:$ALT_PORT" ]]; then
  echo "ERROR: deploy.sh --smoke no usó el puerto alterno $ALT_PORT" >&2
  echo "$out_smoke" >&2
  exit 1
fi
echo "Prueba 3 OK: deploy.sh develop --local --smoke pasó contra servidor en puerto $ALT_PORT."

echo "=== Prueba 4: smoke.sh directamente con WEB_PORT alterno ==="
out_direct_smoke="$(WEB_PORT=$ALT_PORT bash "$ROOT/deploy/smoke.sh" develop)"
if ! [[ "$out_direct_smoke" =~ "smoke ok (develop) http://127.0.0.1:$ALT_PORT" ]]; then
  echo "ERROR: smoke.sh no usó el puerto alterno $ALT_PORT" >&2
  echo "$out_direct_smoke" >&2
  exit 1
fi
echo "Prueba 4 OK: smoke.sh usó directamente el puerto alterno."

# Detener el servidor HTTP de pruebas
kill "$PY_SERVER_PID" 2>/dev/null || true
wait "$PY_SERVER_PID" 2>/dev/null || true
PY_SERVER_PID=""

echo "=== Prueba 5: Precedencia en conteos.sh ==="
: > "$LOG_FILE"
# Se ejecuta con CONTAINER para validar que toma los puertos del entorno
POSTGRES_PORT=5442 API_PORT=8093 WEB_PORT=8089 bash "$ROOT/deploy/conteos.sh" develop || true
if ! grep -q "EXEC_PG=5442 EXEC_API=8093 EXEC_WEB=8089" "$LOG_FILE"; then
  echo "ERROR: conteos.sh no preservó puertos del entorno" >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
echo "Prueba 5 OK: conteos.sh preservó puertos del entorno."

echo "=== Prueba 6: Makefile pasa variables de puertos a deploy/deploy.sh ==="
# Verificar que make up con variables en el entorno genera la línea correcta
mk_out="$(POSTGRES_PORT=5442 API_PORT=8093 WEB_PORT=8089 make -n up ENV=develop)"
if ! [[ "$mk_out" =~ POSTGRES_PORT=\"5442\" ]] || ! [[ "$mk_out" =~ API_PORT=\"8093\" ]] || ! [[ "$mk_out" =~ WEB_PORT=\"8089\" ]]; then
  echo "ERROR: make up no pasó las variables de puertos" >&2
  echo "$mk_out" >&2
  exit 1
fi

# Verificar que make smoke con variables en el entorno pasa los puertos
mk_smoke_out="$(WEB_PORT=8089 make -n smoke ENV=develop)"
if ! [[ "$mk_smoke_out" =~ WEB_PORT=\"8089\" ]]; then
  echo "ERROR: make smoke no pasó WEB_PORT" >&2
  echo "$mk_smoke_out" >&2
  exit 1
fi

echo "Prueba 6 OK: Makefile pasa correctamente las variables sobrescritas."

echo "test_puertos_env ok"
