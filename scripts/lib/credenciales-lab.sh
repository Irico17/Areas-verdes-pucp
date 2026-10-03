#!/usr/bin/env bash
# Funciones para leer el bloque «AWS CLI» del Learner Lab.
# No llaman a gh, no imprimen valores y no leen ~/.aws/credentials.

# Quita export, espacios y comillas externas. Rechaza valores con espacio.
valor_asignacion() {
  local raw="$1" q
  raw="${raw#*=}"
  raw="${raw#"${raw%%[![:space:]]*}"}"
  raw="${raw%"${raw##*[![:space:]]}"}"
  if [ "${#raw}" -ge 2 ]; then
    q="${raw:0:1}"
    if [ "$q" = '"' ] || [ "$q" = "'" ]; then
      if [ "${raw: -1}" = "$q" ]; then
        raw="${raw:1:${#raw}-2}"
      fi
    fi
  fi
  case "$raw" in
    *[[:space:]]*|*'`'*|*'$(*)'*)
      return 1
      ;;
  esac
  printf '%s' "$raw"
}

# Lee el bloque por stdin. Deja PARSE_KEY, PARSE_SECRET, PARSE_TOKEN y PARSE_REGION.
parsear_bloque_aws() {
  local linea limpia clave valor
  PARSE_KEY=""
  PARSE_SECRET=""
  PARSE_TOKEN=""
  PARSE_REGION=""
  while IFS= read -r linea || [ -n "${linea:-}" ]; do
    linea="${linea%$'\r'}"
    case "$linea" in
      ""|"#"*)
        continue
        ;;
    esac
    limpia="${linea#export }"
    limpia="${limpia#"${limpia%%[![:space:]]*}"}"
    clave="${limpia%%=*}"
    clave="${clave%"${clave##*[![:space:]]}"}"
    clave="$(printf '%s' "$clave" | tr '[:upper:]' '[:lower:]')"
    valor="$(valor_asignacion "$limpia")" || return 1
    case "$clave" in
      aws_access_key_id)
        PARSE_KEY="$valor"
        ;;
      aws_secret_access_key)
        PARSE_SECRET="$valor"
        ;;
      aws_session_token)
        PARSE_TOKEN="$valor"
        ;;
      aws_default_region|aws_region|region)
        PARSE_REGION="$valor"
        ;;
    esac
  done
}

credenciales_completas() {
  [ -n "${PARSE_KEY:-}" ] && [ -n "${PARSE_SECRET:-}" ] && [ -n "${PARSE_TOKEN:-}" ]
}

validar_forma_credenciales() {
  local region="${1:-}"
  if ! credenciales_completas; then
    echo "El bloque no trae aws_access_key_id, aws_secret_access_key y aws_session_token." >&2
    return 1
  fi
  if [ "${#PARSE_KEY}" -lt 16 ] || [ "${#PARSE_KEY}" -gt 128 ]; then
    echo "aws_access_key_id tiene una longitud inesperada." >&2
    return 1
  fi
  if [ "${#PARSE_SECRET}" -lt 16 ] || [ "${#PARSE_SECRET}" -gt 256 ]; then
    echo "aws_secret_access_key tiene una longitud inesperada." >&2
    return 1
  fi
  if [ "${#PARSE_TOKEN}" -lt 16 ] || [ "${#PARSE_TOKEN}" -gt 4096 ]; then
    echo "aws_session_token tiene una longitud inesperada." >&2
    return 1
  fi
  case "$PARSE_KEY" in
    *[!A-Za-z0-9/+_=.-]*)
      echo "aws_access_key_id tiene caracteres no permitidos." >&2
      return 1
      ;;
  esac
  if [ -z "$region" ]; then
    region="${PARSE_REGION:-us-east-1}"
  fi
  case "$region" in
    us-east-1|us-west-2)
      PARSE_REGION="$region"
      ;;
    *)
      echo "La región debe ser us-east-1 o us-west-2 (Learner Lab)." >&2
      return 1
      ;;
  esac
}
