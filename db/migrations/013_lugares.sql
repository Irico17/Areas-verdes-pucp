-- Lugares del campus (HUID 06). Sin filas: el CSV lo carga el frente 1E.

CREATE TABLE IF NOT EXISTS lugares (
  id                    BIGSERIAL PRIMARY KEY,
  nombre                TEXT NOT NULL,
  nombre_norm           TEXT NOT NULL UNIQUE,
  lat                   DOUBLE PRECISION NOT NULL,
  lon                   DOUBLE PRECISION NOT NULL,
  zona_supervision_id   BIGINT REFERENCES zonas_supervision (id),
  origen_ref            TEXT,
  activo                BOOLEAN NOT NULL DEFAULT TRUE,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT lugares_lat_chk CHECK (lat BETWEEN -12.20 AND -11.90),
  CONSTRAINT lugares_lon_chk CHECK (lon BETWEEN -77.30 AND -76.90)
);

CREATE INDEX IF NOT EXISTS lugares_zona_idx ON lugares (zona_supervision_id);

COMMENT ON TABLE lugares IS
  'Diccionario de lugares. nombre_norm es único, en minúsculas y sin tildes.';
COMMENT ON COLUMN lugares.nombre_norm IS
  'Nombre normalizado para no duplicar el mismo lugar con otra grafía.';
