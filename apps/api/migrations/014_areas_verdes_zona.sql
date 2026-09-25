-- Área verde: zona de supervisión y validaciones de la ficha (HUID 04).
-- El catastro real repite códigos (F 26, D 14, …). No se borra ninguna fila:
-- el menor id conserva el código y los demás reciben el sufijo « ·<id>»,
-- con el antes/después en cambios para corregirlo desde la ficha.

DO $$
DECLARE
  r RECORD;
  nuevo TEXT;
  n INT;
BEGIN
  FOR r IN
    SELECT id, perimetro_m, area_m2, referencia
    FROM areas_verdes
    WHERE (perimetro_m IS NOT NULL AND perimetro_m < 0)
       OR (area_m2 IS NOT NULL AND area_m2 < 0)
       OR (referencia IS NOT NULL AND char_length(referencia) > 500)
    ORDER BY id
  LOOP
    INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
    VALUES (
      'areas_verdes', r.id::text, 'edicion',
      jsonb_build_object('perimetro_m', r.perimetro_m, 'area_m2', r.area_m2, 'referencia', r.referencia),
      jsonb_build_object(
        'perimetro_m', CASE WHEN r.perimetro_m < 0 THEN NULL ELSE r.perimetro_m END,
        'area_m2', CASE WHEN r.area_m2 < 0 THEN NULL ELSE r.area_m2 END,
        'referencia', CASE
          WHEN r.referencia IS NOT NULL AND char_length(r.referencia) > 500 THEN left(r.referencia, 500)
          ELSE r.referencia
        END,
        'motivo', 'valor incompatible con la ficha; el original queda en antes para corregirlo'
      )
    );
    UPDATE areas_verdes SET
      perimetro_m = CASE WHEN perimetro_m < 0 THEN NULL ELSE perimetro_m END,
      area_m2 = CASE WHEN area_m2 < 0 THEN NULL ELSE area_m2 END,
      referencia = CASE
        WHEN referencia IS NOT NULL AND char_length(referencia) > 500 THEN left(referencia, 500)
        ELSE referencia
      END,
      updated_at = now()
    WHERE id = r.id;
  END LOOP;

  FOR r IN
    SELECT id, codigo
    FROM areas_verdes a
    WHERE codigo IS NOT NULL
      AND btrim(codigo) <> ''
      AND id <> (
        SELECT min(id) FROM areas_verdes b
        WHERE b.codigo = a.codigo
          AND b.codigo IS NOT NULL
          AND btrim(b.codigo) <> ''
      )
    ORDER BY id
  LOOP
    nuevo := r.codigo || ' ·' || r.id::text;
    n := 2;
    WHILE EXISTS (SELECT 1 FROM areas_verdes WHERE codigo = nuevo AND id <> r.id) LOOP
      nuevo := r.codigo || ' ·' || r.id::text || '-' || n::text;
      n := n + 1;
    END LOOP;
    INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
    VALUES (
      'areas_verdes', r.id::text, 'edicion',
      jsonb_build_object('codigo', r.codigo),
      jsonb_build_object(
        'codigo', nuevo,
        'motivo', 'codigo duplicado; se conservo el de menor id. Corrija el codigo desde la ficha'
      )
    );
    UPDATE areas_verdes SET codigo = nuevo, updated_at = now() WHERE id = r.id;
  END LOOP;
END $$;

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
