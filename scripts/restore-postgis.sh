#!/usr/bin/env bash
# Restaura un volcado -Fc en una base que debe existir y estar vacía.
# Uso:
#   DATABASE_URL=postgres://campus:clave@127.0.0.1:5432/campus_restore bash scripts/restore-postgis.sh backup.dump
#   TARGET_DATABASE_URL=... bash scripts/restore-postgis.sh backup.dump
set -euo pipefail

dump="${1:?falta el archivo .dump}"
[ -f "$dump" ] || { echo "no existe: $dump" >&2; exit 1; }

url="${TARGET_DATABASE_URL:-${DATABASE_URL:-}}"
if [ -z "$url" ]; then
  echo "Defina TARGET_DATABASE_URL o DATABASE_URL de la base destino (vacía)." >&2
  exit 1
fi

# pg_restore devuelve 1 si hay avisos de objetos ya existentes. En una base vacía no debe haberlos.
pg_restore --no-owner --no-acl --exit-on-error -d "$url" "$dump"
echo "restaurado en la base destino"
