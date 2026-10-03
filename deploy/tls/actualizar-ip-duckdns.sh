#!/usr/bin/env bash
# Actualiza la IP de verde-pucp.duckdns.org en DuckDNS.
# Exige --ip o --auto. --dry-run no llama a la red.
# No se ejecuta desde CI ni desde un runner de GitHub: apuntaría el dominio
# a la IP de esa máquina. El token no se imprime ni viaja en argv.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=deploy/tls/comun.sh
source "$SCRIPT_DIR/comun.sh"

usage() {
  cat <<'EOF'
Uso: deploy/tls/actualizar-ip-duckdns.sh (--ip <IPv4> | --auto) [--dry-run]

  --ip <IPv4>   registro A explícito (obligatorio si no usa --auto)
  --auto        DuckDNS usa la IP de origen de esta máquina
  --dry-run     comprueba el token y muestra dominio e IP, sin enviar nada

El token sale de DUCKDNS_TOKEN o de DUCKDNS_ENV_FILE (modo 600,
por defecto /opt/campus/duckdns.env). No lo pase como argumento.
EOF
}

if [ "${GITHUB_ACTIONS:-}" = "true" ] || [ "${CI:-}" = "true" ]; then
  tls_error "este script no se ejecuta desde CI ni desde un runner de GitHub: apuntaría el dominio a la IP del runner."
fi

ip=""
auto=0
dry=0
vio_ip=0

while [ $# -gt 0 ]; do
  case "$1" in
    --ip)
      shift
      ip="${1:-}"
      vio_ip=1
      if [ -z "$ip" ]; then
        tls_error "falta el valor de --ip."
      fi
      ;;
    --auto)
      auto=1
      ;;
    --dry-run)
      dry=1
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Opción desconocida: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
  shift
done

if [ "$vio_ip" -eq 1 ] && [ "$auto" -eq 1 ]; then
  tls_error "use solo uno de --ip o --auto."
fi
if [ "$vio_ip" -eq 0 ] && [ "$auto" -eq 0 ]; then
  usage >&2
  tls_error "hace falta --ip <IPv4> o --auto. No hay un valor por defecto y no se envía nada."
fi

modo_ip="auto"
if [ "$auto" -eq 0 ]; then
  if ! [[ "$ip" =~ ^([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})\.([0-9]{1,3})$ ]]; then
    tls_error "--ip debe ser una IPv4. No se envía nada."
  fi
  IFS=. read -r o1 o2 o3 o4 <<< "$ip"
  for octeto in "$o1" "$o2" "$o3" "$o4"; do
    if [ "${#octeto}" -gt 1 ] && [ "${octeto:0:1}" = "0" ]; then
      tls_error "--ip no admite ceros a la izquierda. No se envía nada."
    fi
    if [ "$((10#$octeto))" -gt 255 ]; then
      tls_error "--ip tiene un octeto mayor que 255. No se envía nada."
    fi
  done
  modo_ip="$ip"
fi

tls_cargar_token
tls_dominio

if [ "$dry" -eq 1 ]; then
  echo "DuckDNS dry-run: no se envía la petición."
  echo "dominio=${DUCKDNS_DOMAIN} fqdn=${DUCKDNS_FQDN} ip=${modo_ip} token=definido"
  exit 0
fi

endpoint="${DUCKDNS_UPDATE_URL:-https://www.duckdns.org/update}"
case "$endpoint" in
  http://*|https://*) ;;
  *) tls_error "DUCKDNS_UPDATE_URL debe empezar por http:// o https://." ;;
esac
if printf '%s' "$endpoint" | grep -q 'token='; then
  tls_error "DUCKDNS_UPDATE_URL no debe incluir el token."
fi

qs="domains=${DUCKDNS_DOMAIN}"
if [ "$modo_ip" != "auto" ]; then
  qs="${qs}&ip=${modo_ip}"
fi

cfg="$(mktemp)"
body="$(mktemp)"
chmod 600 "$cfg" "$body"
cleanup() {
  rm -f "$cfg" "$body"
}
trap cleanup EXIT

{
  printf 'silent\n'
  printf 'max-time = 20\n'
  printf 'output = "%s"\n' "$body"
  printf 'url = "%s?%s&token=%s"\n' "$endpoint" "$qs" "$DUCKDNS_TOKEN"
} > "$cfg"

set +e
curl --config "$cfg"
rc=$?
set -e
if [ "$rc" -ne 0 ]; then
  tls_error "no se pudo contactar DuckDNS (curl ${rc}). No se imprime la URL ni el token."
fi

resp="$(tr -d '\r\n[:space:]' < "$body" || true)"
if printf '%s' "$resp" | grep -qF "$DUCKDNS_TOKEN"; then
  tls_error "la respuesta incluye el token; se omite el texto."
fi
case "$resp" in
  OK)
    echo "DuckDNS respondió OK (dominio=${DUCKDNS_DOMAIN} ip=${modo_ip})."
    ;;
  KO)
    tls_error "DuckDNS respondió KO (dominio=${DUCKDNS_DOMAIN}). Revise el token sin volcarlo al log."
    ;;
  *)
    tls_error "respuesta inesperada de DuckDNS (no es OK ni KO). No se imprime el cuerpo."
    ;;
esac
