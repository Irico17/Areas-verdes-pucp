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

# Rechazo temprano de claves de laboratorio en producción (antes de migrar).
norm_env="$(printf "%s" "${APP_ENV:-}" | tr '[:upper:]' '[:lower:]')"
case "$norm_env" in
  produccion|producción|production|prod)
    trimmed_dev_pass="$(printf "%s" "$CAMPUS_DEV_PASSWORD" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
    if [ "$trimmed_dev_pass" = "pando-local" ] || [ "$trimmed_dev_pass" = "campus-lab" ]; then
      echo "ERROR: en producción, CAMPUS_DEV_PASSWORD no puede ser una clave de laboratorio." >&2
      exit 1
    fi
    if [ "${#trimmed_dev_pass}" -lt 16 ]; then
      echo "ERROR: en producción, CAMPUS_DEV_PASSWORD debe tener al menos 16 caracteres (longitud actual: ${#trimmed_dev_pass})." >&2
      exit 1
    fi

    if [ -n "${DATABASE_URL:-}" ]; then
      pg_pass="$(printf "%s" "$DATABASE_URL" | sed -n 's|^[^:]*://[^:]*:\([^@]*\)@.*$|\1|p')"
      if [ -n "$pg_pass" ]; then
        trimmed_pg_pass="$(printf "%s" "$pg_pass" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')"
        if [ "$trimmed_pg_pass" = "pando-local" ] || [ "$trimmed_pg_pass" = "campus-lab" ]; then
          echo "ERROR: en producción, la clave de Postgres en DATABASE_URL no puede ser una clave de laboratorio." >&2
          exit 1
        fi
        if [ "${#trimmed_pg_pass}" -lt 16 ]; then
          echo "ERROR: en producción, la clave de Postgres en DATABASE_URL debe tener al menos 16 caracteres (longitud actual: ${#trimmed_pg_pass})." >&2
          exit 1
        fi
      fi
    fi
    ;;
esac

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

# etl: carga data/raw. Base vacía → carga inicial. Catastro a medias (p. ej. un
# solo polígono ficticio) → etl-lote, upsert, sin TRUNCATE. Catastro publicado
# (521 áreas y 534 sectores) → no recarga.
# develop y qa, o SEED_PROFILE=ficticio: además inserta polígonos de demostración
# (ON CONFLICT DO NOTHING). Producción no recibe esa semilla.
# Ninguno borra filas ya cargadas. .etl-done no impide completar un catastro a medias.
SEED_PROFILE="${SEED_PROFILE:-etl}"
SEED_FILE="${SEED_FILE:-/opt/campus/seed/ficticio.sql}"
ETL_LOTE_BIN="${ETL_LOTE_BIN:-/usr/local/bin/etl-lote}"
export SEED_FILE

aplicar_semilla=0
case "$SEED_PROFILE" in
  ficticio) aplicar_semilla=1 ;;
esac
case "$norm_env" in
  develop|qa) aplicar_semilla=1 ;;
esac
if [ "$aplicar_semilla" -eq 1 ]; then
  if ! "$MIGRATE_BIN" -semilla-ficticia; then
    echo "ERROR: no se pudo aplicar la semilla ficticia. No se borra nada." >&2
    exit 1
  fi
fi

case "$SEED_PROFILE" in
  ficticio)
    touch "$ETL_DONE_FILE"
    ;;
  etl)
    incompleto=0
    "$MIGRATE_BIN" -catastro-incompleto || incompleto=$?
    if [ "$incompleto" -eq 10 ]; then
      echo "Catastro completo: no se vuelve a cargar"
      touch "$ETL_DONE_FILE"
    elif [ "$incompleto" -eq 0 ]; then
      vacio=0
      "$MIGRATE_BIN" -necesita-etl || vacio=$?
      if [ "$vacio" -eq 0 ]; then
        "$ETL_BIN"
        touch "$ETL_DONE_FILE"
      elif [ "$vacio" -eq 10 ]; then
        echo "Catastro incompleto: upsert con etl-lote, sin TRUNCATE"
        "$ETL_LOTE_BIN"
        touch "$ETL_DONE_FILE"
      else
        echo "ERROR al verificar si la base necesita ETL (código $vacio)" >&2
        exit "$vacio"
      fi
    else
      echo "ERROR al verificar si el catastro está completo (código $incompleto)" >&2
      exit "$incompleto"
    fi
    ;;
  *)
    echo "ERROR: SEED_PROFILE=$SEED_PROFILE no es etl ni ficticio" >&2
    exit 1
    ;;
esac
exec "$API_BIN"
