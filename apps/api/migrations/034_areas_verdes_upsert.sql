-- Upsert de áreas verdes desde etl-lote. No trunca. origen_ref único para ON CONFLICT.
-- Varios NULL conviven. Un origen_ref repetido y no nulo se parte: menor id intacto, resto con « ·<id>».

DO $$
DECLARE
  r RECORD;
  nuevo TEXT;
  sufijo INT;
BEGIN
  FOR r IN
    SELECT id, origen_ref
    FROM areas_verdes a
    WHERE origen_ref IS NOT NULL
      AND id <> (SELECT min(id) FROM areas_verdes b WHERE b.origen_ref = a.origen_ref)
    ORDER BY id
  LOOP
    sufijo := 0;
    LOOP
      IF sufijo = 0 THEN
        nuevo := r.origen_ref || ' ·' || r.id::text;
      ELSE
        nuevo := r.origen_ref || ' ·' || r.id::text || '-' || sufijo::text;
      END IF;
      EXIT WHEN NOT EXISTS (
        SELECT 1 FROM areas_verdes WHERE origen_ref = nuevo AND id <> r.id
      );
      sufijo := sufijo + 1;
    END LOOP;
    INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
    VALUES (
      'areas_verdes', r.id::text, 'edicion',
      jsonb_build_object('origen_ref', r.origen_ref),
      jsonb_build_object(
        'origen_ref', nuevo,
        'motivo', 'origen_ref duplicado; se conservo el de menor id. Corrija el identificador desde la ficha'
      )
    );
    UPDATE areas_verdes SET origen_ref = nuevo, updated_at = now() WHERE id = r.id;
  END LOOP;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS areas_verdes_origen_ref_uidx
  ON areas_verdes (origen_ref);

COMMENT ON COLUMN areas_verdes.origen_ref IS
  'Identificador de la feature de catastro (AV-NNNN). etl-lote hace upsert por esta clave, sin TRUNCATE.';
