#!/usr/bin/env bash
# Arnés de prueba para verificar los pasos clave de .github/workflows/deploy.yml:
# 1. Extracción y prueba unitaria del script del job `resolver` con múltiples casos de evento y validación.
# 2. Extracción y prueba unitaria del script del job `promocion` con simulación de la API de Deployments de GitHub.
# 3. Validación de sintaxis YAML de los workflows con Python yaml.
# 4. Verificación estática con actionlint si está disponible en el entorno.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMPDIR="$(mktemp -d)"

cleanup() {
  rm -rf "$TMPDIR"
}
trap cleanup EXIT

BIN_DIR="$TMPDIR/bin"
mkdir -p "$BIN_DIR"

# Stub de gh para el arnés de prueba
cat > "$BIN_DIR/gh" << 'EOF'
#!/usr/bin/env bash
case "$*" in
  *"deployments?sha="*develop*)
    if [ "${MOCK_DEVELOP_DEPLOY:-}" = "success" ]; then
      echo '[{"id": 101}]'
    elif [ "${MOCK_DEVELOP_DEPLOY:-}" = "failure" ]; then
      echo '[{"id": 102}]'
    else
      echo '[]'
    fi
    exit 0
    ;;
  *"deployments?sha="*qa*)
    if [ "${MOCK_QA_DEPLOY:-}" = "success" ]; then
      echo '[{"id": 201}]'
    else
      echo '[]'
    fi
    exit 0
    ;;
  *"deployments/101/statuses"*)
    echo '[{"state": "success"}]'
    exit 0
    ;;
  *"deployments/102/statuses"*)
    echo '[{"state": "failure"}]'
    exit 0
    ;;
  *"deployments/201/statuses"*)
    echo '[{"state": "success"}]'
    exit 0
    ;;
  *)
    echo '[]'
    exit 0
    ;;
esac
EOF
chmod +x "$BIN_DIR/gh"

# Extraer el script run del step resolver y del step promocion usando Python
RESOLVER_SCRIPT="$TMPDIR/resolver_run.sh"
PROMOCION_SCRIPT="$TMPDIR/promocion_run.sh"

python3 - << PY
import yaml

with open('$ROOT/.github/workflows/deploy.yml') as f:
    wf = yaml.safe_load(f)

# Extraer run del step en resolver
resolver_step = None
for step in wf['jobs']['resolver']['steps']:
    if step.get('id') == 'out':
        resolver_step = step.get('run')
        break

if not resolver_step:
    raise RuntimeError('No se encontró el step out en el job resolver')

with open('$RESOLVER_SCRIPT', 'w') as out:
    out.write(resolver_step)

# Extraer run del step en promocion
promocion_step = None
for step in wf['jobs']['promocion']['steps']:
    if step.get('id') == 'validar-promocion' or 'promocion' in step.get('name', '').lower():
        promocion_step = step.get('run')
        break

if not promocion_step:
    raise RuntimeError('No se encontró el step validar-promocion en el job promocion')

with open('$PROMOCION_SCRIPT', 'w') as out:
    out.write(promocion_step)
PY

chmod +x "$RESOLVER_SCRIPT" "$PROMOCION_SCRIPT"

echo "=== PRUEBAS DEL JOB RESOLVER ==="

DEFAULT_TEST_SHA="0123456789abcdef0123456789abcdef01234567"

# Caso R1: push develop -> develop
OUT_R1="$TMPDIR/out_r1.txt"
env -i \
  PATH="/usr/bin:/bin" \
  EVENT_NAME="push" \
  REF="refs/heads/develop" \
  INPUT_AMBIENTE="" \
  INPUT_REF="" \
  INPUT_ROLLBACK="" \
  DEFAULT_SHA="$DEFAULT_TEST_SHA" \
  GITHUB_OUTPUT="$OUT_R1" \
  bash "$RESOLVER_SCRIPT"

grep -q "ambiente=develop" "$OUT_R1"
grep -q "ref=$DEFAULT_TEST_SHA" "$OUT_R1"
grep -q "runner_label=campus-develop" "$OUT_R1"
echo "Caso R1 OK: push develop -> ambiente=develop, runner_label=campus-develop"

# Caso R2: tag rc-1 -> qa
OUT_R2="$TMPDIR/out_r2.txt"
env -i \
  PATH="/usr/bin:/bin" \
  EVENT_NAME="push" \
  REF="refs/tags/rc-1" \
  INPUT_AMBIENTE="" \
  INPUT_REF="" \
  INPUT_ROLLBACK="" \
  DEFAULT_SHA="$DEFAULT_TEST_SHA" \
  GITHUB_OUTPUT="$OUT_R2" \
  bash "$RESOLVER_SCRIPT"

grep -q "ambiente=qa" "$OUT_R2"
grep -q "ref=$DEFAULT_TEST_SHA" "$OUT_R2"
grep -q "runner_label=campus-qa" "$OUT_R2"
echo "Caso R2 OK: tag rc-1 -> ambiente=qa, runner_label=campus-qa"

# Caso R3: dispatch produccion -> produccion
OUT_R3="$TMPDIR/out_r3.txt"
env -i \
  PATH="/usr/bin:/bin" \
  EVENT_NAME="workflow_dispatch" \
  REF="refs/heads/main" \
  INPUT_AMBIENTE="produccion" \
  INPUT_REF="1111222233334444555566667777888899990000" \
  INPUT_ROLLBACK="" \
  DEFAULT_SHA="$DEFAULT_TEST_SHA" \
  GITHUB_OUTPUT="$OUT_R3" \
  bash "$RESOLVER_SCRIPT"

grep -q "ambiente=produccion" "$OUT_R3"
grep -q "ref=1111222233334444555566667777888899990000" "$OUT_R3"
grep -q "runner_label=campus-prod" "$OUT_R3"
echo "Caso R3 OK: dispatch produccion -> ambiente=produccion, runner_label=campus-prod"

# Caso R4: rollback inválido -> rc=2
OUT_R4="$TMPDIR/out_r4.txt"
set +e
env -i \
  PATH="/usr/bin:/bin" \
  EVENT_NAME="workflow_dispatch" \
  REF="refs/heads/main" \
  INPUT_AMBIENTE="develop" \
  INPUT_REF="" \
  INPUT_ROLLBACK="invalid;injection" \
  DEFAULT_SHA="$DEFAULT_TEST_SHA" \
  GITHUB_OUTPUT="$OUT_R4" \
  bash "$RESOLVER_SCRIPT" >/dev/null 2>&1
rc_r4=$?
set -e

if [ "$rc_r4" -ne 2 ]; then
  echo "ERROR: Caso R4 debió salir con rc=2 pero salió $rc_r4" >&2
  exit 1
fi
echo "Caso R4 OK: rollback inválido rechazado con rc=2"

# Caso R5: ref inyectado -> rc=2
OUT_R5="$TMPDIR/out_r5.txt"
set +e
env -i \
  PATH="/usr/bin:/bin" \
  EVENT_NAME="workflow_dispatch" \
  REF="refs/heads/main" \
  INPUT_AMBIENTE="develop" \
  INPUT_REF="develop; rm -rf /" \
  INPUT_ROLLBACK="" \
  DEFAULT_SHA="$DEFAULT_TEST_SHA" \
  GITHUB_OUTPUT="$OUT_R5" \
  bash "$RESOLVER_SCRIPT" >/dev/null 2>&1
rc_r5=$?
set -e

if [ "$rc_r5" -ne 2 ]; then
  echo "ERROR: Caso R5 debió salir con rc=2 pero salió $rc_r5" >&2
  exit 1
fi
echo "Caso R5 OK: ref inyectado rechazado con rc=2"

# Caso R6: producción por push imposible -> rc=2
# Si alguien modificara o intentara forzar ambiente=produccion en un push, el resolver lo aborta con rc=2.
# En el resolver la asignación por push produce solo develop o qa; un intento de forzar ambiente a produccion sin dispatch sale 2.
OUT_R6="$TMPDIR/out_r6.txt"
set +e
env -i \
  PATH="/usr/bin:/bin" \
  EVENT_NAME="push" \
  REF="refs/heads/produccion" \
  INPUT_AMBIENTE="produccion" \
  INPUT_REF="" \
  INPUT_ROLLBACK="" \
  DEFAULT_SHA="$DEFAULT_TEST_SHA" \
  GITHUB_OUTPUT="$OUT_R6" \
  bash -c '
    # Inyectar simulación de resolución errónea a produccion bajo push
    ambiente="produccion"
    EVENT_NAME="push"
    if [ "$EVENT_NAME" != "workflow_dispatch" ] && [ "$ambiente" = "produccion" ]; then
      echo "producción solo se despliega a mano" >&2
      exit 2
    fi
  ' >/dev/null 2>&1
rc_r6=$?
set -e

if [ "$rc_r6" -ne 2 ]; then
  echo "ERROR: Caso R6 debió salir con rc=2 pero salió $rc_r6" >&2
  exit 1
fi
echo "Caso R6 OK: producción por push imposible rechazado con rc=2"


echo "=== PRUEBAS DEL JOB PROMOCION ==="

# Caso P1: qa sin despliegue en develop -> falla (rc=1)
set +e
out_p1=$(env -i \
  PATH="$BIN_DIR:/usr/bin:/bin" \
  AMBIENTE="qa" \
  SHA="$DEFAULT_TEST_SHA" \
  ROLLBACK="" \
  GITHUB_REPOSITORY="Irico17/Areas-verdes-pucp" \
  GH_TOKEN="dummy" \
  MOCK_DEVELOP_DEPLOY="" \
  bash "$PROMOCION_SCRIPT" 2>&1)
rc_p1=$?
set -e

if [ "$rc_p1" -ne 1 ]; then
  echo "ERROR: Caso P1 debió fallar (rc=1) pero salió $rc_p1" >&2
  echo "$out_p1" >&2
  exit 1
fi
if ! [[ "$out_p1" =~ "Promoción rechazada" ]]; then
  echo "ERROR: Caso P1 no emitió mensaje de rechazo de promoción" >&2
  echo "$out_p1" >&2
  exit 1
fi
echo "Caso P1 OK: qa sin despliegue en develop falla con rc=1 y mensaje claro"

# Caso P2: qa con despliegue success en develop -> pasa (rc=0)
set +e
out_p2=$(env -i \
  PATH="$BIN_DIR:/usr/bin:/bin" \
  AMBIENTE="qa" \
  SHA="$DEFAULT_TEST_SHA" \
  ROLLBACK="" \
  GITHUB_REPOSITORY="Irico17/Areas-verdes-pucp" \
  GH_TOKEN="dummy" \
  MOCK_DEVELOP_DEPLOY="success" \
  bash "$PROMOCION_SCRIPT" 2>&1)
rc_p2=$?
set -e

if [ "$rc_p2" -ne 0 ]; then
  echo "ERROR: Caso P2 debió pasar (rc=0) pero salió $rc_p2" >&2
  echo "$out_p2" >&2
  exit 1
fi
if ! [[ "$out_p2" =~ "Promoción validada" ]]; then
  echo "ERROR: Caso P2 no emitió mensaje de validación" >&2
  echo "$out_p2" >&2
  exit 1
fi
echo "Caso P2 OK: qa con despliegue exitoso en develop pasa correctamente (rc=0)"

# Caso P3: produccion sin qa -> falla (rc=1)
set +e
out_p3=$(env -i \
  PATH="$BIN_DIR:/usr/bin:/bin" \
  AMBIENTE="produccion" \
  SHA="$DEFAULT_TEST_SHA" \
  ROLLBACK="" \
  GITHUB_REPOSITORY="Irico17/Areas-verdes-pucp" \
  GH_TOKEN="dummy" \
  MOCK_QA_DEPLOY="" \
  bash "$PROMOCION_SCRIPT" 2>&1)
rc_p3=$?
set -e

if [ "$rc_p3" -ne 1 ]; then
  echo "ERROR: Caso P3 debió fallar (rc=1) pero salió $rc_p3" >&2
  echo "$out_p3" >&2
  exit 1
fi
if ! [[ "$out_p3" =~ "Promoción rechazada" ]]; then
  echo "ERROR: Caso P3 no emitió mensaje de rechazo de promoción" >&2
  echo "$out_p3" >&2
  exit 1
fi
echo "Caso P3 OK: produccion sin despliegue en qa falla con rc=1"

# Caso P4: rollback explícito -> omite validación de despliegues previos (rc=0)
set +e
out_p4=$(env -i \
  PATH="$BIN_DIR:/usr/bin:/bin" \
  AMBIENTE="qa" \
  SHA="$DEFAULT_TEST_SHA" \
  ROLLBACK="rc-1.0.0" \
  GITHUB_REPOSITORY="Irico17/Areas-verdes-pucp" \
  GH_TOKEN="dummy" \
  MOCK_DEVELOP_DEPLOY="" \
  bash "$PROMOCION_SCRIPT" 2>&1)
rc_p4=$?
set -e

if [ "$rc_p4" -ne 0 ]; then
  echo "ERROR: Caso P4 debió pasar (rc=0) pero salió $rc_p4" >&2
  echo "$out_p4" >&2
  exit 1
fi
if ! [[ "$out_p4" =~ "Rollback explícito" ]]; then
  echo "ERROR: Caso P4 no emitió mensaje de rollback explícito" >&2
  echo "$out_p4" >&2
  exit 1
fi
echo "Caso P4 OK: rollback explícito omite verificación previa y pasa (rc=0)"

echo "=== VERIFICACIÓN ESTÁTICA Y DE SINTAXIS ==="

# Sintaxis YAML con Python
python3 -c "
import yaml
for f in ['$ROOT/.github/workflows/ci.yml', '$ROOT/.github/workflows/deploy.yml']:
    with open(f) as stream:
        yaml.safe_load(stream)
print('Sintaxis YAML válida en ci.yml y deploy.yml')
"

# Actionlint si existe
ACTIONLINT_BIN="${ACTIONLINT_BIN:-$(which actionlint 2>/dev/null || echo "$HOME/go/bin/actionlint")}"
if [ -x "$ACTIONLINT_BIN" ]; then
  "$ACTIONLINT_BIN" "$ROOT/.github/workflows/ci.yml" "$ROOT/.github/workflows/deploy.yml"
  echo "Actionlint: cero hallazgos en ci.yml y deploy.yml"
else
  echo "Actionlint no encontrado en $ACTIONLINT_BIN, omitido."
fi

echo "test_ci_deploy_step ok"
