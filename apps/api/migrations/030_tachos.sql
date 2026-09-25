-- Tachos con los 11 conteos (sección 4.11). Sin TRUNCATE.

CREATE TABLE IF NOT EXISTS tachos (
  id                    BIGSERIAL PRIMARY KEY,
  codigo                TEXT NOT NULL,
  lat                   DOUBLE PRECISION,
  lon                   DOUBLE PRECISION,
  nota                  TEXT,
  lugar                 TEXT,
  espacios              TEXT,
  accion                TEXT,
  tacho_actual          TEXT,
  tacho_nuevo           TEXT,
  recomendaciones       TEXT,
  no_aprovechables      INTEGER NOT NULL DEFAULT 0,
  papel_carton          INTEGER NOT NULL DEFAULT 0,
  plastico              INTEGER NOT NULL DEFAULT 0,
  vidrio                INTEGER NOT NULL DEFAULT 0,
  pilas                 INTEGER NOT NULL DEFAULT 0,
  peligrosos            INTEGER NOT NULL DEFAULT 0,
  raee                  INTEGER NOT NULL DEFAULT 0,
  metales               INTEGER NOT NULL DEFAULT 0,
  aniquem               INTEGER NOT NULL DEFAULT 0,
  intermedios_plastico  INTEGER NOT NULL DEFAULT 0,
  intermedios_metal     INTEGER NOT NULL DEFAULT 0,
  foto                  TEXT,
  zona_supervision_id   BIGINT REFERENCES zonas_supervision (id),
  geom                  geometry(Point, 4326),
  origen_ref            TEXT NOT NULL,
  activo                BOOLEAN NOT NULL DEFAULT TRUE,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT tachos_codigo_chk CHECK (codigo ~ '^PT'),
  CONSTRAINT tachos_lat_chk CHECK (lat IS NULL OR lat BETWEEN -12.20 AND -11.90),
  CONSTRAINT tachos_lon_chk CHECK (lon IS NULL OR lon BETWEEN -77.30 AND -76.90),
  CONSTRAINT tachos_conteos_chk CHECK (
    no_aprovechables >= 0 AND papel_carton >= 0 AND plastico >= 0 AND vidrio >= 0
    AND pilas >= 0 AND peligrosos >= 0 AND raee >= 0 AND metales >= 0
    AND aniquem >= 0 AND intermedios_plastico >= 0 AND intermedios_metal >= 0
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS tachos_origen_ref_uidx ON tachos (origen_ref);
CREATE UNIQUE INDEX IF NOT EXISTS tachos_codigo_activo_uidx ON tachos (codigo) WHERE activo;
CREATE INDEX IF NOT EXISTS tachos_geom_gix ON tachos USING GIST (geom);

COMMENT ON TABLE tachos IS
  'Punto de residuos. Los 11 conteos son columnas propias. foto es un archivo local, nunca una URL de Drive.';
