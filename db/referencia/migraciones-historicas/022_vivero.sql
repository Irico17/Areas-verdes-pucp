-- Frente 2A. Vivero: área de catálogo y subproceso/etapa dados de alta al importar.

CREATE TABLE IF NOT EXISTS vivero_catalogo (
  id      BIGSERIAL PRIMARY KEY,
  clase   TEXT NOT NULL,
  nombre  TEXT NOT NULL,
  activo  BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE (clase, nombre)
);

COMMENT ON TABLE vivero_catalogo IS
  'Área, subproceso y etapa. Después de importar no se escribe texto libre.';

INSERT INTO vivero_catalogo (clase, nombre) VALUES
  ('area', 'Fauna'),
  ('area', 'Flora'),
  ('area', 'Ambiental'),
  ('area', 'Otros')
ON CONFLICT (clase, nombre) DO NOTHING;

CREATE TABLE IF NOT EXISTS vivero_registros (
  id              UUID PRIMARY KEY,
  fecha           DATE,
  area            TEXT NOT NULL DEFAULT '',
  subproceso      TEXT NOT NULL DEFAULT '',
  etapa           TEXT NOT NULL DEFAULT '',
  descripcion     TEXT NOT NULL DEFAULT '',
  observaciones   TEXT NOT NULL DEFAULT '',
  responsables    TEXT NOT NULL DEFAULT '',
  lugar_id        TEXT,
  lugar_libre     TEXT NOT NULL DEFAULT '',
  origen_ref      TEXT,
  archivada_en    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS vivero_origen_ref_uidx
  ON vivero_registros (origen_ref) WHERE origen_ref IS NOT NULL AND origen_ref <> '';

CREATE INDEX IF NOT EXISTS vivero_fecha_idx ON vivero_registros (fecha);

COMMENT ON COLUMN vivero_registros.responsables IS
  'Lista de nombres ficticios. No hay personas reales.';
COMMENT ON COLUMN vivero_registros.lugar_libre IS
  'Texto de lugar cuando el nombre no calza con un lugar conocido.';
