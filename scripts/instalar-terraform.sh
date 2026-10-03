#!/usr/bin/env bash
# Instala Terraform 1.11.4 en tmp/bin y comprueba el SHA-256 publicado por HashiCorp.
set -euo pipefail

VERSION="1.11.4"
SHA256="1ce994251c00281d6845f0f268637ba50c0005657eb3cf096b92f753b42ef4dc"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="${TERRAFORM_INSTALL_DIR:-$ROOT/tmp/bin}"
mkdir -p "$DEST"

if [ -x "$DEST/terraform" ] && "$DEST/terraform" version 2>/dev/null | grep -q "Terraform v${VERSION}\$"; then
  echo "Terraform ${VERSION} ya está en ${DEST}."
  exit 0
fi

archivo="$(mktemp)"
trap 'rm -f "$archivo"' EXIT
curl -fsSL -o "$archivo" "https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_linux_amd64.zip"
echo "${SHA256}  ${archivo}" | sha256sum -c -
unzip -o "$archivo" -d "$DEST" >/dev/null
chmod 755 "$DEST/terraform"
echo "Terraform ${VERSION} instalado en ${DEST}/terraform."
