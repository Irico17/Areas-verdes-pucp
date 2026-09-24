#!/bin/sh
set -eu
mkdir -p /data/evidencias /data/v1
/usr/local/bin/migrate
if [ ! -f /data/.etl-done ]; then
  /usr/local/bin/etl
  touch /data/.etl-done
fi
exec /usr/local/bin/api
