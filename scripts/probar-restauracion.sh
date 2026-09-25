#!/usr/bin/env bash
# Prueba los scripts de backup y de restauración contra una base vacía.
# No usa AWS. Requiere Docker y el cliente de Postgres 16.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
name="campus-restore-$$"
pass="clave-restauracion-16"
port="${RESTORE_PORT:-54329}"
docker rm -f "$name" >/dev/null 2>&1 || true
docker run -d --name "$name" -p "127.0.0.1:${port}:5432" \
  -e POSTGRES_USER=campus \
  -e POSTGRES_PASSWORD="$pass" \
  -e POSTGRES_DB=campus_verde \
  postgis/postgis:16-3.4 >/dev/null

dump="$(mktemp --suffix=.dump)"
cleanup() {
  docker rm -f "$name" >/dev/null 2>&1 || true
  rm -f "$dump"
}
trap cleanup EXIT

for _ in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20; do
  if pg_isready -h 127.0.0.1 -p "$port" -U campus -d campus_verde >/dev/null 2>&1; then
    break
  fi
  sleep 2
done
pg_isready -h 127.0.0.1 -p "$port" -U campus -d campus_verde

export PGPASSWORD="$pass"
psql -h 127.0.0.1 -p "$port" -U campus -d campus_verde -v ON_ERROR_STOP=1 -c \
  "CREATE EXTENSION IF NOT EXISTS postgis; CREATE TABLE prueba_restore (id int primary key, geom geometry(Point,4326)); INSERT INTO prueba_restore VALUES (1, ST_SetSRID(ST_MakePoint(-77.08,-12.07),4326));"

base="postgres://campus:${pass}@127.0.0.1:${port}"
DATABASE_URL="${base}/campus_verde" bash "$ROOT/scripts/backup-postgis.sh" "$dump"
psql -h 127.0.0.1 -p "$port" -U campus -d postgres -v ON_ERROR_STOP=1 -c \
  "CREATE DATABASE campus_restore OWNER campus"
TARGET_DATABASE_URL="${base}/campus_restore" bash "$ROOT/scripts/restore-postgis.sh" "$dump"

n="$(psql -h 127.0.0.1 -p "$port" -U campus -d campus_restore -tA -c \
  "SELECT count(*) FROM prueba_restore WHERE ST_X(geom) < -77")"
if [ "$n" != "1" ]; then
  echo "la restauración no devolvió la fila ($n)" >&2
  exit 1
fi
echo "restauración ok: 1 fila en campus_restore"
