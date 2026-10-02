#!/bin/sh
# Compatible con el sh de busybox (Alpine). Sin bash.
set -eu

# La imagen no trae una clave. El valor de laboratorio (cuentas ficticias) vive
# solo en el compose local; ver .env.example (CAMPUS_DEV_PASSWORD=pando-local).
if [ -z "${CAMPUS_DEV_PASSWORD:-}" ]; then
  echo "ERROR: falta CAMPUS_DEV_PASSWORD. La imagen no incluye una clave." >&2
  echo "En desarrollo local el valor documentado, solo para cuentas ficticias, es pando-local (docker-compose.yml y .env.example)." >&2
  echo "En cualquier otro entorno inyecte CAMPUS_DEV_PASSWORD. No reutilice la clave de laboratorio." >&2
  exit 1
fi

MIGRATIONS_DIR="${MIGRATIONS_DIR:-/opt/campus/migrations}"
export MIGRATIONS_DIR
if [ ! -d "$MIGRATIONS_DIR" ]; then
  echo "ERROR: MIGRATIONS_DIR=$MIGRATIONS_DIR no existe. El esquema se copia desde db/migrations." >&2
  exit 1
fi

MIGRATE_BIN="${MIGRATE_BIN:-/usr/local/bin/migrate}"
ETL_BIN="${ETL_BIN:-/usr/local/bin/etl}"
API_BIN="${API_BIN:-/usr/local/bin/api}"
ETL_DONE_FILE="${ETL_DONE_FILE:-/data/.etl-done}"
DATA_DIR="${DATA_DIR:-/data}"

if ! mkdir -p "$DATA_DIR/evidencias" "$DATA_DIR/v1" 2>/dev/null; then
  echo "ERROR: $DATA_DIR no es escribible por el usuario de la API (UID 10001). En el host: chown -R 10001:10001 <directorio montado en /data>." >&2
  exit 1
fi

if ! "$MIGRATE_BIN"; then
  echo "FALLO DE MIGRACION: la API no arranca. El error de arriba nombra el archivo y la sentencia. Reiniciar el contenedor no lo corrige: hay que desplegar el SQL arreglado. No borre filas ni haga TRUNCATE." >&2
  exit 1
fi
if [ ! -f "$ETL_DONE_FILE" ]; then
  status=0
  "$MIGRATE_BIN" -necesita-etl || status=$?
  if [ "$status" -eq 0 ]; then
    "$ETL_BIN"
    touch "$ETL_DONE_FILE"
  elif [ "$status" -eq 10 ]; then
    echo "BD con datos: no se ejecuta la carga inicial"
    touch "$ETL_DONE_FILE"
  else
    echo "ERROR al verificar si la base necesita ETL (código $status)" >&2
    exit "$status"
  fi
fi
exec "$API_BIN"
