#!/usr/bin/env bash
# Arnés de prueba para verificar el comportamiento del paso «Desplegar»
# de .github/workflows/deploy.yml con simulación de entorno (env -i).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMPDIR="$(mktemp -d)"

cleanup() {
  rm -rf "$TMPDIR"
}
trap cleanup EXIT

BIN_DIR="$TMPDIR/bin"
mkdir -p "$BIN_DIR"

# Stub git
cat > "$BIN_DIR/git" << 'EOF'
#!/usr/bin/env bash
if [ "${1:-}" = "rev-parse" ]; then
  echo "0123456789abcdef0123456789abcdef01234567"
  exit 0
fi
exit 0
EOF
chmod +x "$BIN_DIR/git"

# Extraer el script run del paso Desplegar usando Python
EXTRACTED_SCRIPT="$TMPDIR/step_run.sh"
python3 -c "
import yaml

with open('$ROOT/.github/workflows/deploy.yml') as f:
    wf = yaml.safe_load(f)

run_cmd = None
for step in wf['jobs']['deploy']['steps']:
    if step.get('id') == 'desplegar' or step.get('name') == 'Desplegar':
        run_cmd = step.get('run')
        break

if not run_cmd:
    raise RuntimeError('No se encontró el paso Desplegar en deploy.yml')

with open('$EXTRACTED_SCRIPT', 'w') as out:
    out.write(run_cmd)
"

chmod +x "$EXTRACTED_SCRIPT"

echo "=== Caso 1: push develop sin credenciales (debe salir 0 con ::warning:: y OMITIDO en summary) ==="
OUT1="$TMPDIR/out1.txt"
SUM1="$TMPDIR/sum1.txt"
set +e
out_case1=$(env -i \
  PATH="$BIN_DIR:/usr/bin:/bin" \
  EVENT_NAME="push" \
  AMBIENTE="develop" \
  GITHUB_OUTPUT="$OUT1" \
  GITHUB_STEP_SUMMARY="$SUM1" \
  bash "$EXTRACTED_SCRIPT" 2>&1)
rc1=$?
set -e

if [ "$rc1" -ne 0 ]; then
  echo "ERROR: Caso 1 debió salir 0 pero salió $rc1" >&2
  echo "$out_case1" >&2
  exit 1
fi
if ! [[ "$out_case1" =~ "::warning::" ]]; then
  echo "ERROR: Caso 1 no emitió ::warning::" >&2
  echo "$out_case1" >&2
  exit 1
fi
if ! grep -q "desplegado=false" "$OUT1"; then
  echo "ERROR: Caso 1 no escribió desplegado=false en GITHUB_OUTPUT" >&2
  exit 1
fi
if ! grep -q "OMITIDO" "$SUM1"; then
  echo "ERROR: Caso 1 no incluyó OMITIDO en GITHUB_STEP_SUMMARY" >&2
  exit 1
fi
echo "Caso 1 OK: rc=0, ::warning:: emitido, desplegado=false y summary con OMITIDO."

echo "=== Caso 2: dispatch develop sin credenciales (debe fallar con exit 1 y ::error::) ==="
OUT2="$TMPDIR/out2.txt"
SUM2="$TMPDIR/sum2.txt"
set +e
out_case2=$(env -i \
  PATH="$BIN_DIR:/usr/bin:/bin" \
  EVENT_NAME="workflow_dispatch" \
  AMBIENTE="develop" \
  GITHUB_OUTPUT="$OUT2" \
  GITHUB_STEP_SUMMARY="$SUM2" \
  bash "$EXTRACTED_SCRIPT" 2>&1)
rc2=$?
set -e

if [ "$rc2" -ne 1 ]; then
  echo "ERROR: Caso 2 debió salir 1 pero salió $rc2" >&2
  echo "$out_case2" >&2
  exit 1
fi
if ! [[ "$out_case2" =~ "::error::" ]]; then
  echo "ERROR: Caso 2 no emitió ::error::" >&2
  echo "$out_case2" >&2
  exit 1
fi
echo "Caso 2 OK: rc=1 y ::error:: emitido para dispatch en develop sin credenciales."

echo "=== Caso 3: dispatch qa sin credenciales (debe fallar con exit 1 y ::error::) ==="
OUT3="$TMPDIR/out3.txt"
SUM3="$TMPDIR/sum3.txt"
set +e
out_case3=$(env -i \
  PATH="$BIN_DIR:/usr/bin:/bin" \
  EVENT_NAME="workflow_dispatch" \
  AMBIENTE="qa" \
  GITHUB_OUTPUT="$OUT3" \
  GITHUB_STEP_SUMMARY="$SUM3" \
  bash "$EXTRACTED_SCRIPT" 2>&1)
rc3=$?
set -e

if [ "$rc3" -ne 1 ]; then
  echo "ERROR: Caso 3 debió salir 1 pero salió $rc3" >&2
  echo "$out_case3" >&2
  exit 1
fi
if ! [[ "$out_case3" =~ "::error::" ]]; then
  echo "ERROR: Caso 3 no emitió ::error::" >&2
  echo "$out_case3" >&2
  exit 1
fi
echo "Caso 3 OK: rc=1 y ::error:: emitido para dispatch en qa sin credenciales."

echo "=== Caso 4: produccion sin credenciales (debe fallar con exit 1 y ::error:: en push o dispatch) ==="
OUT4="$TMPDIR/out4.txt"
SUM4="$TMPDIR/sum4.txt"
set +e
out_case4=$(env -i \
  PATH="$BIN_DIR:/usr/bin:/bin" \
  EVENT_NAME="push" \
  AMBIENTE="produccion" \
  GITHUB_OUTPUT="$OUT4" \
  GITHUB_STEP_SUMMARY="$SUM4" \
  bash "$EXTRACTED_SCRIPT" 2>&1)
rc4=$?
set -e

if [ "$rc4" -ne 1 ]; then
  echo "ERROR: Caso 4 debió salir 1 pero salió $rc4" >&2
  echo "$out_case4" >&2
  exit 1
fi
if ! [[ "$out_case4" =~ "::error::" ]]; then
  echo "ERROR: Caso 4 no emitió ::error::" >&2
  echo "$out_case4" >&2
  exit 1
fi
echo "Caso 4 OK: rc=1 y ::error:: emitido para produccion sin credenciales."

echo "=== Caso 5: push qa (tag rc-*) sin credenciales (rc=0, warning y OMITIDO) ==="
OUT5="$TMPDIR/out5.txt"
SUM5="$TMPDIR/sum5.txt"
set +e
out_case5=$(env -i \
  PATH="$BIN_DIR:/usr/bin:/bin" \
  EVENT_NAME="push" \
  AMBIENTE="qa" \
  GITHUB_OUTPUT="$OUT5" \
  GITHUB_STEP_SUMMARY="$SUM5" \
  bash "$EXTRACTED_SCRIPT" 2>&1)
rc5=$?
set -e

if [ "$rc5" -ne 0 ]; then
  echo "ERROR: Caso 5 debió salir 0 pero salió $rc5" >&2
  echo "$out_case5" >&2
  exit 1
fi
if ! [[ "$out_case5" =~ "::warning::" ]]; then
  echo "ERROR: Caso 5 no emitió ::warning::" >&2
  exit 1
fi
if ! grep -q "OMITIDO" "$SUM5"; then
  echo "ERROR: Caso 5 no incluyó OMITIDO en summary" >&2
  exit 1
fi
echo "Caso 5 OK: rc=0, warning y summary con OMITIDO para push qa."

echo "test_ci_deploy_step ok"
