#!/bin/sh
set -eu
mkdir -p /data/evidencias /data/v1
if ! /usr/local/bin/migrate; then
  echo "FALLO DE MIGRACION: la API no arranca. El error de arriba nombra el archivo y la sentencia. Reiniciar el contenedor no lo corrige: hay que desplegar el SQL arreglado. No borre filas ni haga TRUNCATE." >&2
  exit 1
fi
if [ ! -f /data/.etl-done ]; then
  /usr/local/bin/etl
  touch /data/.etl-done
fi
exec /usr/local/bin/api
