-- Capa de vías, vacía a propósito. Se llena solo con un GeoJSON importado.
-- No hay filas de ejemplo ni geometría inventada.

CREATE TABLE IF NOT EXISTS vias (
  id          BIGSERIAL PRIMARY KEY,
  feature_id  TEXT NOT NULL,
  nombre      TEXT,
  geom        geometry(Geometry, 4326),
  activo      BOOLEAN NOT NULL DEFAULT TRUE,
  origen_ref  TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS vias_feature_id_uidx ON vias (feature_id);
CREATE INDEX IF NOT EXISTS vias_geom_gix ON vias USING GIST (geom);

COMMENT ON TABLE vias IS
  'Referente lineal. Nace vacía. La capa del mapa permanece apagada hasta que se importe un GeoJSON.';
