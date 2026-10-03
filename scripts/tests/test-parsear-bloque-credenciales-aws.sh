#!/usr/bin/env bash
# Bloques ficticios. Comprueba la extracción y que el script no imprime valores.
set +x
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCRIPT="$ROOT/scripts/parsear-bloque-credenciales-aws.sh"
ASSERT="$ROOT/scripts/assert-plan-sin-secretos.sh"
TMP="$(mktemp -d)"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT

fail() {
  printf '%s\n' "FALLO: $*" >&2
  exit 1
}

KEY="FAKE""ACCESSKEYID0001"
SECRET="fake""SecretValue000000"
TOKEN="fake""SessionTok/en+000="
OTRA_KEY="OTRA""KEY000000000000"
OTRA_SECRET="otra""secret00000000"
OTRA_TOKEN="otra""token0000000000"

[ "${#KEY}" -ge 16 ] || fail "el access key ficticio es corto"
[ "${#SECRET}" -ge 16 ] || fail "el secret ficticio es corto"
[ "${#TOKEN}" -ge 16 ] || fail "el token ficticio es corto"

ENVFILE="$TMP/github.env"
OUT="$TMP/out"
ERR="$TMP/err"

correr() {
  local bloque="$1"
  shift
  : >"$ENVFILE"
  : >"$OUT"
  : >"$ERR"
  set +e
  env -u AWS_REGION -u AWS_DEFAULT_REGION -u AWS_ACCESS_KEY_ID \
    -u AWS_SECRET_ACCESS_KEY -u AWS_SESSION_TOKEN -u CREDENCIALES_DESDE_BLOQUE \
    "$@" \
    BLOQUE_CREDENCIALES_AWS="$bloque" \
    GITHUB_ENV="$ENVFILE" \
    bash "$SCRIPT" >"$OUT" 2>"$ERR"
  RC=$?
  set -e
}

leer_var() {
  python3 - "$1" "$ENVFILE" <<'PY'
import sys
nombre, path = sys.argv[1], sys.argv[2]
lines = open(path, encoding="utf-8").read().splitlines()
i = 0
found = None
while i < len(lines):
    line = lines[i]
    if line.startswith(nombre + "<<"):
        delim = line.split("<<", 1)[1]
        i += 1
        buf = []
        while i < len(lines) and lines[i] != delim:
            buf.append(lines[i])
            i += 1
        found = "\n".join(buf)
        break
    if line.startswith(nombre + "="):
        found = line.split("=", 1)[1]
        break
    i += 1
if found is None:
    sys.exit(2)
sys.stdout.write(found)
PY
}

salida_solo_mascaras() {
  local archivo
  if [ -s "$ERR" ]; then
    for archivo in "$KEY" "$SECRET" "$TOKEN" "$OTRA_KEY" "$OTRA_SECRET" "$OTRA_TOKEN"; do
      if grep -F -q -- "$archivo" "$ERR"; then
        fail "stderr contiene un valor"
      fi
    done
  fi
  if [ -s "$OUT" ]; then
    if grep -v '^::add-mask::' "$OUT" | grep -q .; then
      fail "stdout tiene líneas que no son ::add-mask::"
    fi
  fi
  if grep -F -q 'export AWS_' "$ENVFILE" || grep -F -q 'BLOQUE_CREDENCIALES' "$ENVFILE"; then
    fail "GITHUB_ENV guardó el bloque en vez de los valores"
  fi
}

esperar_valores() {
  local region="$1"
  [ "$RC" -eq 0 ] || fail "el parser salió $RC"
  salida_solo_mascaras
  [ "$(leer_var AWS_ACCESS_KEY_ID)" = "$KEY" ] || fail "access key id"
  [ "$(leer_var AWS_SECRET_ACCESS_KEY)" = "$SECRET" ] || fail "secret"
  [ "$(leer_var AWS_SESSION_TOKEN)" = "$TOKEN" ] || fail "session token"
  [ "$(leer_var AWS_REGION)" = "$region" ] || fail "región"
  [ "$(leer_var AWS_DEFAULT_REGION)" = "$region" ] || fail "región por defecto"
  [ "$(leer_var CREDENCIALES_DESDE_BLOQUE)" = "1" ] || fail "marca de origen"
  if grep -F -q -- "$OTRA_KEY" "$ENVFILE"; then
    fail "se usó un perfil que no es default"
  fi
}

echo "=== [default] sin espacios alrededor de = ==="
correr "$(printf '[default]\naws_access_key_id=%s\naws_secret_access_key=%s\naws_session_token=%s\nregion=%s\n' "$KEY" "$SECRET" "$TOKEN" "us-west-2")"
esperar_valores "us-west-2"

echo "=== [default] con espacios alrededor de = y CRLF ==="
correr "$(printf '[default]\r\naws_access_key_id = %s\r\naws_secret_access_key = %s\r\naws_session_token = %s\r\nregion = %s\r\n' "$KEY" "$SECRET" "$TOKEN" "us-east-1")"
esperar_valores "us-east-1"

echo "=== export con comillas dobles y CRLF ==="
correr "$(printf 'export AWS_ACCESS_KEY_ID="%s"\r\nexport AWS_SECRET_ACCESS_KEY="%s"\r\nexport AWS_SESSION_TOKEN="%s"\r\nexport AWS_DEFAULT_REGION="%s"\r\n' "$KEY" "$SECRET" "$TOKEN" "us-west-2")"
esperar_valores "us-west-2"

echo "=== export con comillas simples ==="
correr "$(printf "export AWS_ACCESS_KEY_ID='%s'\nexport AWS_SECRET_ACCESS_KEY='%s'\nexport AWS_SESSION_TOKEN='%s'\nexport AWS_REGION='%s'\n" "$KEY" "$SECRET" "$TOKEN" "us-east-1")"
esperar_valores "us-east-1"

echo "=== set de cmd ==="
correr "$(printf 'set AWS_ACCESS_KEY_ID=%s\nset AWS_SECRET_ACCESS_KEY=%s\nset AWS_SESSION_TOKEN=%s\nset AWS_DEFAULT_REGION=%s\n' "$KEY" "$SECRET" "$TOKEN" "us-west-2")"
esperar_valores "us-west-2"

echo "=== PowerShell \$Env ==="
correr "$(printf '$Env:AWS_ACCESS_KEY_ID="%s"\n$Env:AWS_SECRET_ACCESS_KEY="%s"\n$Env:AWS_SESSION_TOKEN="%s"\n$Env:AWS_DEFAULT_REGION="%s"\n' "$KEY" "$SECRET" "$TOKEN" "us-east-1")"
esperar_valores "us-east-1"

echo "=== una línea separada por punto y coma ==="
correr "$(printf 'export AWS_ACCESS_KEY_ID=%s; export AWS_SECRET_ACCESS_KEY=%s; export AWS_SESSION_TOKEN=%s; export AWS_DEFAULT_REGION=%s' "$KEY" "$SECRET" "$TOKEN" "us-west-2")"
esperar_valores "us-west-2"

echo "=== una línea separada por espacios ==="
correr "$(printf 'export AWS_ACCESS_KEY_ID=%s export AWS_SECRET_ACCESS_KEY=%s export AWS_SESSION_TOKEN=%s export AWS_DEFAULT_REGION=%s' "$KEY" "$SECRET" "$TOKEN" "us-east-1")"
esperar_valores "us-east-1"

echo "=== [default] y otra sección; solo vale default ==="
correr "$(printf '[otro]\naws_access_key_id=%s\naws_secret_access_key=%s\naws_session_token=%s\n[default]\naws_access_key_id=%s\naws_secret_access_key=%s\naws_session_token=%s\nregion=%s\n' "$OTRA_KEY" "$OTRA_SECRET" "$OTRA_TOKEN" "$KEY" "$SECRET" "$TOKEN" "us-west-2")"
esperar_valores "us-west-2"

echo "=== región ausente: usa AWS_REGION del entorno ==="
correr "$(printf 'aws_access_key_id=%s aws_secret_access_key=%s aws_session_token=%s' "$KEY" "$SECRET" "$TOKEN")" AWS_REGION=us-west-2
esperar_valores "us-west-2"

echo "=== región ausente y sin AWS_REGION: us-east-1 ==="
correr "$(printf 'aws_access_key_id=%s aws_secret_access_key=%s aws_session_token=%s' "$KEY" "$SECRET" "$TOKEN")"
esperar_valores "us-east-1"

echo "=== bloque vacío no escribe nada ==="
correr ""
[ "$RC" -eq 0 ] || fail "bloque vacío salió $RC"
[ ! -s "$OUT" ] || fail "bloque vacío escribió en stdout"
[ ! -s "$ERR" ] || fail "bloque vacío escribió en stderr"
[ ! -s "$ENVFILE" ] || fail "bloque vacío escribió GITHUB_ENV"

echo "=== bloque incompleto: error sin valores y sin GITHUB_ENV ==="
correr "$(printf 'export AWS_ACCESS_KEY_ID="%s"\nexport AWS_SECRET_ACCESS_KEY="%s"\n' "$KEY" "$SECRET")"
[ "$RC" -ne 0 ] || fail "el bloque incompleto debía fallar"
salida_solo_mascaras
grep -q 'aws_session_token' "$ERR" || fail "el error no dice que falta el token"
[ ! -s "$ENVFILE" ] || fail "un bloque inválido no debe exportar"

echo "=== región ajena, sin imprimir el valor ==="
correr "$(printf 'aws_access_key_id=%s aws_secret_access_key=%s aws_session_token=%s region=%s' "$KEY" "$SECRET" "$TOKEN" "eu-central-1")"
[ "$RC" -ne 0 ] || fail "eu-central-1 debía rechazarse"
salida_solo_mascaras
if grep -q 'eu-central-1' "$ERR"; then
  fail "el error imprimió la región recibida"
fi
grep -q 'us-east-1' "$ERR" || fail "el error no dice qué regiones valen"
[ ! -s "$ENVFILE" ] || fail "región inválida no debe exportar"

echo "=== valor con espacio, sin imprimirlo ==="
correr "$(printf 'aws_access_key_id=%s EXTRA aws_secret_access_key=%s aws_session_token=%s' "$KEY" "$SECRET" "$TOKEN")"
[ "$RC" -ne 0 ] || fail "el valor con espacio debía fallar"
salida_solo_mascaras
if grep -F -q -- "$KEY" "$ERR"; then
  fail "el error imprimió el access key"
fi

echo "=== el script no activa set -x ni hace echo del bloque ==="
if grep -n 'set -x' "$SCRIPT" | grep -v 'set +x' | grep -q .; then
  fail "el script activa set -x"
fi
if grep -n -E 'echo .*(BLOQUE|PARSE_KEY|PARSE_SECRET|PARSE_TOKEN|VALOR)' "$SCRIPT" | grep -q .; then
  fail "el script hace echo de un valor"
fi

echo "=== el plan rechaza los valores extraídos y una línea del bloque ==="
printf 'plan sin secretos\n' >"$TMP/plan-limpio.txt"
env AWS_ACCESS_KEY_ID="$KEY" AWS_SECRET_ACCESS_KEY="$SECRET" AWS_SESSION_TOKEN="$TOKEN" \
  bash "$ASSERT" "$TMP/plan-limpio.txt" >/dev/null

printf 'colado %s\n' "$TOKEN" >"$TMP/plan-token.txt"
set +e
env AWS_ACCESS_KEY_ID="$KEY" AWS_SECRET_ACCESS_KEY="$SECRET" AWS_SESSION_TOKEN="$TOKEN" \
  bash "$ASSERT" "$TMP/plan-token.txt" >"$TMP/assert.out" 2>"$TMP/assert.err"
rc=$?
set -e
[ "$rc" -eq 1 ] || fail "el plan con el token debía fallar"
if grep -F -q -- "$TOKEN" "$TMP/assert.err" || grep -F -q -- "$KEY" "$TMP/assert.err" || grep -F -q -- "$SECRET" "$TMP/assert.err"; then
  fail "assert imprimió un valor"
fi
grep -q 'AWS_SESSION_TOKEN' "$TMP/assert.err" || fail "assert no nombró el token"

linea_bloque="$(printf 'export AWS_ACCESS_KEY_ID="%s"' "$KEY")"
printf 'texto\n%s\n' "$linea_bloque" >"$TMP/plan-bloque.txt"
set +e
env BLOQUE_CREDENCIALES_AWS="$linea_bloque" \
  bash "$ASSERT" "$TMP/plan-bloque.txt" >"$TMP/assert2.out" 2>"$TMP/assert2.err"
rc=$?
set -e
[ "$rc" -eq 1 ] || fail "el plan con la línea del bloque debía fallar"
if grep -F -q -- "$KEY" "$TMP/assert2.err"; then
  fail "assert imprimió la línea del bloque"
fi
grep -q 'bloque_credenciales_aws' "$TMP/assert2.err" || fail "assert no nombró el bloque"

echo "=== los workflows usan el input y el parser ==="
for wf in aprovisionar-cuenta.yml pausar-cuenta.yml; do
  grep -q 'bloque_credenciales_aws:' "$ROOT/.github/workflows/$wf" || fail "$wf sin el input"
  grep -q 'parsear-bloque-credenciales-aws.sh' "$ROOT/.github/workflows/$wf" || fail "$wf no llama al parser"
  grep -q 'CREDENCIALES_DESDE_BLOQUE' "$ROOT/.github/workflows/$wf" || fail "$wf no prioriza el bloque"
  grep -q 'aws sts get-caller-identity' "$ROOT/scripts/aprovisionar-cuenta.sh" "$ROOT/scripts/bootstrap-estado-terraform.sh" "$ROOT/scripts/pausar-cuenta.sh" || fail "falta get-caller-identity"
done

echo "Pruebas del bloque de credenciales AWS pasaron."
