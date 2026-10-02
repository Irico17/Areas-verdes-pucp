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

-- La copia no hereda una medida negativa: el CHECK la rechazaría y pararía el arranque.
WITH src AS (
  SELECT id, feature_id, clase, riego_act, area_m2, perimetro_m, geom
  FROM capas_auxiliares
  WHERE capa = 'xerofitica'
),
ins AS (
  INSERT INTO xerofiticas (feature_id, clase, riego, area_m2, perimetro_m, geom, origen_ref)
  SELECT
    feature_id, clase, riego_act,
    CASE WHEN area_m2 < 0 THEN NULL ELSE area_m2 END,
    CASE WHEN perimetro_m < 0 THEN NULL ELSE perimetro_m END,
    geom, feature_id
  FROM src
  ON CONFLICT (feature_id) DO NOTHING
  RETURNING id, feature_id, area_m2, perimetro_m
)
INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
SELECT
  'xerofiticas', ins.id::text, 'edicion',
  jsonb_build_object('area_m2', src.area_m2, 'perimetro_m', src.perimetro_m, 'capa_id', src.id),
  jsonb_build_object(
    'area_m2', ins.area_m2,
    'perimetro_m', ins.perimetro_m,
    'motivo', 'medida negativa al copiar xerofitica. El original queda en antes. No se borro la capa'
  )
FROM ins
JOIN src ON src.feature_id = ins.feature_id
WHERE src.area_m2 IS DISTINCT FROM ins.area_m2
   OR src.perimetro_m IS DISTINCT FROM ins.perimetro_m;

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

WITH src AS (
  SELECT
    id, feature_id, codigo, nombre, uso, proy_riego, riego_act, referencia,
    pertenecen, perimetro_m, area_m2, geom
  FROM capas_auxiliares
  WHERE capa = 'jardines_reserva'
),
ins AS (
  INSERT INTO jardines_reserva (
    feature_id, codigo, nombre, uso, proy_riego, riego_act, referencia,
    pertenecen, perimetro_m, area_m2, geom, origen_ref
  )
  SELECT
    feature_id, codigo, nombre, uso, proy_riego, riego_act,
    CASE
      WHEN referencia IS NOT NULL AND char_length(referencia) > 500 THEN left(referencia, 500)
      ELSE referencia
    END,
    pertenecen,
    CASE WHEN perimetro_m < 0 THEN NULL ELSE perimetro_m END,
    CASE WHEN area_m2 < 0 THEN NULL ELSE area_m2 END,
    geom, feature_id
  FROM src
  ON CONFLICT (feature_id) DO NOTHING
  RETURNING id, feature_id, referencia, perimetro_m, area_m2
)
INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
SELECT
  'jardines_reserva', ins.id::text, 'edicion',
  jsonb_build_object(
    'referencia', src.referencia,
    'perimetro_m', src.perimetro_m,
    'area_m2', src.area_m2,
    'capa_id', src.id
  ),
  jsonb_build_object(
    'referencia', ins.referencia,
    'perimetro_m', ins.perimetro_m,
    'area_m2', ins.area_m2,
    'motivo', 'la copia no admite medida negativa ni referencia de mas de 500. Corrija desde la ficha. La capa origen sigue intacta'
  )
FROM ins
JOIN src ON src.feature_id = ins.feature_id
WHERE src.area_m2 IS DISTINCT FROM ins.area_m2
   OR src.perimetro_m IS DISTINCT FROM ins.perimetro_m
   OR src.referencia IS DISTINCT FROM ins.referencia;
