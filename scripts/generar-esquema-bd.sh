#!/usr/bin/env bash
# Levanta un PostGIS temporal, aplica db/migrations con cmd/migrate y escribe
# docs/BASE-DE-DATOS.md. No usa AutoMigrate. No toca campus_verde ni el compose
# del repositorio. Borra la base y el contenedor al terminar.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${ESQUEMA_MD:-$ROOT/docs/BASE-DE-DATOS.md}"
IMAGEN="${POSTGIS_IMAGE:-postgis/postgis:16-3.4}"
NOMBRE="vp-esquema-$$"
DB="vp_c_esquema_$(od -An -N3 -tx1 /dev/urandom | tr -d ' \n')"
PORT=""
TMP_TSV=""
TMP_MD=""

if docker info >/dev/null 2>&1; then
  DOCKER=(docker)
elif sudo docker info >/dev/null 2>&1; then
  DOCKER=(sudo docker)
else
  echo "Hace falta el daemon de Docker para levantar ${IMAGEN}." >&2
  exit 1
fi

if [ -x /usr/local/go/bin/go ]; then
  export PATH="/usr/local/go/bin:${PATH}"
fi
command -v go >/dev/null
command -v python3 >/dev/null

limpiar() {
  set +e
  if [ -n "${PORT}" ]; then
    PGPASSWORD=campus psql -h 127.0.0.1 -p "${PORT}" -U campus -d postgres -v ON_ERROR_STOP=1 \
      -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${DB}' AND pid <> pg_backend_pid();" \
      >/dev/null 2>&1
    PGPASSWORD=campus psql -h 127.0.0.1 -p "${PORT}" -U campus -d postgres -v ON_ERROR_STOP=1 \
      -c "DROP DATABASE IF EXISTS ${DB};" >/dev/null 2>&1
  fi
  "${DOCKER[@]}" rm -f "${NOMBRE}" >/dev/null 2>&1
  rm -f "${TMP_TSV}" "${TMP_MD}"
}
trap limpiar EXIT

PORT="$(python3 - <<'PY'
import socket
s = socket.socket()
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()
PY
)"

echo "PostGIS temporal en 127.0.0.1:${PORT}, base ${DB}."
"${DOCKER[@]}" run -d --name "${NOMBRE}" \
  -e POSTGRES_USER=campus \
  -e POSTGRES_PASSWORD=campus \
  -e POSTGRES_DB=postgres \
  -p "127.0.0.1:${PORT}:5432" \
  "${IMAGEN}" >/dev/null

listo=0
for _ in $(seq 1 60); do
  if PGPASSWORD=campus psql -h 127.0.0.1 -p "${PORT}" -U campus -d postgres -At -c "SELECT 1" >/dev/null 2>&1; then
    listo=1
    break
  fi
  sleep 1
done
if [ "${listo}" -ne 1 ]; then
  echo "Postgres no aceptó conexiones desde el host." >&2
  "${DOCKER[@]}" logs "${NOMBRE}" >&2 || true
  exit 1
fi

PGPASSWORD=campus psql -h 127.0.0.1 -p "${PORT}" -U campus -d postgres -v ON_ERROR_STOP=1 \
  -c "CREATE DATABASE ${DB};"

export DATABASE_URL="postgres://campus:campus@127.0.0.1:${PORT}/${DB}?sslmode=disable"
export MIGRATIONS_DIR="${ROOT}/db/migrations"
export CAMPUS_DEV_PASSWORD="clave-demo-local"
export APP_ENV=""
unset SEED_PROFILE || true

(cd "${ROOT}/backend/app" && go run ./cmd/migrate)

n_sql="$(find "${MIGRATIONS_DIR}" -maxdepth 1 -name '*.sql' | wc -l | tr -d ' ')"
n_apl="$(PGPASSWORD=campus psql -h 127.0.0.1 -p "${PORT}" -U campus -d "${DB}" -At -c "SELECT count(*) FROM schema_migrations;")"
if [ "${n_sql}" != "${n_apl}" ]; then
  echo "Inconsistente: ${n_sql} archivos SQL y ${n_apl} filas en schema_migrations." >&2
  exit 1
fi

TMP_TSV="$(mktemp)"
TMP_MD="$(mktemp)"
PGPASSWORD=campus psql -h 127.0.0.1 -p "${PORT}" -U campus -d "${DB}" -v ON_ERROR_STOP=1 -At -F $'\t' > "${TMP_TSV}" <<'SQL'
SELECT 'COL', c.table_name, c.column_name, c.data_type, c.udt_name, c.is_nullable
FROM information_schema.columns c
JOIN information_schema.tables t
  ON t.table_schema = c.table_schema AND t.table_name = c.table_name
WHERE c.table_schema = 'public'
  AND t.table_type = 'BASE TABLE'
  AND c.table_name <> 'spatial_ref_sys'
ORDER BY c.table_name, c.ordinal_position;

SELECT 'PK', tc.table_name, kcu.column_name, '', '', ''
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON kcu.constraint_name = tc.constraint_name
 AND kcu.table_schema = tc.table_schema
 AND kcu.table_name = tc.table_name
WHERE tc.table_schema = 'public'
  AND tc.constraint_type = 'PRIMARY KEY'
  AND tc.table_name <> 'spatial_ref_sys'
ORDER BY tc.table_name, kcu.ordinal_position;

SELECT 'FK', tc.table_name, kcu.column_name, ccu.table_name, ccu.column_name, col.is_nullable
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON kcu.constraint_name = tc.constraint_name
 AND kcu.table_schema = tc.table_schema
 AND kcu.table_name = tc.table_name
JOIN information_schema.constraint_column_usage ccu
  ON ccu.constraint_name = tc.constraint_name
 AND ccu.table_schema = tc.table_schema
JOIN information_schema.columns col
  ON col.table_schema = kcu.table_schema
 AND col.table_name = kcu.table_name
 AND col.column_name = kcu.column_name
WHERE tc.table_schema = 'public'
  AND tc.constraint_type = 'FOREIGN KEY'
  AND tc.table_name <> 'spatial_ref_sys'
  AND ccu.table_name <> 'spatial_ref_sys'
ORDER BY tc.table_name, kcu.column_name, ccu.table_name;
SQL

python3 - "${TMP_TSV}" "${TMP_MD}" "${n_apl}" <<'PY'
import sys
from collections import defaultdict

src, dest, n_apl = sys.argv[1], sys.argv[2], sys.argv[3]
cols = defaultdict(list)
pks = defaultdict(set)
fks = []
seen_fk = set()

with open(src, encoding="utf-8") as fh:
    for line in fh:
        parts = line.rstrip("\n").split("\t")
        if len(parts) < 6:
            continue
        kind, table, column, a, b, nullable = parts[:6]
        if kind == "COL":
            cols[table].append((column, a, b, nullable))
        elif kind == "PK":
            pks[table].add(column)
        elif kind == "FK":
            key = (table, column, a, b)
            if key in seen_fk:
                continue
            seen_fk.add(key)
            fks.append((table, column, a, b, nullable))

if not cols:
    raise SystemExit("no se leyeron tablas")

def tipo(data_type, udt):
    if data_type == "USER-DEFINED":
        return udt
    return data_type

def tipo_mermaid(data_type, udt):
    raw = tipo(data_type, udt)
    return {
        "character varying": "varchar",
        "timestamp with time zone": "timestamptz",
        "timestamp without time zone": "timestamp",
        "double precision": "float8",
        "character": "char",
    }.get(raw, raw.replace(" ", "_"))

def ident(name):
    return name.replace('"', "")

lines = []
lines.append("# Base de datos")
lines.append("")
lines.append("v0 hasta la Ola 5. Generado por `scripts/generar-esquema-bd.sh` a partir de un PostGIS vacío con las migraciones de `db/migrations` aplicadas por `cmd/migrate`. No es el esquema definitivo.")
lines.append("")
lines.append(f"Migraciones aplicadas: **{n_apl}**. No hay filas de negocio en este documento. Se omiten `spatial_ref_sys` y las vistas del catálogo de PostGIS.")
lines.append("")
lines.append("## Tablas")
lines.append("")
lines.append("| Tabla | Columnas | Clave primaria |")
lines.append("| --- | ---: | --- |")
for table in sorted(cols):
    pk = ", ".join(sorted(pks[table])) or "—"
    lines.append(f"| `{table}` | {len(cols[table])} | `{pk}` |")
lines.append("")
lines.append("## Columnas")
lines.append("")
for table in sorted(cols):
    lines.append(f"### `{table}`")
    lines.append("")
    lines.append("| Columna | Tipo | Nulo | Clave |")
    lines.append("| --- | --- | --- | --- |")
    fk_cols = {col for t, col, _, _, _ in fks if t == table}
    for column, data_type, udt, nullable in cols[table]:
        marcas = []
        if column in pks[table]:
            marcas.append("PK")
        if column in fk_cols:
            marcas.append("FK")
        lines.append(
            f"| `{column}` | `{tipo(data_type, udt)}` | {nullable} | {' '.join(marcas) or '—'} |"
        )
    lines.append("")
lines.append("## Relaciones")
lines.append("")
if not fks:
    lines.append("No hay claves foráneas.")
    lines.append("")
else:
    lines.append("| Tabla | Columna | Referencia | Nulo |")
    lines.append("| --- | --- | --- | --- |")
    for table, column, ref, ref_col, nullable in fks:
        lines.append(f"| `{table}` | `{column}` | `{ref}.{ref_col}` | {nullable} |")
    lines.append("")
lines.append("## Diagrama")
lines.append("")
lines.append("```mermaid")
lines.append("erDiagram")
for table in sorted(cols):
    lines.append(f"    {ident(table)} {{")
    for column, data_type, udt, _nullable in cols[table]:
        marcas = []
        if column in pks[table]:
            marcas.append("PK")
        if any(t == table and col == column for t, col, _, _, _ in fks):
            marcas.append("FK")
        suffix = (" " + " ".join(marcas)) if marcas else ""
        lines.append(f"        {tipo_mermaid(data_type, udt)} {ident(column)}{suffix}")
    lines.append("    }")
for table, column, ref, _ref_col, nullable in fks:
    card = "||--o{" if nullable == "YES" else "||--|{"
    lines.append(f"    {ident(ref)} {card} {ident(table)} : {ident(column)}")
lines.append("```")
lines.append("")
with open(dest, "w", encoding="utf-8") as fh:
    fh.write("\n".join(lines))
PY

if ! grep -q '^# Base de datos' "${TMP_MD}"; then
  echo "La salida no es el markdown esperado." >&2
  exit 1
fi
if ! grep -q 'erDiagram' "${TMP_MD}"; then
  echo "Falta el diagrama." >&2
  exit 1
fi
if grep -Eq 'clave-demo-local|pando-local|AKIA' "${TMP_MD}"; then
  echo "La salida contiene una clave. No se publica." >&2
  exit 1
fi

mkdir -p "$(dirname "${OUT}")"
cp "${TMP_MD}" "${OUT}"
echo "Escrito ${OUT} (${n_apl} migraciones)."
