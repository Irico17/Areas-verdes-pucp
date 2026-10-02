#!/usr/bin/env bash
# Conteos de tablas public (el SQL del §4.4). No modifica datos.
# Uso local: deploy/conteos.sh <develop|qa|produccion> [--todas]
# Contra un contenedor ya elegido: CONTAINER=campus-develop-db POSTGRES_USER=campus POSTGRES_DB=... deploy/conteos.sh [--todas]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SQL="$ROOT/deploy/conteos.sql"
EXCLUIR="$ROOT/deploy/conteos.excluir"

todas=0
for arg in "$@"; do
  if [ "$arg" = "--todas" ]; then
    todas=1
  fi
done

filtrar_conteos() {
  if [ "$todas" = "1" ] || [ ! -f "$EXCLUIR" ]; then
    cat
  else
    python3 "$ROOT/deploy/comparar_conteos.py" --filtrar "$EXCLUIR"
  fi
}

if [ -n "${CONTAINER:-}" ]; then
  docker exec -i "$CONTAINER" psql -U "${POSTGRES_USER:?}" -d "${POSTGRES_DB:?}" -v ON_ERROR_STOP=1 -A -F "$(printf '\t')" -P footer=off -f - <"$SQL" | filtrar_conteos
  exit 0
fi

AMBIENTE="${1:?falta el ambiente: develop, qa o produccion}"
case "$AMBIENTE" in
  develop|qa|produccion) ;;
  *)
    echo "ambiente desconocido: $AMBIENTE" >&2
    exit 2
    ;;
esac

ENV_FILE="$ROOT/deploy/env/${AMBIENTE}.env"
if [ ! -f "$ENV_FILE" ]; then
  cp "$ROOT/deploy/env/${AMBIENTE}.env.example" "$ENV_FILE"
  chmod 600 "$ENV_FILE"
fi
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

docker exec -i "campus-${APP_ENV}-db" \
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -A -F "$(printf '\t')" -P footer=off -f - <"$SQL" | filtrar_conteos
