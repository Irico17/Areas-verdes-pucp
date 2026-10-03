#!/usr/bin/env bash
# Instala el Session Manager Plugin 1.2.707.0 (deb de Ubuntu) tras comprobar el SHA-256.
set -euo pipefail

VERSION="1.2.707.0"
SHA256="91cc9a1a8df6f730622d5e11a8610be6e06d5858408e8029b56d0d63aedd0a68"
URL="https://s3.amazonaws.com/session-manager-downloads/plugin/${VERSION}/ubuntu_64bit/session-manager-plugin.deb"

if command -v session-manager-plugin >/dev/null 2>&1; then
  actual="$(session-manager-plugin --version 2>/dev/null | awk '{print $1}')"
  if [ "$actual" = "$VERSION" ]; then
    echo "session-manager-plugin ${VERSION} ya está instalado."
    exit 0
  fi
fi

archivo="$(mktemp --suffix=.deb)"
trap 'rm -f "$archivo"' EXIT
curl -fsSL -o "$archivo" "$URL"
echo "${SHA256}  ${archivo}" | sha256sum -c -
if [ "$(id -u)" -eq 0 ]; then
  dpkg -i "$archivo"
else
  sudo dpkg -i "$archivo"
fi
session-manager-plugin --version
