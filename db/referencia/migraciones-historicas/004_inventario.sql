-- Inventario Fase F. Capas opcionales de data/raw. No reemplaza el catastro.

CREATE OR REPLACE FUNCTION inventario_geom_4326(geojson text)
RETURNS geometry(Geometry, 4326)
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
  g geometry;
BEGIN
  IF geojson IS NULL OR btrim(geojson) = '' OR lower(btrim(geojson)) = 'null' THEN
    RETURN NULL;
  END IF;

  g := ST_MakeValid(ST_Force2D(ST_SetSRID(ST_GeomFromGeoJSON(geojson), 4326)));
  IF g IS NULL OR ST_IsEmpty(g) THEN
    RETURN NULL;
  END IF;
  RETURN g;
END;
$$;

COMMENT ON FUNCTION inventario_geom_4326(text) IS
  'GeoJSON (punto o polígono, lon/lat, con o sin Z) → geometría EPSG:4326.';

CREATE TABLE IF NOT EXISTS inventario (
  id          BIGSERIAL PRIMARY KEY,
  capa        TEXT NOT NULL,
  feature_id  TEXT NOT NULL,
  nombre      TEXT,
  subtipo     TEXT,
  detalle     TEXT,
  lugar       TEXT,
  foto        TEXT,
  geom        geometry(Geometry, 4326),
  UNIQUE (capa, feature_id)
);

CREATE INDEX IF NOT EXISTS inventario_capa_idx ON inventario (capa);
CREATE INDEX IF NOT EXISTS inventario_geom_gix ON inventario USING GIST (geom);

COMMENT ON TABLE inventario IS
  'Overlays legacy: bebederos, fauna, playas, puertas, vereda, flora, cafetos, tachos.';
COMMENT ON COLUMN inventario.foto IS
  'Nombre de archivo JPEG local si existe bajo data/raw/drive_fotos. Nunca un id de Drive.';
