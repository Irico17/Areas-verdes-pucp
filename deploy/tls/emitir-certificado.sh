#!/usr/bin/env bash
# Emite el certificado de verde-pucp.duckdns.org con lego (DNS-01 de DuckDNS)
# y lo instala en /opt/campus/certs. No recrea contenedores.
# Por defecto usa el staging de Let's Encrypt. El real pide ACME_STAGING=0.
#
#   DUCKDNS_ENV_FILE=/opt/campus/duckdns.env ACME_EMAIL=operador@example.com \
#     bash deploy/tls/emitir-certificado.sh
#   ACME_STAGING=0 bash deploy/tls/emitir-certificado.sh
#
# El token se lee de DUCKDNS_TOKEN o de un archivo modo 600. No se imprime.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=deploy/tls/comun.sh
source "$SCRIPT_DIR/comun.sh"

echo "Emisión de certificado para DuckDNS. Sin token definido, este script se detiene y no llama a Let's Encrypt."
tls_lego_run run
echo "Emisión terminada."
