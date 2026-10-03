#!/usr/bin/env bash
# Pruebas locales de credenciales, confirmaciones, bootstrap simulado e inventario.
# No llama a AWS ni a GitHub reales. Los valores se arman en tiempo de ejecución.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMP="$(mktemp -d)"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT

# shellcheck source=scripts/lib/credenciales-lab.sh
source "$ROOT/scripts/lib/credenciales-lab.sh"

fail() {
  echo "FALLO: $*" >&2
  exit 1
}

KEY="ASIA""FAKEEXAMPLEKEY01"
SECRET="wJalrXUtn""FEMI7MDENGFAKESECRETKEY"
TOKEN="FwoGZXIvYXdz""EFAKESESSIONTOKEN000000"
export KEY SECRET TOKEN

echo "=== parser del bloque AWS CLI ==="
bloque="$(printf 'export AWS_ACCESS_KEY_ID="%s"\r\nexport AWS_SECRET_ACCESS_KEY='\''%s'\''\nexport AWS_SESSION_TOKEN=%s\nexport AWS_DEFAULT_REGION=us-west-2\n' "$KEY" "$SECRET" "$TOKEN")"
parsear_bloque_aws <<<"$bloque"
validar_forma_credenciales
[ "$PARSE_KEY" = "$KEY" ] || fail "key"
[ "$PARSE_SECRET" = "$SECRET" ] || fail "secret"
[ "$PARSE_TOKEN" = "$TOKEN" ] || fail "token"
[ "$PARSE_REGION" = "us-west-2" ] || fail "region"
echo "parser OK"

echo "=== parser rechaza región de fuera del lab ==="
PARSE_REGION="eu-central-1"
if validar_forma_credenciales; then
  fail "eu-central-1 debió rechazarse"
fi
echo "región ajena OK"

echo "=== parser rechaza bloque incompleto ==="
parsear_bloque_aws <<EOF
export AWS_ACCESS_KEY_ID=${KEY}
export AWS_SECRET_ACCESS_KEY=${SECRET}
EOF
if credenciales_completas; then
  fail "faltaba el token y se dio por bueno"
fi
echo "bloque incompleto OK"

echo "=== script de credenciales con gh simulado ==="
BIN="$TMP/bin"
mkdir -p "$BIN" "$TMP/secretos"
export GH_LOG="$TMP/gh.log"
export GH_DIR="$TMP/secretos"
cat >"$BIN/gh" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$GH_LOG"
nombre=""
previo=""
for parte in "$@"; do
  if [ "$previo" = "set" ]; then
    nombre="$parte"
  fi
  previo="$parte"
done
if [ -n "$nombre" ] && [ "$nombre" != "AWS_REGION" ]; then
  cat > "$GH_DIR/$nombre"
fi
exit 0
EOF
chmod +x "$BIN/gh"
salida="$TMP/cred.out"
PATH="$BIN:/usr/bin:/bin" bash "$ROOT/scripts/actualizar-credenciales-lab.sh" >"$salida" <<<"$(printf 'export AWS_ACCESS_KEY_ID=%s\nexport AWS_SECRET_ACCESS_KEY=%s\nexport AWS_SESSION_TOKEN=%s\nexport AWS_DEFAULT_REGION=us-east-1\n' "$KEY" "$SECRET" "$TOKEN")"
grep -q 'environment aws-lab' "$salida" || fail "no mencionó aws-lab"
if grep -q "$KEY" "$salida" || grep -q "$SECRET" "$salida" || grep -q "$TOKEN" "$salida"; then
  fail "el script imprimió un valor"
fi
grep -q -- '--env aws-lab' "$GH_LOG" || fail "gh no recibió --env aws-lab"
[ "$(cat "$GH_DIR/AWS_ACCESS_KEY_ID")" = "$KEY" ] || fail "el secreto no llegó a gh"
[ "$(cat "$GH_DIR/AWS_SECRET_ACCESS_KEY")" = "$SECRET" ] || fail "secret access key"
[ "$(cat "$GH_DIR/AWS_SESSION_TOKEN")" = "$TOKEN" ] || fail "session token"
grep -q 'AWS_REGION' "$GH_LOG" || fail "no fijó la variable de región"
echo "credenciales OK"

echo "=== confirmaciones de aprovisionar y pausar ==="
APROVISIONAR_SOLO_VALIDAR=1 ACCION=plan AMBIENTES=nonprod bash "$ROOT/scripts/aprovisionar-cuenta.sh" >/dev/null
set +e
APROVISIONAR_SOLO_VALIDAR=1 ACCION=apply AMBIENTES=todos CONFIRMAR=aplicar bash "$ROOT/scripts/aprovisionar-cuenta.sh" >/dev/null 2>&1
rc=$?
set -e
[ "$rc" -eq 2 ] || fail "apply sin APLICAR debía salir 2, salió $rc"
set +e
APROVISIONAR_SOLO_VALIDAR=1 ACCION=destroy AMBIENTES=prod CONFIRMAR=APLICAR bash "$ROOT/scripts/aprovisionar-cuenta.sh" >/dev/null 2>&1
rc=$?
set -e
[ "$rc" -eq 2 ] || fail "destroy sin DESTRUIR debía salir 2"
APROVISIONAR_SOLO_VALIDAR=1 ACCION=apply AMBIENTES=todos CONFIRMAR=APLICAR DISCO_DATOS_GB=40 TIPO_NONPROD=t3.medium bash "$ROOT/scripts/aprovisionar-cuenta.sh" >/dev/null
set +e
APROVISIONAR_SOLO_VALIDAR=1 ACCION=apply AMBIENTES=nonprod CONFIRMAR=APLICAR DISCO_DATOS_GB=500 bash "$ROOT/scripts/aprovisionar-cuenta.sh" >/dev/null 2>&1
rc=$?
set -e
[ "$rc" -eq 2 ] || fail "disco 500 debía rechazarse"
PAUSAR_SOLO_VALIDAR=1 ACCION=stop AMBIENTES=todos CONFIRMAR=DETENER bash "$ROOT/scripts/pausar-cuenta.sh" >/dev/null
set +e
PAUSAR_SOLO_VALIDAR=1 ACCION=start AMBIENTES=prod CONFIRMAR=DETENER bash "$ROOT/scripts/pausar-cuenta.sh" >/dev/null 2>&1
rc=$?
set -e
[ "$rc" -eq 2 ] || fail "start con DETENER debía salir 2"
echo "confirmaciones OK"

echo "=== bootstrap con aws simulado ==="
export AWS_LOG="$TMP/aws.log"
: >"$AWS_LOG"
cat >"$BIN/aws" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$AWS_LOG"
if [[ "$*" == *get-caller-identity* ]]; then
  printf '%s\n' '{"Account":"999999999999","Arn":"arn:aws:sts::999999999999:assumed-role/voclabs/estudiante"}'
  exit 0
fi
if [[ "$*" == *head-bucket* ]]; then
  echo "An error occurred (404) when calling the HeadBucket operation: Not Found" >&2
  exit 1
fi
exit 0
EOF
chmod +x "$BIN/aws"
export AWS_ACCESS_KEY_ID="$KEY" AWS_SECRET_ACCESS_KEY="$SECRET" AWS_SESSION_TOKEN="$TOKEN"
boot="$TMP/boot.out"
PATH="$BIN:/usr/bin:/bin" AWS_REGION=us-east-1 bash "$ROOT/scripts/bootstrap-estado-terraform.sh" "$TMP/backend.hcl" >"$boot"
grep -q 'STATE_BUCKET=campus-verde-tfstate-999999999999' "$boot" || fail "cubo de estado"
grep -q 'SSM_BUCKET=campus-verde-ssm-999999999999' "$boot" || fail "cubo ssm"
grep -q 'bucket = "campus-verde-tfstate-999999999999"' "$TMP/backend.hcl" || fail "backend.hcl"
if grep -q 'LocationConstraint' "$AWS_LOG"; then
  fail "us-east-1 no debe enviar LocationConstraint"
fi
grep -q 'Status=Enabled' "$AWS_LOG" || fail "el estado debe versionarse"
grep -q 'Status=Suspended' "$AWS_LOG" || fail "el cubo SSM no debe versionarse"
if grep -q "$SECRET" "$boot" || grep -q "$TOKEN" "$boot"; then
  fail "bootstrap imprimió un secreto"
fi
echo "bootstrap us-east-1 OK"

: >"$AWS_LOG"
PATH="$BIN:/usr/bin:/bin" AWS_REGION=us-west-2 bash "$ROOT/scripts/bootstrap-estado-terraform.sh" >/dev/null
grep -q 'LocationConstraint=us-west-2' "$AWS_LOG" || fail "us-west-2 debe fijar LocationConstraint"
echo "bootstrap us-west-2 OK"

echo "=== credenciales caducadas ==="
cat >"$BIN/aws" <<'EOF'
#!/usr/bin/env bash
if [[ "$*" == *get-caller-identity* ]]; then
  echo "An error occurred (ExpiredToken) when calling the GetCallerIdentity operation: The security token included in the request is expired" >&2
  exit 254
fi
exit 0
EOF
chmod +x "$BIN/aws"
set +e
PATH="$BIN:/usr/bin:/bin" AWS_REGION=us-east-1 bash "$ROOT/scripts/bootstrap-estado-terraform.sh" >"$TMP/caduca.out" 2>"$TMP/caduca.err"
rc=$?
set -e
[ "$rc" -eq 1 ] || fail "caducadas debía salir 1, salió $rc"
grep -q 'caducaron' "$TMP/caduca.err" || fail "mensaje de caducidad"
if grep -q "$SECRET" "$TMP/caduca.err" || grep -q "$TOKEN" "$TMP/caduca.err"; then
  fail "el error imprimió un secreto"
fi
echo "caducidad OK"

echo "=== inventario y plan sin secretos ==="
cat >"$TMP/nonprod.json" <<'EOF'
{"instance_id":{"value":"i-0000000000000000a"},"public_ip":{"value":"203.0.113.10"}}
EOF
cat >"$TMP/prod.json" <<'EOF'
{"instance_id":{"value":"i-0000000000000000b"},"public_ip":{"value":"203.0.113.20"}}
EOF
bash "$ROOT/scripts/generar-inventario-ansible.sh" \
  --region us-east-1 --ssm-bucket campus-verde-ssm-999999999999 \
  --nonprod "$TMP/nonprod.json" --prod "$TMP/prod.json" --out "$TMP/hosts.yml"
grep -q 'ansible_connection: amazon.aws.aws_ssm' "$TMP/hosts.yml" || fail "conexión SSM"
grep -q 'i-0000000000000000a' "$TMP/hosts.yml" || fail "instancia nonprod"
grep -q 'campus_perfil: produccion' "$TMP/hosts.yml" || fail "perfil prod"
printf 'solo texto\n' >"$TMP/plan.txt"
AWS_SECRET_ACCESS_KEY="$SECRET" bash "$ROOT/scripts/assert-plan-sin-secretos.sh" "$TMP/plan.txt"
printf 'colado %s\n' "$SECRET" >"$TMP/plan-malo.txt"
set +e
AWS_SECRET_ACCESS_KEY="$SECRET" bash "$ROOT/scripts/assert-plan-sin-secretos.sh" "$TMP/plan-malo.txt" >"$TMP/assert.out" 2>"$TMP/assert.err"
rc=$?
set -e
[ "$rc" -eq 1 ] || fail "el plan con secreto debía fallar"
if grep -q "$SECRET" "$TMP/assert.err"; then
  fail "assert imprimió el secreto"
fi
echo "inventario y plan OK"

echo "=== recuperar estado, solo imprime ==="
cat >"$TMP/instances.json" <<'EOF'
{"Reservations":[{"Instances":[{"InstanceId":"i-0000000000000000a"}]}]}
EOF
cat >"$TMP/volumes.json" <<'EOF'
{"Volumes":[{"VolumeId":"vol-0000000000000000a","Attachments":[{"Device":"/dev/sdf","InstanceId":"i-0000000000000000a"}]}]}
EOF
cat >"$TMP/addresses.json" <<'EOF'
{"Addresses":[{"AllocationId":"eipalloc-0000000000000000a","AssociationId":"eipassoc-0000000000000000a"}]}
EOF
cat >"$TMP/groups.json" <<'EOF'
{"SecurityGroups":[{"GroupId":"sg-0000000000000000a"}]}
EOF
bash "$ROOT/scripts/recuperar-estado-terraform.sh" --stack nonprod \
  --instancias "$TMP/instances.json" --volumenes "$TMP/volumes.json" \
  --direcciones "$TMP/addresses.json" --grupos "$TMP/groups.json" >"$TMP/import.txt"
grep -q 'terraform -chdir=infra/terraform/nonprod import aws_instance.nonprod i-0000000000000000a' "$TMP/import.txt" || fail "import instancia"
grep -q 'aws_volume_attachment.data /dev/sdf:vol-0000000000000000a:i-0000000000000000a' "$TMP/import.txt" || fail "import volumen"
echo "recuperar OK"

echo "=== pausar elige instancias por etiqueta, aws simulado ==="
export AWS_LOG="$TMP/aws-pausa.log"
: >"$AWS_LOG"
cat >"$BIN/aws" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$AWS_LOG"
if [[ "$*" == *get-caller-identity* ]]; then
  printf '%s\n' '999999999999'
  exit 0
fi
if [[ "$*" == *describe-instances* ]]; then
  printf '%s\n' '{"Reservations":[{"Instances":[{"InstanceId":"i-0000000000000000a"}]}]}'
  exit 0
fi
exit 0
EOF
chmod +x "$BIN/aws"
PATH="$BIN:/usr/bin:/bin" AWS_REGION=us-east-1 ACCION=stop AMBIENTES=nonprod CONFIRMAR=DETENER \
  bash "$ROOT/scripts/pausar-cuenta.sh" >"$TMP/pausa.out"
grep -q 'tag:CampusGestion,Values=cuenta-lab' "$AWS_LOG" || fail "filtro CampusGestion"
grep -q 'tag:Stack,Values=nonprod' "$AWS_LOG" || fail "filtro stack"
grep -q 'stop-instances' "$AWS_LOG" || fail "no detuvo"
grep -q 'i-0000000000000000a' "$TMP/pausa.out" || fail "no anunció el id"
echo "pausar OK"

echo "=== el código ya no fija la cuenta anterior ==="
cuenta_vieja="8909""91908027"
if grep -R -n -F "$cuenta_vieja" "$ROOT/infra" "$ROOT/.github" "$ROOT/scripts" "$ROOT/docs" >/dev/null 2>&1; then
  fail "sigue el id de cuenta fijo"
fi
echo "sin id fijo OK"

echo "Todas las pruebas de cuenta-lab pasaron."
