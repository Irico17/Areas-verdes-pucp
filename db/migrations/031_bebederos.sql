-- Bebederos con estado y sede (sección 4.12). El HTML descriptio no se guarda.

CREATE TABLE IF NOT EXISTS bebederos (
  id          BIGSERIAL PRIMARY KEY,
  codigo      TEXT NOT NULL,
  subtipo     TEXT NOT NULL,
  estado      TEXT NOT NULL,
  sede        TEXT,
  lat         DOUBLE PRECISION,
  lon         DOUBLE PRECISION,
  foto        TEXT,
  geom        geometry(Point, 4326),
  origen_ref  TEXT NOT NULL,
  activo      BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT bebederos_subtipo_chk CHECK (subtipo IN ('fuente', 'llenador', 'nuevo', 'deterioro', 'baja')),
  CONSTRAINT bebederos_codigo_chk CHECK (codigo ~ '^PT_'),
  CONSTRAINT bebederos_lat_chk CHECK (lat IS NULL OR lat BETWEEN -12.20 AND -11.90),
  CONSTRAINT bebederos_lon_chk CHECK (lon IS NULL OR lon BETWEEN -77.30 AND -76.90)
);

CREATE UNIQUE INDEX IF NOT EXISTS bebederos_origen_ref_uidx ON bebederos (origen_ref);
CREATE UNIQUE INDEX IF NOT EXISTS bebederos_codigo_activo_uidx ON bebederos (codigo) WHERE activo;
CREATE INDEX IF NOT EXISTS bebederos_geom_gix ON bebederos USING GIST (geom);

COMMENT ON TABLE bebederos IS
  'Bebedero. estado viene de la columna del archivo, no solo del nombre. sede es la sede KML.';
COMMENT ON COLUMN bebederos.foto IS
  'Nombre JPEG local si el índice de 70 códigos lo encuentra. No es un id de Drive.';
