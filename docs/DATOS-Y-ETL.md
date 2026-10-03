# Datos y ETL

La fuente de verdad del esquema es `db/migrations`. El cargador que se ejecuta es el de `backend/app` (`cmd/etl`, `cmd/etl-lote`, `cmd/migrate`). Desde la raíz: `make etl` y `make etl-lote`.

No se usa `AutoMigrate`. Postgres no monta SQL en `docker-entrypoint-initdb.d`. `db/referencia/` no se aplica.

## Directorios

| Ruta | Qué es |
| --- | --- |
| `data/raw` | GeoJSON, JSON y CSV recuperados. No se editan para «arreglar» el mapa. |
| `data/raw/lote` | Fuentes del frente de carga por lotes (`etl-lote`). |
| `data/v1` | GeoJSON normalizado. Lo escribe el ETL. |
| `data/osm/edificios_pando.geojson` | Huellas de edificios para la vista Relieve. |
| `data/mocks` | Agenda de reservas ficticia. No es la hoja institucional. |
| `deploy/seed/ficticio.sql` | Polígonos de demostración en el parque oeste. `ON CONFLICT` solo reescribe esas filas. |

Variables: `DATA_RAW_DIR`, `DATA_V1_DIR`, `EDIFICIOS_PATH`, `RESERVAS_MOCK_PATH`, `DRIVE_FOTOS_DIR`, `MIGRATIONS_DIR`, `SEED_FILE`. Los valores locales están comentados en `.env.example`.

## Migraciones

Cada archivo `NNN_*.sql` se aplica una vez, en su propia transacción, y queda anotado en `schema_migrations`. Volver a correr `migrate` no las repite. No hay migraciones «down»: un rollback de imagen no quita columnas.

La serie vigente son dos archivos. `001_esquema_base.sql` crea extensiones, tipos, tablas, índices, restricciones, vistas, funciones y triggers. `002_catalogos_base.sql` inserta los catálogos y los datos de referencia que la serie histórica dejaba en una base vacía; cada `INSERT` usa `ON CONFLICT DO NOTHING`. La semilla de demostración del despliegue sigue en `deploy/seed/ficticio.sql` y no entra en las migraciones. Usuarios y sesiones los crea la semilla de accesos, no el SQL.

La serie `001_postgis.sql` … `078_medidas_palmera_baja.sql` está en `db/referencia/migraciones-historicas/` y ya no se ejecuta. Sirve para la transición y para las pruebas que recorren una base a medias. Si `schema_migrations` ya tiene `078_medidas_palmera_baja.sql`, `001` y `002` se anotan como aplicadas y no se ejecutan; el historial viejo permanece. Una serie a medias, sin ese archivo, no recibe la consolidada. Los cambios siguientes van en `003_*.sql` y siguientes: son aditivos, no se reordenan y no se renombran.

Desde la raíz, con Postgres arriba y `.env` cargado:

```bash
make migrate
```

Equivalente:

```bash
cd backend/app
go run ./cmd/migrate
```

Flags:

| Flag | Efecto |
| --- | --- |
| (ninguno) | Aplica lo pendiente y asegura las seis cuentas ficticias. Exige `CAMPUS_DEV_PASSWORD`. |
| `-necesita-etl` | Sale 0 si la base está vacía y 10 si ya hay datos. |
| `-catastro-incompleto` | Sale 0 si faltan áreas con geometría o sectores respecto del catastro publicado (521 y 534). Sale 10 si ya está. |
| `-semilla-ficticia` | Aplica `deploy/seed/ficticio.sql` después de migrar. |

`Ensure` crea `norte`, `sur`, `riego`, `coordinacion`, `jefatura` y `admin` si no existen. No reescribe el hash de una cuenta que ya está.

## etl

```bash
make etl
```

Normaliza `data/raw` hacia `data/v1` y carga PostGIS. Conteos que exige el modo estricto: **521** áreas y **534** zonas. Jardines de reserva: 21. Xerofítica: 10.

```bash
cd backend/app
go run ./cmd/etl --skip-load          # solo escribe data/v1
go run ./cmd/etl --no-strict          # no exige 521 / 534
go run ./cmd/etl --raw-dir ... --v1-dir ...
```

El ETL histórico puede truncar el catastro semilla cuando se lanza a mano sobre una base de desarrollo. En el arranque del contenedor no se usa así: ver la sección siguiente.

`make sectores` regenera `data/v1/zonas_sector.json` desde `data/raw/lote/jefe_de_grupo.json`.

## Qué hace el contenedor al arrancar

`backend/docker-entrypoint.sh` llama a `migrate` y después decide la carga. `SEED_PROFILE` vale `etl` o `ficticio`.

1. Si el catastro visible ya tiene 521 áreas con geometría y 534 sectores, no recarga.
2. Si la base está vacía, corre `etl` y marca `/data/.etl-done`.
3. Si hay filas pero el catastro está incompleto, corre `etl-lote` (upsert, sin `TRUNCATE`). Si ese comando falla, la API arranca igual y no borra lo cargado.
4. En `develop`, `qa`, o con `SEED_PROFILE=ficticio`, aplica la semilla ficticia **después** de la carga. Producción no la recibe.

No copiar un volcado de producción a develop ni a QA. Este repositorio no trae un script que haga esa copia.

## etl-lote

```bash
make etl-lote
```

Upsert de zonas de supervisión, polígonos de cuadrilla, lugares, especies, ejemplares, medidas, cafetos y catálogos que viven bajo `data/raw/lote`. No trunca.

```bash
cd backend/app
go run ./cmd/etl-lote -solo-lectura
```

Eso solo resuelve las fuentes y cuenta filas.

## Conteos

```bash
make counts
```

Con un ambiente de `deploy/`: `make counts ENV=develop`.

## Esquema legible

`scripts/generar-esquema-bd.sh` aplica estas migraciones sobre un PostGIS vacío y escribe `db/esquema.sql` y `docs/BASE-DE-DATOS.md`. `db/esquema.sql` es la foto del resultado, sin dueños ni privilegios. La fuente de verdad siguen siendo las migraciones. `bash scripts/generar-esquema-bd.sh --comprobar` falla si alguno de los dos archivos se desvía. El control de deriva de `modelgen` sigue comparando los modelos Go con la misma base.
