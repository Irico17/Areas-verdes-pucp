#!/usr/bin/env bash
# Conteos de catastro cargado en PostGIS.
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

docker compose exec -T db psql -U "$user" -d "$db" -c "
SELECT 'areas_verdes' AS tabla, count(*) AS n,
       count(geom) AS con_geom,
       count(*) FILTER (WHERE ST_SRID(geom) = 4326) AS srid_4326
FROM areas_verdes
UNION ALL
SELECT 'zonas', count(*), count(geom),
       count(*) FILTER (WHERE ST_SRID(geom) = 4326)
FROM zonas
UNION ALL
SELECT capa, count(*), count(geom),
       count(*) FILTER (WHERE ST_SRID(geom) = 4326)
FROM capas_auxiliares
GROUP BY capa
ORDER BY 1;
"
