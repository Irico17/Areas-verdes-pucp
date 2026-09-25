-- Fauna, puertas y playas en tablas propias. Se copian filas ya cargadas; no se lee GeoJSON.

CREATE TABLE IF NOT EXISTS fauna (
  id          BIGSERIAL PRIMARY KEY,
  feature_id  TEXT NOT NULL UNIQUE,
  nombre      TEXT,
  geom        geometry(Geometry, 4326),
  origen_ref  TEXT,
  activo      BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS fauna_geom_gix ON fauna USING GIST (geom);

COMMENT ON TABLE fauna IS 'Avistamiento. nombre es el del animal, no el de una persona.';

INSERT INTO fauna (feature_id, nombre, geom, origen_ref)
SELECT feature_id, nombre, geom, feature_id
FROM inventario
WHERE capa = 'fauna'
ON CONFLICT (feature_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS puertas (
  id          BIGSERIAL PRIMARY KEY,
  feature_id  TEXT NOT NULL UNIQUE,
  codigo      TEXT,
  nombre      TEXT,
  geom        geometry(Geometry, 4326),
  origen_ref  TEXT,
  activo      BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS puertas_geom_gix ON puertas USING GIST (geom);

COMMENT ON TABLE puertas IS
  'Acceso del campus. nombre queda vacío hasta que lo entreguen.';

INSERT INTO puertas (feature_id, codigo, nombre, geom, origen_ref)
SELECT feature_id, NULLIF(btrim(COALESCE(subtipo, '')), ''), NULLIF(btrim(COALESCE(nombre, '')), ''), geom, feature_id
FROM inventario
WHERE capa = 'puertas'
ON CONFLICT (feature_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS playas_estacionamiento (
  id          BIGSERIAL PRIMARY KEY,
  feature_id  TEXT NOT NULL UNIQUE,
  codigo      TEXT,
  geom        geometry(Geometry, 4326),
  origen_ref  TEXT,
  activo      BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS playas_estacionamiento_geom_gix ON playas_estacionamiento USING GIST (geom);

COMMENT ON TABLE playas_estacionamiento IS 'Playa de estacionamiento.';

INSERT INTO playas_estacionamiento (feature_id, codigo, geom, origen_ref)
SELECT feature_id, NULLIF(btrim(COALESCE(nombre, '')), ''), geom, feature_id
FROM inventario
WHERE capa = 'playas_estacionamiento'
ON CONFLICT (feature_id) DO NOTHING;
