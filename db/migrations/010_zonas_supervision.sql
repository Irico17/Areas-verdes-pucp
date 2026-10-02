-- Zonas de supervisión Z1–Z4 (HUID 06 y 33). Vacías: la carga es el frente 1E.

CREATE TABLE IF NOT EXISTS zonas_supervision (
  id          BIGSERIAL PRIMARY KEY,
  codigo      TEXT NOT NULL UNIQUE,
  nombre      TEXT NOT NULL,
  area_m2     DOUBLE PRECISION,
  geom        geometry(MultiPolygon, 4326) NOT NULL,
  activo      BOOLEAN NOT NULL DEFAULT TRUE,
  origen_ref  TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT zonas_supervision_codigo_chk CHECK (codigo IN ('Z1', 'Z2', 'Z3', 'Z4')),
  CONSTRAINT zonas_supervision_area_chk CHECK (area_m2 IS NULL OR area_m2 >= 0)
);

CREATE INDEX IF NOT EXISTS zonas_supervision_geom_gix ON zonas_supervision USING GIST (geom);

COMMENT ON TABLE zonas_supervision IS
  'Cuatro zonas de supervisión del campus. No son los polígonos de cuadrilla.';
