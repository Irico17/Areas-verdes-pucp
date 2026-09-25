#!/usr/bin/env bash
# Volcado custom de PostGIS (pg_dump -Fc). Incluye CREATE EXTENSION.
# Uso:
#   DATABASE_URL=postgres://campus:clave@127.0.0.1:5432/campus_verde bash scripts/backup-postgis.sh [salida.dump]
#   # o, con el compose local en marcha:
#   bash scripts/backup-postgis.sh
set -euo pipefail

out="${1:-backup-postgis-$(date -u +%Y%m%dT%H%M%SZ).dump}"
mkdir -p "$(dirname "$out")" 2>/dev/null || true

dump_args=(-Fc --no-owner --no-acl)

if [ -n "${DATABASE_URL:-}" ]; then
  pg_dump "${dump_args[@]}" -d "$DATABASE_URL" -f "$out"
elif command -v docker >/dev/null 2>&1 && docker compose ps --status running -q db >/dev/null 2>&1; then
  docker compose exec -T db \
    pg_dump -U "${POSTGRES_USER:-campus}" "${dump_args[@]}" "${POSTGRES_DB:-campus_verde}" >"$out"
else
  echo "Defina DATABASE_URL o levante el servicio db de docker compose." >&2
  exit 1
fi

echo "backup: $out"
