#!/usr/bin/env bash
# Falla (código 1) si /health no responde status=ok. Pensado para cron o systemd.
# HEALTH_URL por defecto es el nginx local.
set -euo pipefail

url="${HEALTH_URL:-http://127.0.0.1/health}"
if ! body="$(curl -fsS --max-time 8 "$url")"; then
  echo "ALERTA: /health no responde ($url)" >&2
  exit 1
fi
case "$body" in
  *'"status":"ok"'*)
    echo "ok $url"
    ;;
  *)
    echo "ALERTA: /health no está ok" >&2
    exit 1
    ;;
esac
