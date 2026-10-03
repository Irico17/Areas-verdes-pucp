#!/usr/bin/env bash
# Comprueba la sintaxis de nginx.tls.conf con un PEM autofirmado de prueba.
# El certificado se crea en un directorio temporal y no se versiona.
# Si no hay Docker, imprime el paso manual y sale 0.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CONF="$ROOT/apps/web/nginx.tls.conf"
HEADERS="$ROOT/apps/web/nginx-headers.conf"

if ! command -v docker >/dev/null 2>&1 || ! docker info >/dev/null 2>&1; then
  cat <<'EOF'
OMITIDO: Docker no está disponible. Paso manual, con un PEM autofirmado en un directorio temporal:

  d=$(mktemp -d)
  openssl req -x509 -newkey rsa:2048 -nodes -days 2 \
    -subj "/CN=verde-pucp.duckdns.org" \
    -addext "subjectAltName=DNS:verde-pucp.duckdns.org" \
    -keyout "$d/privkey.pem" -out "$d/fullchain.pem"
  docker run --rm \
    -v "$PWD/apps/web/nginx.tls.conf:/etc/nginx/conf.d/default.conf:ro" \
    -v "$PWD/apps/web/nginx-headers.conf:/etc/nginx/snippets/campus-headers.conf:ro" \
    -v "$d:/etc/nginx/certs:ro" \
    --add-host api:127.0.0.1 \
    nginx:1.27-alpine nginx -t
EOF
  exit 0
fi

tmp="$(mktemp -d)"
cleanup() { rm -rf "$tmp"; }
trap cleanup EXIT

openssl req -x509 -newkey rsa:2048 -nodes -days 2 \
  -subj "/CN=verde-pucp.duckdns.org" \
  -addext "subjectAltName=DNS:verde-pucp.duckdns.org" \
  -keyout "$tmp/privkey.pem" -out "$tmp/fullchain.pem" >/dev/null 2>&1

docker run --rm \
  -v "$CONF:/etc/nginx/conf.d/default.conf:ro" \
  -v "$HEADERS:/etc/nginx/snippets/campus-headers.conf:ro" \
  -v "$tmp:/etc/nginx/certs:ro" \
  --add-host api:127.0.0.1 \
  nginx:1.27-alpine nginx -t

echo "nginx -t aceptó apps/web/nginx.tls.conf"
