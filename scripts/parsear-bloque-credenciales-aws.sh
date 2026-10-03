#!/usr/bin/env bash
# Extrae access key, secret, session token y región de un bloque AWS CLI
# del Learner Lab. El bloque entra solo por BLOQUE_CREDENCIALES_AWS.
#
# Acepta [default] con aws_access_key_id (espacios opcionales alrededor de =),
# export AWS_ACCESS_KEY_ID=..., set AWS_..., $Env:AWS_..., comillas simples o
# dobles, CRLF y una sola línea separada por ';' o espacios.
#
# No imprime valores. Lo primero que hace es ::add-mask:: de cada línea del
# bloque y de access key, secret y session token. Si el bloque trae credenciales,
# las escribe en GITHUB_ENV. La región us-east-1 / us-west-2 no se enmascara
# sola: no es un secreto y ocultarla taparía el log; la línea del bloque que
# la contiene sí se enmascara.
set +x
set -euo pipefail
umask 077

fallar() {
  printf '%s\n' "$1" >&2
  exit 1
}

enmascarar() {
  if [ -n "${1:-}" ]; then
    printf '%s\n' "::add-mask::$1"
  fi
}

# Quita espacios y comillas externas. Rechaza espacios y sustitución de shell.
valor_limpio() {
  VALOR="$1"
  VALOR="${VALOR#"${VALOR%%[![:space:]]*}"}"
  VALOR="${VALOR%"${VALOR##*[![:space:]]}"}"
  if [ "${#VALOR}" -ge 2 ]; then
    local q="${VALOR:0:1}"
    if { [ "$q" = '"' ] || [ "$q" = "'" ]; } && [ "${VALOR: -1}" = "$q" ]; then
      VALOR="${VALOR:1:${#VALOR}-2}"
    fi
  fi
  case "$VALOR" in
    ""|*[[:space:]]*|*'`'*|*'$(*)'*)
      return 1
      ;;
  esac
  return 0
}

forma_ok() {
  local nombre="$1" valor="$2" min="$3" max="$4"
  local n="${#valor}"
  if [ "$n" -lt "$min" ] || [ "$n" -gt "$max" ]; then
    fallar "${nombre} tiene una longitud inesperada."
  fi
  case "$valor" in
    *[!A-Za-z0-9/+_=.-]*)
      fallar "${nombre} tiene caracteres no permitidos."
      ;;
  esac
}

VIO_SECCION=0
EN_DEFAULT=1
PARSE_KEY=""
PARSE_SECRET=""
PARSE_TOKEN=""
PARSE_REGION=""

procesar_trozo() {
  local trozo="$1" lower prefijo clave resto_valor
  trozo="${trozo#"${trozo%%[![:space:]]*}"}"
  trozo="${trozo%"${trozo##*[![:space:]]}"}"
  [ -n "$trozo" ] || return 0
  case "$trozo" in
    \#*) return 0 ;;
  esac
  lower="$(printf '%s' "$trozo" | tr '[:upper:]' '[:lower:]')"
  if [[ "$lower" =~ ^\[([^]]+)\]$ ]]; then
    VIO_SECCION=1
    if [ "${BASH_REMATCH[1]}" = "default" ]; then
      EN_DEFAULT=1
    else
      EN_DEFAULT=0
    fi
    return 0
  fi
  [[ "$lower" == *"="* ]] || return 0
  if [ "$VIO_SECCION" -eq 1 ] && [ "$EN_DEFAULT" -eq 0 ]; then
    return 0
  fi
  if [[ "$lower" =~ ^export[[:space:]]+ ]]; then
    prefijo="${#BASH_REMATCH[0]}"
    trozo="${trozo:$prefijo}"
    lower="${lower:$prefijo}"
  elif [[ "$lower" =~ ^set[[:space:]]+ ]]; then
    prefijo="${#BASH_REMATCH[0]}"
    trozo="${trozo:$prefijo}"
    lower="${lower:$prefijo}"
  elif [[ "$lower" =~ ^\$env: ]]; then
    prefijo="${#BASH_REMATCH[0]}"
    trozo="${trozo:$prefijo}"
    lower="${lower:$prefijo}"
  fi
  trozo="${trozo#"${trozo%%[![:space:]]*}"}"
  lower="$(printf '%s' "$trozo" | tr '[:upper:]' '[:lower:]')"
  clave="${lower%%=*}"
  clave="${clave%"${clave##*[![:space:]]}"}"
  resto_valor="${trozo#*=}"
  if ! valor_limpio "$resto_valor"; then
    fallar "Un valor del bloque tiene espacios o caracteres de shell. No se usa."
  fi
  case "$clave" in
    aws_access_key_id) PARSE_KEY="$VALOR" ;;
    aws_secret_access_key) PARSE_SECRET="$VALOR" ;;
    aws_session_token) PARSE_TOKEN="$VALOR" ;;
    aws_default_region|aws_region|region) PARSE_REGION="$VALOR" ;;
  esac
}

procesar_linea() {
  local linea="$1" lower n i resto prev salto start end k
  local -a inicios=()
  [ -n "$linea" ] || return 0
  lower="$(printf '%s' "$linea" | tr '[:upper:]' '[:lower:]')"
  n="${#lower}"
  i=0
  while [ "$i" -lt "$n" ]; do
    resto="${lower:$i}"
    if [[ "$resto" =~ ^(export[[:space:]]+|set[[:space:]]+|\$env:)?(aws_access_key_id|aws_secret_access_key|aws_session_token|aws_default_region|aws_region|region)[[:space:]]*= ]]; then
      prev=""
      if [ "$i" -gt 0 ]; then
        prev="${lower:$((i - 1)):1}"
      fi
      if [ "$i" -eq 0 ] || [[ "$prev" == [[:space:]] ]]; then
        salto="${#BASH_REMATCH[0]}"
        [ "$salto" -ge 1 ] || salto=1
        inicios+=("$i")
        i=$((i + salto))
        continue
      fi
    fi
    i=$((i + 1))
  done
  if [ "${#inicios[@]}" -eq 0 ]; then
    procesar_trozo "$linea"
    return 0
  fi
  if [ "${inicios[0]}" -gt 0 ]; then
    procesar_trozo "${linea:0:${inicios[0]}}"
  fi
  k=0
  while [ "$k" -lt "${#inicios[@]}" ]; do
    start="${inicios[$k]}"
    if [ $((k + 1)) -lt "${#inicios[@]}" ]; then
      end="${inicios[$((k + 1))]}"
    else
      end="$n"
    fi
    procesar_trozo "${linea:$start:$((end - start))}"
    k=$((k + 1))
  done
}

iterar_lineas() {
  local rest="$1" linea
  while [ -n "$rest" ]; do
    linea="${rest%%$'\n'*}"
    if [ "$rest" = "$linea" ]; then
      rest=""
    else
      rest="${rest#*$'\n'}"
    fi
    if [ -n "$linea" ]; then
      if [ "${2:-}" = "enmascarar" ]; then
        enmascarar "$linea"
      else
        procesar_linea "$linea"
      fi
    fi
  done
}

escribir_env() {
  local nombre="$1" valor="$2" delim="AWSCLI_VALOR_EOF"
  case "$valor" in
    *"$delim"*)
      fallar "No se pudo exportar una credencial."
      ;;
  esac
  {
    printf '%s<<%s\n' "$nombre" "$delim"
    printf '%s\n' "$valor"
    printf '%s\n' "$delim"
  } >>"$GITHUB_ENV"
}

bloque="${BLOQUE_CREDENCIALES_AWS:-}"
bloque="${bloque#$'\ufeff'}"
bloque="${bloque//$'\r'/}"
compacto="$(printf '%s' "$bloque" | tr -d '[:space:];')"
if [ -z "$compacto" ]; then
  exit 0
fi
unset compacto

# Antes de validar, fallar o escribir: enmascarar cada línea del bloque.
iterar_lineas "$bloque" enmascarar

parseado="${bloque//;/$'\n'}"
iterar_lineas "$parseado" parsear

enmascarar "$PARSE_KEY"
enmascarar "$PARSE_SECRET"
enmascarar "$PARSE_TOKEN"

if [ -z "$PARSE_KEY" ] || [ -z "$PARSE_SECRET" ] || [ -z "$PARSE_TOKEN" ]; then
  fallar "El bloque no trae aws_access_key_id, aws_secret_access_key y aws_session_token."
fi

forma_ok "aws_access_key_id" "$PARSE_KEY" 16 128
forma_ok "aws_secret_access_key" "$PARSE_SECRET" 16 256
forma_ok "aws_session_token" "$PARSE_TOKEN" 16 4096

if [ -z "$PARSE_REGION" ]; then
  PARSE_REGION="${AWS_REGION:-us-east-1}"
fi
case "$PARSE_REGION" in
  us-east-1|us-west-2) ;;
  *)
    fallar "La región debe ser us-east-1 o us-west-2 (Learner Lab)."
    ;;
esac

if [ -z "${GITHUB_ENV:-}" ]; then
  fallar "Falta GITHUB_ENV; no se escriben ni se imprimen las credenciales."
fi

escribir_env AWS_ACCESS_KEY_ID "$PARSE_KEY"
escribir_env AWS_SECRET_ACCESS_KEY "$PARSE_SECRET"
escribir_env AWS_SESSION_TOKEN "$PARSE_TOKEN"
escribir_env AWS_REGION "$PARSE_REGION"
escribir_env AWS_DEFAULT_REGION "$PARSE_REGION"
escribir_env CREDENCIALES_DESDE_BLOQUE "1"

unset PARSE_KEY PARSE_SECRET PARSE_TOKEN PARSE_REGION VALOR bloque parseado
exit 0
