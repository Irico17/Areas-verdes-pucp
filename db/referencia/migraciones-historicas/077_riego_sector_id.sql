-- RF-26. El riego apunta al sector de capataz del catálogo.
-- El texto ya guardado en riego_registros.sector no se borra ni se reescribe.

ALTER TABLE riego_registros ADD COLUMN IF NOT EXISTS sector_id BIGINT;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'riego_registros_sector_id_fkey'
  ) THEN
    ALTER TABLE riego_registros
      ADD CONSTRAINT riego_registros_sector_id_fkey
      FOREIGN KEY (sector_id) REFERENCES sectores_capataz (id);
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS riego_registros_sector_id_idx
  ON riego_registros (sector_id);

COMMENT ON COLUMN riego_registros.sector_id IS
  'Sector de capataz del catálogo. Nulo en registros anteriores, que conservan solo el texto de sector.';
