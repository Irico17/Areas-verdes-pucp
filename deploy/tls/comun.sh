#!/usr/bin/env bash
# Funciones compartidas por los scripts de TLS. No se ejecuta sola.
if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  echo "Este archivo se incluye desde los otros scripts de deploy/tls; no se ejecuta solo." >&2
  exit 1
fi

# Imagen fijada por versión y digest del índice publicado en Docker Hub (v5.5.2).
# El proveedor duckdns lee DUCKDNS_TOKEN del entorno; no va en la línea de comandos.
LEGO_IMAGE_DEFAULT="goacme/lego:v5.5.2@sha256:1944e8c36055beec47c7de6f15202b41128be75eea0ffa257f0c14d93c5155fd"

tls_error() {
  echo "ERROR: $*" >&2
  exit 1
}

tls_cargar_token() {
  local env_file="${DUCKDNS_ENV_FILE:-/opt/campus/duckdns.env}"
  if [ -z "${DUCKDNS_TOKEN:-}" ]; then
    if [ ! -e "$env_file" ]; then
      tls_error "DUCKDNS_TOKEN no está definido y no existe ${env_file}. Cree ese archivo con modo 600 fuera del repositorio. El valor no se imprime."
    fi
    if [ ! -f "$env_file" ]; then
      tls_error "${env_file} existe pero no es un archivo regular."
    fi
    local mode
    mode="$(stat -c '%a' "$env_file")"
    case "$mode" in
      600|400) ;;
      *) tls_error "${env_file} debe tener modo 600 (tiene ${mode}). No se lee el token." ;;
    esac
    set -a
    # shellcheck disable=SC1090
    source "$env_file"
    set +a
  fi
  if [ -z "${DUCKDNS_TOKEN:-}" ]; then
    tls_error "DUCKDNS_TOKEN está vacío. Defínalo en el entorno o en ${env_file}. El valor no se imprime."
  fi
  if ! [[ "$DUCKDNS_TOKEN" =~ ^[A-Za-z0-9_-]{8,128}$ ]]; then
    tls_error "DUCKDNS_TOKEN tiene un formato no admitido. Use solo letras, dígitos, guion y guion bajo. El valor no se imprime."
  fi
}

tls_dominio() {
  DUCKDNS_DOMAIN="${DUCKDNS_DOMAIN:-verde-pucp}"
  if ! [[ "$DUCKDNS_DOMAIN" =~ ^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$ ]]; then
    tls_error "DUCKDNS_DOMAIN debe ser el subdominio (por defecto verde-pucp), sin puntos ni el sufijo duckdns.org."
  fi
  DUCKDNS_FQDN="${DUCKDNS_DOMAIN}.duckdns.org"
}

tls_certs_dir() {
  CAMPUS_CERTS_DIR="${CAMPUS_CERTS_DIR:-/opt/campus/certs}"
}

tls_lego_dir() {
  LEGO_DATA_DIR="${LEGO_DATA_DIR:-/opt/campus/letsencrypt}"
}

tls_instalar_pem() {
  local src_chain="$1"
  local src_key="$2"
  tls_certs_dir
  if [ ! -s "$src_chain" ] || [ ! -s "$src_key" ]; then
    tls_error "el cliente ACME no dejó cadena o clave para instalar."
  fi
  mkdir -p "$CAMPUS_CERTS_DIR"
  local tmp_chain tmp_key
  tmp_chain="$(mktemp "${CAMPUS_CERTS_DIR}/.fullchain.XXXXXX")"
  tmp_key="$(mktemp "${CAMPUS_CERTS_DIR}/.privkey.XXXXXX")"
  cat "$src_chain" > "$tmp_chain"
  cat "$src_key" > "$tmp_key"
  chmod 644 "$tmp_chain"
  chmod 600 "$tmp_key"
  local owner="${CERT_OWNER:-$(id -u):$(id -g)}"
  if ! chown "$owner" "$tmp_chain" "$tmp_key" 2>/dev/null; then
    echo "AVISO: no se pudo hacer chown de los PEM a ${owner}. Siguen del usuario que ejecuta el script." >&2
  fi
  mv -f "$tmp_chain" "${CAMPUS_CERTS_DIR}/fullchain.pem"
  mv -f "$tmp_key" "${CAMPUS_CERTS_DIR}/privkey.pem"
  chmod 644 "${CAMPUS_CERTS_DIR}/fullchain.pem"
  chmod 600 "${CAMPUS_CERTS_DIR}/privkey.pem"
  echo "PEM instalados en ${CAMPUS_CERTS_DIR} (fullchain.pem 644, privkey.pem 600)."
}

tls_recargar_nginx_si_ya_habla_tls() {
  local ambiente="${APP_ENV:-produccion}"
  local container="${WEB_CONTAINER:-campus-${ambiente}-web}"
  if ! command -v docker >/dev/null 2>&1; then
    echo "AVISO: no hay docker en el PATH; no se recarga nginx." >&2
    return 0
  fi
  if ! docker ps --format '{{.Names}}' 2>/dev/null | grep -qx "$container"; then
    echo "AVISO: ${container} no está en marcha. La primera vez hay que crearlo con los PEM ya montados. Este script no lo crea."
    return 0
  fi
  if docker exec "$container" grep -q 'listen 443' /etc/nginx/conf.d/default.conf 2>/dev/null; then
    docker exec "$container" nginx -s reload
    echo "nginx recargado en ${container} sin reiniciar el contenedor."
    return 0
  fi
  echo "AVISO: ${container} arrancó sin la config TLS. Hay que recrearlo una vez para que el entrypoint copie nginx.tls.conf. Este script no lo recrea."
}

tls_aviso_primera_vez() {
  cat <<'EOF'
Este script no recrea contenedores ni reinicia la API.
Orden del operador, después de la primera instalación del PEM:
  1. Confirmar fullchain.pem (644) y privkey.pem (600) en el directorio de certificados.
  2. En el host.env de producción: CAMPUS_COOKIE_SECURE=true, PUBLIC_URL y CAMPUS_CORS_ORIGINS en https://verde-pucp.duckdns.org.
  3. Recrear el contenedor web (el entrypoint elige la config al arrancar) y reiniciar la API.
  4. Verificar con curl. Solo entonces, si una persona lo decide, mover la IP del dominio.
Hasta ese cambio de DNS el nombre sigue en la IP doméstica. No se mueve desde aquí.
EOF
}

tls_lego_run() {
  local lego_cmd="$1"
  shift
  tls_cargar_token
  tls_dominio
  tls_lego_dir
  local image="${LEGO_IMAGE:-$LEGO_IMAGE_DEFAULT}"
  local email="${ACME_EMAIL:-}"
  if [ -z "$email" ]; then
    tls_error "defina ACME_EMAIL (contacto de Let's Encrypt; no es un secreto). No se llama a la API."
  fi
  if ! [[ "$email" =~ ^[^@[:space:]]+@[^@[:space:]]+\.[^@[:space:]]+$ ]]; then
    tls_error "ACME_EMAIL no parece un correo. No se llama a Let's Encrypt."
  fi
  if ! command -v docker >/dev/null 2>&1; then
    tls_error "hace falta Docker en el host para la imagen de lego. No se llama a Let's Encrypt."
  fi
  mkdir -p "$LEGO_DATA_DIR"
  chmod 700 "$LEGO_DATA_DIR" || true

  local server_args=()
  # Por defecto staging. Solo ACME_STAGING=0 usa el directorio real.
  if [ "${ACME_STAGING:-1}" = "0" ]; then
    echo "Directorio ACME de producción (ACME_STAGING=0). El dominio es ${DUCKDNS_FQDN}."
  else
    server_args=(--server https://acme-staging-v02.api.letsencrypt.org/directory)
    echo "Directorio ACME de staging (ACME_STAGING distinto de 0). El certificado no será de confianza pública. Dominio ${DUCKDNS_FQDN}."
  fi

  local envf
  envf="$(mktemp)"
  chmod 600 "$envf"
  printf 'DUCKDNS_TOKEN=%s\n' "$DUCKDNS_TOKEN" > "$envf"

  echo "Ejecutando lego (${lego_cmd}) con el token en un env-file de modo 600. No se imprime."
  set +e
  (
    trap 'rm -f "$envf"' EXIT
    docker run --rm \
      --user "$(id -u):$(id -g)" \
      --env-file "$envf" \
      -v "${LEGO_DATA_DIR}:/lego" \
      "$image" \
      --path /lego \
      --accept-tos \
      --email "$email" \
      --dns duckdns \
      --domains "$DUCKDNS_FQDN" \
      "${server_args[@]}" \
      "$lego_cmd" "$@"
  )
  local lego_rc=$?
  set -e
  rm -f "$envf"
  if [ "$lego_rc" -ne 0 ]; then
    tls_error "lego terminó con código ${lego_rc}. No se imprime el token ni la URL del proveedor."
  fi

  local crt="${LEGO_DATA_DIR}/certificates/${DUCKDNS_FQDN}.crt"
  local issuer="${LEGO_DATA_DIR}/certificates/${DUCKDNS_FQDN}.issuer.crt"
  local key="${LEGO_DATA_DIR}/certificates/${DUCKDNS_FQDN}.key"
  if [ ! -s "$crt" ] || [ ! -s "$key" ]; then
    tls_error "lego terminó sin ${DUCKDNS_FQDN}.crt o .key en ${LEGO_DATA_DIR}/certificates."
  fi
  local chain
  chain="$(mktemp)"
  chmod 600 "$chain"
  cat "$crt" > "$chain"
  if [ -s "$issuer" ]; then
    cat "$issuer" >> "$chain"
  else
    echo "AVISO: no apareció el certificado del emisor; fullchain.pem queda solo con el certificado hoja." >&2
  fi
  tls_instalar_pem "$chain" "$key"
  rm -f "$chain"
  tls_recargar_nginx_si_ya_habla_tls
  tls_aviso_primera_vez
}
