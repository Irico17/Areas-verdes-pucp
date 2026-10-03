-- 016 creó actividades.lugar_id y zona_supervision_id como BIGINT.
-- Una 020 antigua las declaraba TEXT; ese archivo ya no lo hace, y no se reescribe.
-- Esta migración alinea un esquema que hubiera quedado en TEXT.

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'actividades'
      AND column_name = 'lugar_id'
      AND udt_name <> 'int8'
  ) THEN
    ALTER TABLE actividades
      ALTER COLUMN lugar_id TYPE BIGINT
      USING NULLIF(btrim(lugar_id::text), '')::bigint;
  END IF;

  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'actividades'
      AND column_name = 'zona_supervision_id'
      AND udt_name <> 'int8'
  ) THEN
    ALTER TABLE actividades
      ALTER COLUMN zona_supervision_id TYPE BIGINT
      USING NULLIF(btrim(zona_supervision_id::text), '')::bigint;
  END IF;
END $$;

COMMENT ON COLUMN actividades.lugar_id IS
  'BIGINT con FK a lugares (016). No es TEXT.';
COMMENT ON COLUMN actividades.zona_supervision_id IS
  'BIGINT con FK a zonas_supervision (016). No es TEXT.';
