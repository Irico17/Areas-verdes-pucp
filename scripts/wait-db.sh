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
port="${POSTGRES_PORT:-5432}"
password="${POSTGRES_PASSWORD:-campus}"

# El chequeo dentro del contenedor puede pasar mientras el puerto publicado
# todavía reinicia. make migrate entra por ese puerto, así que también se prueba.
host_ok() {
  if ! command -v pg_isready >/dev/null 2>&1 || ! command -v psql >/dev/null 2>&1; then
    return 0
  fi
  pg_isready -h 127.0.0.1 -p "$port" -U "$user" -d "$db" >/dev/null 2>&1 \
    && PGPASSWORD="$password" psql -h 127.0.0.1 -p "$port" -U "$user" -d "$db" -tAc "SELECT 1" >/dev/null 2>&1
}

for _ in $(seq 1 60); do
  if docker compose exec -T db pg_isready -U "$user" -d "$db" >/dev/null 2>&1 \
    && docker compose exec -T db psql -U "$user" -d "$db" -tAc "SELECT 1" >/dev/null 2>&1 \
    && host_ok; then
    echo "PostGIS listo (${user}@${db})"
    exit 0
  fi
  sleep 1
done

echo "Tiempo de espera agotado: Postgres no respondió" >&2
exit 1
