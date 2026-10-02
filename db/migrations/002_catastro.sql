-- Catastro Fase B.
-- CRS de almacenamiento y servicio: EPSG:4326 (lon/lat), compatible con MapLibre / RFC 7946.
-- geom NULL está permitido: el catastro es progresivo (un área o zona puede existir sin GPS).

CREATE OR REPLACE FUNCTION catastro_geom_4326(geojson text)
RETURNS geometry(MultiPolygon, 4326)
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
  g geometry;
BEGIN
  IF geojson IS NULL OR btrim(geojson) = '' OR lower(btrim(geojson)) = 'null' THEN
    RETURN NULL;
  END IF;

  g := ST_CollectionExtract(
    ST_MakeValid(
      ST_Force2D(
        ST_SetSRID(ST_GeomFromGeoJSON(geojson), 4326)
      )
    ),
    3
  );

  IF g IS NULL OR ST_IsEmpty(g) THEN
    RETURN NULL;
  END IF;

  RETURN ST_Multi(g)::geometry(MultiPolygon, 4326);
END;
$$;

COMMENT ON FUNCTION catastro_geom_4326(text) IS
  'GeoJSON (Polygon o MultiPolygon, lon/lat) → MultiPolygon EPSG:4326. NULL si viene vacío.';

CREATE TABLE IF NOT EXISTS areas_verdes (
  id            BIGSERIAL PRIMARY KEY,
  feature_id    TEXT NOT NULL UNIQUE,
  source_index  INTEGER NOT NULL UNIQUE,
  codigo        TEXT,
  nombre        TEXT,
  uso           TEXT,
  proy_riego    TEXT,
  riego_act     TEXT,
  referencia    TEXT,
  perimetro_m   DOUBLE PRECISION,
  area_m2       DOUBLE PRECISION,
  geom          geometry(MultiPolygon, 4326),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS areas_verdes_geom_gix ON areas_verdes USING GIST (geom);

COMMENT ON TABLE areas_verdes IS
  'Áreas verdes de catastro. Semilla: data/raw/areas_verdes.geojson.';
COMMENT ON COLUMN areas_verdes.geom IS
  'MultiPolygon EPSG:4326. NULL = geometría progresiva (registro sin GPS).';
COMMENT ON COLUMN areas_verdes.feature_id IS
  'Identificador estable AV-NNNN, independiente del id serial.';

CREATE TABLE IF NOT EXISTS zonas (
  id            BIGSERIAL PRIMARY KEY,
  feature_id    TEXT NOT NULL UNIQUE,
  source_index  INTEGER NOT NULL UNIQUE,
  codigo        TEXT,
  nombre        TEXT,
  uso           TEXT,
  proy_riego    TEXT,
  riego_act     TEXT,
  referencia    TEXT,
  perimetro_m   DOUBLE PRECISION,
  area_m2       DOUBLE PRECISION,
  geom          geometry(MultiPolygon, 4326),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS zonas_geom_gix ON zonas USING GIST (geom);

COMMENT ON TABLE zonas IS
  'Zonas de catastro. Semilla: data/raw/jefe_de_grupo.json sin el campo jefes (PII).';
COMMENT ON COLUMN zonas.feature_id IS
  'Identificador de zona Z-NNNN. No es un nombre de persona.';
COMMENT ON COLUMN zonas.geom IS
  'MultiPolygon EPSG:4326. NULL = geometría progresiva.';

CREATE TABLE IF NOT EXISTS capas_auxiliares (
  id            BIGSERIAL PRIMARY KEY,
  capa          TEXT NOT NULL,
  feature_id    TEXT NOT NULL UNIQUE,
  source_index  INTEGER NOT NULL,
  codigo        TEXT,
  nombre        TEXT,
  uso           TEXT,
  proy_riego    TEXT,
  clase         TEXT,
  riego_act     TEXT,
  referencia    TEXT,
  pertenecen    TEXT,
  perimetro_m   DOUBLE PRECISION,
  area_m2       DOUBLE PRECISION,
  geom          geometry(MultiPolygon, 4326),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT capas_auxiliares_capa_source_unique UNIQUE (capa, source_index)
);

CREATE INDEX IF NOT EXISTS capas_auxiliares_geom_gix ON capas_auxiliares USING GIST (geom);
CREATE INDEX IF NOT EXISTS capas_auxiliares_capa_idx ON capas_auxiliares (capa);

COMMENT ON TABLE capas_auxiliares IS
  'Capas opcionales de catastro: jardines_reserva y xerofitica.';
COMMENT ON COLUMN capas_auxiliares.pertenecen IS
  'Unidad organizacional (p. ej. DAF), no una persona.';
