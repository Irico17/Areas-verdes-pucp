-- Vereda en riesgo, xerofítica y jardín de reserva. Copia de lo ya cargado.

CREATE TABLE IF NOT EXISTS veredas_riesgo (
  id          BIGSERIAL PRIMARY KEY,
  feature_id  TEXT NOT NULL UNIQUE,
  nota        TEXT,
  geom        geometry(Geometry, 4326),
  origen_ref  TEXT,
  activo      BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS veredas_riesgo_geom_gix ON veredas_riesgo USING GIST (geom);

COMMENT ON TABLE veredas_riesgo IS 'Tramo de vereda en riesgo.';

INSERT INTO veredas_riesgo (feature_id, nota, geom, origen_ref)
SELECT feature_id, NULLIF(btrim(COALESCE(detalle, '')), ''), geom, feature_id
FROM inventario
WHERE capa = 'area_vereda_peligro'
ON CONFLICT (feature_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS xerofiticas (
  id           BIGSERIAL PRIMARY KEY,
  feature_id   TEXT NOT NULL UNIQUE,
  clase        TEXT,
  riego        TEXT,
  area_m2      DOUBLE PRECISION,
  perimetro_m  DOUBLE PRECISION,
  geom         geometry(MultiPolygon, 4326),
  origen_ref   TEXT,
  activo       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT xerofiticas_area_chk CHECK (area_m2 IS NULL OR area_m2 >= 0),
  CONSTRAINT xerofiticas_perimetro_chk CHECK (perimetro_m IS NULL OR perimetro_m >= 0)
);

CREATE INDEX IF NOT EXISTS xerofiticas_geom_gix ON xerofiticas USING GIST (geom);

COMMENT ON TABLE xerofiticas IS 'Área xerofítica.';

INSERT INTO xerofiticas (feature_id, clase, riego, area_m2, perimetro_m, geom, origen_ref)
SELECT feature_id, clase, riego_act, area_m2, perimetro_m, geom, feature_id
FROM capas_auxiliares
WHERE capa = 'xerofitica'
ON CONFLICT (feature_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS jardines_reserva (
  id           BIGSERIAL PRIMARY KEY,
  feature_id   TEXT NOT NULL UNIQUE,
  codigo       TEXT,
  nombre       TEXT,
  uso          TEXT,
  proy_riego   TEXT,
  riego_act    TEXT,
  referencia   TEXT,
  pertenecen   TEXT,
  perimetro_m  DOUBLE PRECISION,
  area_m2      DOUBLE PRECISION,
  geom         geometry(MultiPolygon, 4326),
  origen_ref   TEXT,
  activo       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT jardines_reserva_area_chk CHECK (area_m2 IS NULL OR area_m2 >= 0),
  CONSTRAINT jardines_reserva_perimetro_chk CHECK (perimetro_m IS NULL OR perimetro_m >= 0),
  CONSTRAINT jardines_reserva_referencia_chk CHECK (referencia IS NULL OR char_length(referencia) <= 500)
);

CREATE INDEX IF NOT EXISTS jardines_reserva_geom_gix ON jardines_reserva USING GIST (geom);

COMMENT ON TABLE jardines_reserva IS
  'Jardín reservable. pertenecen es una unidad organizativa, no una persona.';

INSERT INTO jardines_reserva (
  feature_id, codigo, nombre, uso, proy_riego, riego_act, referencia,
  pertenecen, perimetro_m, area_m2, geom, origen_ref
)
SELECT
  feature_id, codigo, nombre, uso, proy_riego, riego_act, referencia,
  pertenecen, perimetro_m, area_m2, geom, feature_id
FROM capas_auxiliares
WHERE capa = 'jardines_reserva'
ON CONFLICT (feature_id) DO NOTHING;
