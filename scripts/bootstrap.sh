#!/usr/bin/env bash
# Levanta PostGIS, migra y carga el catastro.
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

test -f .env || cp .env.example .env
set -a
# shellcheck disable=SC1091
source .env
set +a

docker compose up -d
./scripts/wait-db.sh
( cd apps/api && go run ./cmd/migrate )
( cd apps/api && go run ./cmd/etl )
echo "Bootstrap listo. Arranca la API con: make api"
