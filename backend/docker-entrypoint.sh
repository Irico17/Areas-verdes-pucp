#!/bin/sh
# Compatible con el sh de busybox (Alpine). Sin bash.
set -eu
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
