-- Área verde: zona de supervisión y validaciones de la ficha (HUID 04).

ALTER TABLE areas_verdes
  ADD COLUMN IF NOT EXISTS zona_supervision_id BIGINT REFERENCES zonas_supervision (id);

ALTER TABLE areas_verdes
  ADD COLUMN IF NOT EXISTS activo BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE areas_verdes
  ADD COLUMN IF NOT EXISTS origen_ref TEXT;

ALTER TABLE areas_verdes DROP CONSTRAINT IF EXISTS areas_verdes_perimetro_chk;
ALTER TABLE areas_verdes ADD CONSTRAINT areas_verdes_perimetro_chk
  CHECK (perimetro_m IS NULL OR perimetro_m >= 0);

ALTER TABLE areas_verdes DROP CONSTRAINT IF EXISTS areas_verdes_area_chk;
ALTER TABLE areas_verdes ADD CONSTRAINT areas_verdes_area_chk
  CHECK (area_m2 IS NULL OR area_m2 >= 0);

ALTER TABLE areas_verdes DROP CONSTRAINT IF EXISTS areas_verdes_referencia_chk;
ALTER TABLE areas_verdes ADD CONSTRAINT areas_verdes_referencia_chk
  CHECK (referencia IS NULL OR char_length(referencia) <= 500);

CREATE UNIQUE INDEX IF NOT EXISTS areas_verdes_codigo_uidx
  ON areas_verdes (codigo)
  WHERE codigo IS NOT NULL AND btrim(codigo) <> '';

CREATE INDEX IF NOT EXISTS areas_verdes_zona_idx ON areas_verdes (zona_supervision_id);

COMMENT ON COLUMN areas_verdes.zona_supervision_id IS
  'Zona de supervisión por intersección. Editable.';
COMMENT ON COLUMN areas_verdes.geom IS
  'MultiPolygon EPSG:4326. NULL = catastro progresivo.';
