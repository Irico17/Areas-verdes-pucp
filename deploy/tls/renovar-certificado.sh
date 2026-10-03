#!/usr/bin/env bash
# Renueva el certificado solo si faltan menos de 30 días.
# Pensado para el timer systemd. Sale 0 si aún no toca renovar.
# El fallo (token ausente, lego, openssl) sale distinto de 0 para journalctl.
# No imprime el token. No recrea contenedores.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=deploy/tls/comun.sh
source "$SCRIPT_DIR/comun.sh"

tls_certs_dir
fullchain="${CAMPUS_CERTS_DIR}/fullchain.pem"
if [ ! -f "$fullchain" ]; then
  tls_error "no hay ${fullchain}. Emita primero con emitir-certificado.sh. No se llama a Let's Encrypt."
fi
if ! command -v openssl >/dev/null 2>&1; then
  tls_error "hace falta openssl para mirar la caducidad. No se renueva a ciegas."
fi

fin="$(openssl x509 -enddate -noout -in "$fullchain" 2>/dev/null || true)"
echo "Certificado actual: ${fin:-sin fecha}."

# 30 días. checkend sale 0 si el certificado sigue vigente pasado ese plazo.
if openssl x509 -checkend 2592000 -noout -in "$fullchain" >/dev/null 2>&1; then
  echo "Faltan más de 30 días. No se renueva."
  exit 0
fi

echo "Faltan menos de 30 días (o ya venció). Se renueva."
tls_lego_run renew --days 30
echo "Renovación terminada."
