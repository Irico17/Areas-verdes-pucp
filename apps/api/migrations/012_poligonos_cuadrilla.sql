-- Polígonos de cuadrilla. Se copia zonas y se deja la tabla original.
-- La vista zonas conserva la lectura de los clientes actuales.

CREATE TABLE IF NOT EXISTS poligonos_cuadrilla (
  id                    BIGSERIAL PRIMARY KEY,
  feature_id            TEXT NOT NULL UNIQUE,
  source_index          INTEGER NOT NULL UNIQUE,
  codigo                TEXT,
  nombre                TEXT,
  uso                   TEXT,
  proy_riego            TEXT,
  riego_act             TEXT,
  referencia            TEXT,
  perimetro_m           DOUBLE PRECISION,
  area_m2               DOUBLE PRECISION,
  geom                  geometry(MultiPolygon, 4326),
  cuadrilla_id          TEXT REFERENCES cuadrillas (id),
  zona_supervision_id   BIGINT REFERENCES zonas_supervision (id),
  origen_ref            TEXT,
  activo                BOOLEAN NOT NULL DEFAULT TRUE,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT poligonos_perimetro_chk CHECK (perimetro_m IS NULL OR perimetro_m >= 0),
  CONSTRAINT poligonos_area_chk CHECK (area_m2 IS NULL OR area_m2 >= 0),
  CONSTRAINT poligonos_referencia_chk CHECK (referencia IS NULL OR char_length(referencia) <= 500)
);

CREATE INDEX IF NOT EXISTS poligonos_cuadrilla_geom_gix ON poligonos_cuadrilla USING GIST (geom);
CREATE INDEX IF NOT EXISTS poligonos_cuadrilla_cuadrilla_idx ON poligonos_cuadrilla (cuadrilla_id);

COMMENT ON TABLE poligonos_cuadrilla IS
  'Polígono de cuadrilla (antes zonas). feature_id histórico Z- se conserva. Los nuevos usan PC-.';
COMMENT ON COLUMN poligonos_cuadrilla.cuadrilla_id IS
  'Cuadrilla ficticia. Sustituye al campo jefes, que no se persiste.';

INSERT INTO poligonos_cuadrilla (
  id, feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act,
  referencia, perimetro_m, area_m2, geom, created_at, updated_at
)
SELECT
  id, feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act,
  referencia, perimetro_m, area_m2, geom, created_at, updated_at
FROM zonas
ON CONFLICT (id) DO NOTHING;

SELECT setval(
  pg_get_serial_sequence('poligonos_cuadrilla', 'id'),
  COALESCE((SELECT MAX(id) FROM poligonos_cuadrilla), 1),
  (SELECT MAX(id) FROM poligonos_cuadrilla) IS NOT NULL
);

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_class
    WHERE relname = 'zonas' AND relkind = 'r'
  ) THEN
    ALTER TABLE zonas RENAME TO zonas_origen;
  END IF;
END $$;

COMMENT ON TABLE zonas_origen IS
  'Copia de la tabla zonas previa al frente 1A. No se borra.';

CREATE OR REPLACE VIEW zonas AS
SELECT
  id, feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act,
  referencia, perimetro_m, area_m2, geom, created_at, updated_at
FROM poligonos_cuadrilla;

COMMENT ON VIEW zonas IS
  'Compatibilidad de lectura. Los polígonos viven en poligonos_cuadrilla.';

CREATE TABLE IF NOT EXISTS asignaciones_poligono (
  id            BIGSERIAL PRIMARY KEY,
  poligono_id   BIGINT NOT NULL REFERENCES poligonos_cuadrilla (id),
  cuadrilla_id  TEXT NOT NULL REFERENCES cuadrillas (id),
  vigente       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS asignaciones_poligono_poligono_idx ON asignaciones_poligono (poligono_id);

COMMENT ON TABLE asignaciones_poligono IS
  'Asignación de un polígono a una cuadrilla ficticia.';
