#!/usr/bin/env bash
# Espera a que el servicio db de docker compose acepte conexiones.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

user="${POSTGRES_USER:-campus}"
db="${POSTGRES_DB:-campus_verde}"

for _ in $(seq 1 60); do
  if docker compose exec -T db pg_isready -U "$user" -d "$db" >/dev/null 2>&1 \
    && docker compose exec -T db psql -U "$user" -d "$db" -tAc "SELECT 1" >/dev/null 2>&1; then
    echo "PostGIS listo (${user}@${db})"
    exit 0
  fi
  sleep 1
done

echo "Tiempo de espera agotado: Postgres no respondió" >&2
exit 1
