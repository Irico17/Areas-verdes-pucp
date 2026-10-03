#!/usr/bin/env bash
# Arnés de prueba para verificar deploy/host-deploy.sh sin Docker real ni AWS.
# Valida orden en producción, fallo por snapshot, dry-run, rollback por smoke,
# prohibición de comandos destructivos, rechazo de claves de laboratorio
# y publicación exclusiva de db/api en 127.0.0.1.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMPDIR="$(mktemp -d)"

cleanup() {
  rm -rf "$TMPDIR"
}
trap cleanup EXIT

BIN_DIR="$TMPDIR/bin"
mkdir -p "$BIN_DIR"
export LOG_FILE="$TMPDIR/calls.log"
: > "$LOG_FILE"

# Stub docker
cat > "$BIN_DIR/docker" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail

full_cmd="$*"

# Detección temprana de comandos destructivos prohibidos
if [[ "$full_cmd" == *"down"* ]] && [[ "$full_cmd" == *"-v"* ]]; then
  echo "FORBIDDEN_DOWN_V: $full_cmd" >> "$LOG_FILE"
  echo "ERROR: Prohibido ejecutar down -v sobre datos" >&2
  exit 1
fi
if [[ "$full_cmd" == *"volume"* ]] && [[ "$full_cmd" == *"rm"* ]]; then
  echo "FORBIDDEN_VOLUME_RM: $full_cmd" >> "$LOG_FILE"
  echo "ERROR: Prohibido ejecutar volume rm sobre datos" >&2
  exit 1
fi

case "${1:-}" in
  pull)
    echo "DOCKER_PULL $2" >> "$LOG_FILE"
    exit 0
    ;;
  compose)
    echo "COMPOSE $*" >> "$LOG_FILE"
    for arg in "$@"; do
      if [ "$arg" = "up" ]; then
        echo "ACTION_COMPOSE_UP" >> "$LOG_FILE"
        exit 0
      fi
      if [ "$arg" = "stop" ]; then
        echo "ACTION_COMPOSE_STOP" >> "$LOG_FILE"
        exit 0
      fi
      if [ "$arg" = "ps" ]; then
        echo "mock-legacy-cid"
        exit 0
      fi
    done
    exit 0
    ;;
  exec)
    if [[ "$full_cmd" == *"conteos.sql"* ]] || [[ "$full_cmd" == *"psql"* ]]; then
      if grep -q "ACTION_COMPOSE_UP" "$LOG_FILE" 2>/dev/null; then
        echo "ACTION_CONTEOS_DESPUES" >> "$LOG_FILE"
      else
        echo "ACTION_CONTEOS_ANTES" >> "$LOG_FILE"
      fi
      printf "table_name\tfilas\nareas_verdes\t521\ncuadrillas\t4\nschema_migrations\t48\nsesiones\t1\nusuarios\t6\n"
      exit 0
    fi
    if [[ "$full_cmd" == *"pg_dump"* ]]; then
      echo "ACTION_PG_DUMP" >> "$LOG_FILE"
      echo "MOCK_PGDUMP_BYTES"
      exit 0
    fi
    echo "DOCKER_EXEC $*" >> "$LOG_FILE"
    exit 0
    ;;
  inspect)
    is_health=0
    is_image=0
    for a in "$@"; do
      if [[ "$a" == *"Health"* ]]; then
        is_health=1
      fi
      if [[ "$a" == *"Image"* ]]; then
        is_image=1
      fi
    done
    if [ "$is_health" -eq 1 ]; then
      echo "healthy"
      exit 0
    fi
    if [ "$is_image" -eq 1 ]; then
      echo "ghcr.io/campus-verde/campus-verde-api:prev-sha-mock"
      exit 0
    fi
    echo "healthy"
    exit 0
    ;;
  ps)
    if [ "${SIMULATE_LEGACY:-0}" = "1" ]; then
      echo "campus-legacy-container"
    else
      echo "campus-produccion-db"
    fi
    exit 0
    ;;
esac

exit 0
EOF
chmod +x "$BIN_DIR/docker"

# Stub aws
cat > "$BIN_DIR/aws" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail
echo "AWS $*" >> "$LOG_FILE"
case "${1:-}" in
  ec2)
    case "${2:-}" in
      describe-volumes)
        echo "vol-mock-data-001"
        exit 0
        ;;
      create-snapshot)
        if [ "${FAIL_SNAPSHOT:-0}" = "1" ]; then
          echo "ERROR: Fallo simulado de snapshot" >&2
          exit 1
        fi
        echo "ACTION_SNAPSHOT_EBS" >> "$LOG_FILE"
        echo "snap-mock-data-001"
        exit 0
        ;;
    esac
    ;;
esac
exit 0
EOF
chmod +x "$BIN_DIR/aws"

# Stub curl para IMDSv2 y smoke test
cat > "$BIN_DIR/curl" << 'EOF'
#!/usr/bin/env bash
set -euo pipefail

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

# IMDSv2 token
if [[ "$url" == *"api/token"* ]]; then
  echo "mock-imds-token-12345"
  exit 0
fi

# IMDSv2 instance-id
if [[ "$url" == *"instance-id"* ]]; then
  echo "i-mock0123456789abcdef"
  exit 0
fi

# IMDSv2 placement region
if [[ "$url" == *"region"* ]]; then
  echo "us-east-1"
  exit 0
fi

# Simulación de fallo en smoke test
if [ "${FAIL_SMOKE:-0}" = "1" ]; then
  if [[ "$url" == *"/api/v1/sesion"* ]]; then
    printf '500'
    exit 0
  fi
fi

# Respuestas normales de smoke test
if [[ "$url" == *"/health"* ]]; then
  if [ -n "$out_file" ]; then
    echo '{"status":"ok"}' > "$out_file"
  fi
  printf '200'
  exit 0
fi

if [[ "$url" == *"/api/v1/sesion"* ]]; then
  if [ -n "$out_file" ]; then
    echo '{"status":"ok"}' > "$out_file"
  fi
  if [ -n "$dump_file" ]; then
    if [ "${FAIL_COOKIE_SECURE:-0}" = "1" ]; then
      printf 'Set-Cookie: cv_sesion=prueba; Path=/; HttpOnly\n' > "$dump_file"
    else
      printf 'Set-Cookie: cv_sesion=prueba; Path=/; HttpOnly; Secure\n' > "$dump_file"
    fi
  fi
  printf '200'
  exit 0
fi

if [[ "$url" == *"/api/v1/geo/resumen"* ]] || [[ "$url" == *"/api/v1/catalogos"* ]]; then
  if [ -n "$out_file" ]; then
    echo '{"data":[]}' > "$out_file"
  fi
  printf '200'
  exit 0
fi

if [[ "$url" == *"/areas-verdes/v1/swagger/index.html"* ]]; then
  if [ -n "$out_file" ]; then
    echo 'Not Found' > "$out_file"
  fi
  printf '404'
  exit 0
fi

if [[ "$url" == *"/api/v1/openapi.yaml"* ]]; then
  if [ -n "$out_file" ]; then
    echo 'Not Found' > "$out_file"
  fi
  printf '404'
  exit 0
fi

printf '200'
exit 0
EOF
chmod +x "$BIN_DIR/curl"

export PATH="$BIN_DIR:$PATH"

setup_env_produccion() {
  local home="$1"
  local pass="${2:-produccion-valida-2026-demo}"
  local pg_pass="${3:-produccion-valida-2026-demo}"
  mkdir -p "$home"
  cat > "$home/host.env" <<EOF
APP_ENV=produccion
POSTGRES_USER=campus
POSTGRES_PASSWORD=$pg_pass
POSTGRES_DB=campus_verde
POSTGRES_PORT=5432
API_PORT=8091
WEB_PORT=80
CAMPUS_DEV_PASSWORD=$pass
CAMPUS_CORS_ORIGINS=http://127.0.0.1,http://localhost
CAMPUS_COOKIE_SECURE=false
CAMPUS_ENV=production
CAMPUS_DATA_DIR=$home/data
PUBLIC_URL=http://127.0.0.1
CAMPUS_CERTS_DIR=$home/certs-ausentes
LEGACY_PROJECT=campus
LEGACY_COMPOSE=$home/docker-compose.yml
EOF
  chmod 600 "$home/host.env"
}

echo "=== Prueba 1: Orden en producción (conteos antes, pg_dump, snapshot, luego compose up) ==="
HOME1="$TMPDIR/env1"
setup_env_produccion "$HOME1"
: > "$LOG_FILE"
VALID_SHA="0123456789abcdef0123456789abcdef01234567"

CAMPUS_HOME="$HOME1" bash "$ROOT/deploy/host-deploy.sh" produccion --sha "$VALID_SHA"

# Verificar orden en LOG_FILE
line_conteos_antes=$(grep -n "ACTION_CONTEOS_ANTES" "$LOG_FILE" | cut -d: -f1 | head -n 1)
line_pg_dump=$(grep -n "ACTION_PG_DUMP" "$LOG_FILE" | cut -d: -f1 | head -n 1)
line_snapshot=$(grep -n "ACTION_SNAPSHOT_EBS" "$LOG_FILE" | cut -d: -f1 | head -n 1)
line_compose_up=$(grep -n "ACTION_COMPOSE_UP" "$LOG_FILE" | cut -d: -f1 | head -n 1)
line_conteos_despues=$(grep -n "ACTION_CONTEOS_DESPUES" "$LOG_FILE" | cut -d: -f1 | head -n 1)

if [ -z "$line_conteos_antes" ] || [ -z "$line_pg_dump" ] || [ -z "$line_snapshot" ] || [ -z "$line_compose_up" ] || [ -z "$line_conteos_despues" ]; then
  echo "ERROR: Una o más acciones esperadas no se registraron en LOG_FILE." >&2
  cat "$LOG_FILE" >&2
  exit 1
fi

if [ "$line_conteos_antes" -gt "$line_pg_dump" ]; then
  echo "ERROR: Conteos antes ($line_conteos_antes) debe ocurrir antes de pg_dump ($line_pg_dump)." >&2
  exit 1
fi
if [ "$line_pg_dump" -gt "$line_snapshot" ]; then
  echo "ERROR: pg_dump ($line_pg_dump) debe ocurrir antes de snapshot EBS ($line_snapshot)." >&2
  exit 1
fi
if [ "$line_snapshot" -gt "$line_compose_up" ]; then
  echo "ERROR: snapshot EBS ($line_snapshot) debe ocurrir antes de compose up ($line_compose_up)." >&2
  exit 1
fi
if [ "$line_compose_up" -gt "$line_conteos_despues" ]; then
  echo "ERROR: compose up ($line_compose_up) debe ocurrir antes de conteos después ($line_conteos_despues)." >&2
  exit 1
fi

if ! grep -q "DEPLOYED_SHA=${VALID_SHA}" "$HOME1/state/deployed.env"; then
  echo "ERROR: deployed.env no contiene el SHA correcto." >&2
  cat "$HOME1/state/deployed.env" >&2
  exit 1
fi
echo "Prueba 1 OK: orden estricto de seguridad verificado."

echo "=== Prueba 2: Fallo si no hay snapshot EBS en producción (salvo DEPLOY_SIN_SNAPSHOT=1) ==="
HOME2="$TMPDIR/env2"
setup_env_produccion "$HOME2"
: > "$LOG_FILE"

set +e
out_fail_snap=$(CAMPUS_HOME="$HOME2" FAIL_SNAPSHOT=1 bash "$ROOT/deploy/host-deploy.sh" produccion --sha "$VALID_SHA" 2>&1)
rc_snap=$?
set -e

if [ "$rc_snap" -eq 0 ]; then
  echo "ERROR: El despliegue debió fallar por snapshot fallido pero salió con código 0." >&2
  exit 1
fi
if grep -q "ACTION_COMPOSE_UP" "$LOG_FILE"; then
  echo "ERROR: compose up se ejecutó a pesar del fallo en snapshot EBS." >&2
  exit 1
fi
if ! [[ "$out_fail_snap" =~ "DEPLOY_SIN_SNAPSHOT=1" ]]; then
  echo "ERROR: Mensaje de error no menciona DEPLOY_SIN_SNAPSHOT=1." >&2
  echo "$out_fail_snap" >&2
  exit 1
fi

# Ahora probar con DEPLOY_SIN_SNAPSHOT=1: debe continuar con advertencia
: > "$LOG_FILE"
CAMPUS_HOME="$HOME2" FAIL_SNAPSHOT=1 DEPLOY_SIN_SNAPSHOT=1 bash "$ROOT/deploy/host-deploy.sh" produccion --sha "$VALID_SHA"
if ! grep -q "ACTION_COMPOSE_UP" "$LOG_FILE"; then
  echo "ERROR: Con DEPLOY_SIN_SNAPSHOT=1 debió permitirse compose up." >&2
  exit 1
fi
echo "Prueba 2 OK: rechazo seguro por falta de snapshot y excepción documentada DEPLOY_SIN_SNAPSHOT=1 funcionan."

echo "=== Prueba 3: --dry-run sin efectos reales ==="
HOME3="$TMPDIR/env3"
setup_env_produccion "$HOME3"
: > "$LOG_FILE"

CAMPUS_HOME="$HOME3" bash "$ROOT/deploy/host-deploy.sh" produccion --sha "$VALID_SHA" --dry-run

if grep -q "ACTION_COMPOSE_UP" "$LOG_FILE" || grep -q "ACTION_SNAPSHOT_EBS" "$LOG_FILE" || grep -q "ACTION_PG_DUMP" "$LOG_FILE"; then
  echo "ERROR: --dry-run ejecutó acciones reales en el sistema." >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
echo "Prueba 3 OK: --dry-run se ejecutó sin mutaciones ni llamadas a docker/aws."

echo "=== Prueba 4: Rollback automático cuando smoke falla ==="
HOME4="$TMPDIR/env4"
setup_env_produccion "$HOME4"
: > "$LOG_FILE"

set +e
out_smoke_fail=$(CAMPUS_HOME="$HOME4" FAIL_SMOKE=1 bash "$ROOT/deploy/host-deploy.sh" produccion --sha "$VALID_SHA" 2>&1)
rc_smoke=$?
set -e

if [ "$rc_smoke" -eq 0 ]; then
  echo "ERROR: Despliegue con smoke fallido debió salir con código 1." >&2
  exit 1
fi

# Debe haberse ejecutado rollback automático a imágenes previas
if ! grep -q "Rollback automático a imágenes previas" <<< "$out_smoke_fail"; then
  echo "ERROR: No se detectó mensaje de rollback automático tras fallo de smoke." >&2
  echo "$out_smoke_fail" >&2
  exit 1
fi
echo "Prueba 4 OK: rollback automático ejecutado correctamente al fallar el smoke test."

echo "=== Prueba 5: Ningún down -v ni volume rm ni borrado de datos ==="
if grep -q "FORBIDDEN" "$LOG_FILE"; then
  echo "ERROR: Se registraron comandos destructivos prohibidos en LOG_FILE." >&2
  grep "FORBIDDEN" "$LOG_FILE" >&2
  exit 1
fi
echo "Prueba 5 OK: no se ejecutó ningún down -v ni volume rm sobre datos."

echo "=== Prueba 6: Rechazo de claves de laboratorio y cortas en producción ==="
for bad_pass in "pando-local" "campus-lab" "corta-123"; do
  HOME6="$TMPDIR/env6_${bad_pass//[^a-zA-Z0-9]/_}"
  setup_env_produccion "$HOME6" "$bad_pass" "produccion-valida-2026-demo"
  : > "$LOG_FILE"

  set +e
  out_pass=$(CAMPUS_HOME="$HOME6" bash "$ROOT/deploy/host-deploy.sh" produccion --sha "$VALID_SHA" 2>&1)
  rc_pass=$?
  set -e

  if [ "$rc_pass" -eq 0 ]; then
    echo "ERROR: Se aceptó CAMPUS_DEV_PASSWORD='$bad_pass' en producción." >&2
    exit 1
  fi
  if grep -q "ACTION_COMPOSE_UP" "$LOG_FILE"; then
    echo "ERROR: compose up se ejecutó a pesar de clave inválida $bad_pass." >&2
    exit 1
  fi
done

# Clave de Postgres de laboratorio
HOME6_PG="$TMPDIR/env6_pg_bad"
setup_env_produccion "$HOME6_PG" "produccion-valida-2026-demo" "pando-local"
set +e
out_pg=$(CAMPUS_HOME="$HOME6_PG" bash "$ROOT/deploy/host-deploy.sh" produccion --sha "$VALID_SHA" 2>&1)
rc_pg=$?
set -e
if [ "$rc_pg" -eq 0 ]; then
  echo "ERROR: Se aceptó POSTGRES_PASSWORD='pando-local' en producción." >&2
  exit 1
fi
echo "Prueba 6 OK: rechazo temprano de claves de laboratorio y contraseñas débiles en producción."

echo "=== Prueba 7: db y api no se publican en 0.0.0.0 en compose.host.yml ==="
python3 - << 'PY'
import sys
from pathlib import Path

compose_file = Path("deploy/compose.host.yml")
text = compose_file.read_text(encoding="utf-8")

# Verificar presencia de 127.0.0.1 en puertos de db y api
lines = text.splitlines()
in_db = False
in_api = False
in_web = False

for line in lines:
    stripped = line.strip()
    if stripped == "db:":
        in_db = True
        in_api = False
        in_web = False
    elif stripped == "api:":
        in_db = False
        in_api = True
        in_web = False
    elif stripped == "web:":
        in_db = False
        in_api = False
        in_web = True
    elif stripped in ("volumes:", "networks:"):
        in_db = False
        in_api = False
        in_web = False

    if in_db and "5432" in stripped and stripped.startswith("-"):
        if "127.0.0.1:" not in stripped:
            sys.exit(f"ERROR: db no está ligada estrictamente a 127.0.0.1: {stripped}")
        if "0.0.0.0" in stripped:
            sys.exit(f"ERROR: db está expuesta en 0.0.0.0: {stripped}")
    if in_api and "8091" in stripped and stripped.startswith("-"):
        if "127.0.0.1:" not in stripped:
            sys.exit(f"ERROR: api no está ligada estrictamente a 127.0.0.1: {stripped}")
        if "0.0.0.0" in stripped:
            sys.exit(f"ERROR: api está expuesta en 0.0.0.0: {stripped}")
    if in_web and "80" in stripped and stripped.startswith("-"):
        if "0.0.0.0:" not in stripped:
            sys.exit(f"ERROR: web debería estar publicada en 0.0.0.0: {stripped}")

# Verificar bind mounts de datos
if "${CAMPUS_DATA_DIR}/pg:/var/lib/postgresql/data" not in text:
    sys.exit("ERROR: compose.host.yml no monta ${CAMPUS_DATA_DIR}/pg en db")
if "${CAMPUS_DATA_DIR}/app:/data" not in text:
    sys.exit("ERROR: compose.host.yml no monta ${CAMPUS_DATA_DIR}/app en api")

# Verificar eliminación de build
if "build: !reset null" not in text:
    sys.exit("ERROR: compose.host.yml no elimina el build con !reset null")

print("compose.host.yml verificado: db y api estrictamente en 127.0.0.1, web en 0.0.0.0, bind mounts y build reseteado.")
PY
echo "Prueba 7 OK: validación de red y configuración de compose.host.yml exitosa."

echo "=== Prueba 8: sin PEM el dry-run no incluye compose.tls.yml ==="
HOME8="$TMPDIR/env8"
setup_env_produccion "$HOME8"
: > "$LOG_FILE"
out_sin_pem="$(CAMPUS_HOME="$HOME8" bash "$ROOT/deploy/host-deploy.sh" produccion --sha "$VALID_SHA" --dry-run)"
if grep -q "compose.tls.yml" <<< "$out_sin_pem"; then
  echo "ERROR: sin PEM el dry-run incluyó compose.tls.yml." >&2
  echo "$out_sin_pem" >&2
  exit 1
fi
if ! grep -q "en http://127.0.0.1:80" <<< "$out_sin_pem"; then
  echo "ERROR: sin PEM el smoke del dry-run no sigue en HTTP." >&2
  echo "$out_sin_pem" >&2
  exit 1
fi
echo "Prueba 8 OK: sin PEM el compose y el smoke siguen en HTTP."

echo "=== Prueba 9: con PEM de prueba el dry-run incluye TLS y el smoke HTTPS ==="
HOME9="$TMPDIR/env9"
setup_env_produccion "$HOME9"
mkdir -p "$HOME9/certs"
printf 'certificado-de-prueba\n' > "$HOME9/certs/fullchain.pem"
printf 'clave-de-prueba\n' > "$HOME9/certs/privkey.pem"
printf 'CAMPUS_CERTS_DIR=%s\n' "$HOME9/certs" >> "$HOME9/host.env"
: > "$LOG_FILE"
out_con_pem="$(CAMPUS_HOME="$HOME9" bash "$ROOT/deploy/host-deploy.sh" produccion --sha "$VALID_SHA" --dry-run)"
if ! grep -q "compose.tls.yml" <<< "$out_con_pem"; then
  echo "ERROR: con PEM el dry-run no incluyó compose.tls.yml." >&2
  echo "$out_con_pem" >&2
  exit 1
fi
if ! grep -q "https://verde-pucp.duckdns.org" <<< "$out_con_pem"; then
  echo "ERROR: con PEM el dry-run no anuncia el smoke HTTPS con el nombre del certificado." >&2
  echo "$out_con_pem" >&2
  exit 1
fi
if grep -q "ACTION_COMPOSE_UP" "$LOG_FILE"; then
  echo "ERROR: el dry-run con PEM ejecutó compose up." >&2
  exit 1
fi
echo "Prueba 9 OK: con PEM el dry-run anuncia TLS y no toca el sistema."

echo "=== Prueba 10: con PEM el despliegue usa compose.tls.yml y el smoke HTTPS pasa ==="
HOME10="$TMPDIR/env10"
setup_env_produccion "$HOME10"
mkdir -p "$HOME10/certs"
printf 'certificado-de-prueba\n' > "$HOME10/certs/fullchain.pem"
printf 'clave-de-prueba\n' > "$HOME10/certs/privkey.pem"
printf 'CAMPUS_CERTS_DIR=%s\n' "$HOME10/certs" >> "$HOME10/host.env"
printf 'CAMPUS_COOKIE_SECURE=true\n' >> "$HOME10/host.env"
: > "$LOG_FILE"
CAMPUS_HOME="$HOME10" bash "$ROOT/deploy/host-deploy.sh" produccion --sha "$VALID_SHA"
if ! grep -q "compose.tls.yml" "$LOG_FILE"; then
  echo "ERROR: el compose registrado no incluyó compose.tls.yml." >&2
  cat "$LOG_FILE" >&2
  exit 1
fi
echo "Prueba 10 OK: el despliegue con PEM incluye TLS y el smoke HTTPS acepta la cookie Secure."

echo "=== Prueba 11: con PEM, si la cookie no es Secure, hay rollback ==="
HOME11="$TMPDIR/env11"
setup_env_produccion "$HOME11"
mkdir -p "$HOME11/certs"
printf 'certificado-de-prueba\n' > "$HOME11/certs/fullchain.pem"
printf 'clave-de-prueba\n' > "$HOME11/certs/privkey.pem"
printf 'CAMPUS_CERTS_DIR=%s\n' "$HOME11/certs" >> "$HOME11/host.env"
: > "$LOG_FILE"
set +e
out_cookie="$(CAMPUS_HOME="$HOME11" FAIL_COOKIE_SECURE=1 bash "$ROOT/deploy/host-deploy.sh" produccion --sha "$VALID_SHA" 2>&1)"
rc_cookie=$?
set -e
if [ "$rc_cookie" -eq 0 ]; then
  echo "ERROR: el despliegue debió fallar cuando la cookie HTTPS no trae Secure." >&2
  exit 1
fi
if ! grep -q "Rollback automático a imágenes previas" <<< "$out_cookie"; then
  echo "ERROR: sin cookie Secure no hubo rollback." >&2
  echo "$out_cookie" >&2
  exit 1
fi
echo "Prueba 11 OK: cookie HTTPS sin Secure dispara el rollback."

echo "=== Prueba 12: docker compose config, con y sin el override TLS ==="
REAL_DOCKER="/usr/bin/docker"
if [ ! -x "$REAL_DOCKER" ] || ! "$REAL_DOCKER" info >/dev/null 2>&1; then
  echo "OMITIDO: no hay Docker real para docker compose config."
else
  cfg="$TMPDIR/compose-env"
  certs_ok="$TMPDIR/certs-ok"
  mkdir -p "$certs_ok"
  printf 'certificado-de-prueba\n' > "$certs_ok/fullchain.pem"
  printf 'clave-de-prueba\n' > "$certs_ok/privkey.pem"
  cat > "$cfg" <<EOF
APP_ENV=produccion
POSTGRES_USER=campus
POSTGRES_PASSWORD=produccion-valida-2026-demo
POSTGRES_DB=campus_verde
POSTGRES_PORT=5432
API_PORT=8091
WEB_PORT=80
CAMPUS_DEV_PASSWORD=produccion-valida-2026-demo
CAMPUS_COOKIE_SECURE=false
CAMPUS_CORS_ORIGINS=http://127.0.0.1
CAMPUS_ENV=production
CAMPUS_DATA_DIR=$TMPDIR/data-compose
API_IMAGE=example.invalid/campus-verde-api:prueba
WEB_IMAGE=example.invalid/campus-verde-web:prueba
CAMPUS_CERTS_DIR=$certs_ok
EOF
  sin="$TMPDIR/compose-sin.json"
  con="$TMPDIR/compose-con.json"
  "$REAL_DOCKER" compose \
    -f "$ROOT/deploy/compose.yml" \
    -f "$ROOT/deploy/compose.produccion.yml" \
    -f "$ROOT/deploy/compose.host.yml" \
    --env-file "$cfg" \
    config --format json > "$sin"
  "$REAL_DOCKER" compose \
    -f "$ROOT/deploy/compose.yml" \
    -f "$ROOT/deploy/compose.produccion.yml" \
    -f "$ROOT/deploy/compose.host.yml" \
    -f "$ROOT/deploy/compose.tls.yml" \
    --env-file "$cfg" \
    config --format json > "$con"
  python3 - "$sin" "$con" "$certs_ok" <<'PY'
import json, sys
sin_path, con_path, certs = sys.argv[1:]
sin = json.load(open(sin_path, encoding="utf-8"))
con = json.load(open(con_path, encoding="utf-8"))

def puertos(doc, servicio):
    return doc["services"][servicio].get("ports") or []

def publicado(lista, target):
    hits = []
    for p in lista:
        t = str(p.get("target"))
        if t == str(target):
            hits.append(p)
    return hits

def exige_loopback(doc, servicio, target):
    hits = publicado(puertos(doc, servicio), target)
    if not hits:
        sys.exit(f"ERROR: {servicio} no publica {target}")
    for p in hits:
        hip = p.get("host_ip") or ""
        if hip != "127.0.0.1":
            sys.exit(f"ERROR: {servicio}:{target} no está solo en 127.0.0.1 ({p})")
        if "0.0.0.0" in json.dumps(p):
            sys.exit(f"ERROR: {servicio}:{target} menciona 0.0.0.0 ({p})")

for doc in (sin, con):
    exige_loopback(doc, "db", 5432)
    exige_loopback(doc, "api", 8091)

if publicado(puertos(sin, "web"), 443):
    sys.exit("ERROR: sin PEM renderizado aparece el puerto 443")
vols_sin = json.dumps(sin["services"]["web"].get("volumes") or [])
if "/etc/nginx/certs" in vols_sin:
    sys.exit("ERROR: sin el override TLS aparece el montaje de certificados")

if not publicado(puertos(con, "web"), 443):
    sys.exit("ERROR: con el override TLS no aparece el puerto 443")
vols_con = json.dumps(con["services"]["web"].get("volumes") or [])
if "/etc/nginx/certs" not in vols_con or certs not in vols_con:
    sys.exit(f"ERROR: el montaje de certificados no apunta al temporal: {vols_con}")
cookie = con["services"]["api"]["environment"].get("CAMPUS_COOKIE_SECURE")
if cookie != "true":
    sys.exit(f"ERROR: con TLS la API no recibe CAMPUS_COOKIE_SECURE=true ({cookie})")
cookie_sin = sin["services"]["api"]["environment"].get("CAMPUS_COOKIE_SECURE")
if cookie_sin != "false":
    sys.exit(f"ERROR: sin TLS la cookie dejó de ser false ({cookie_sin})")
print("compose config: sin TLS no hay 443 ni montaje; con TLS sí; db y api siguen en 127.0.0.1.")
PY
  echo "Prueba 12 OK: docker compose config distingue el override TLS."
fi

echo "TODAS LAS PRUEBAS DE test_host_deploy.sh PASARON SATISFACTORIAMENTE."
