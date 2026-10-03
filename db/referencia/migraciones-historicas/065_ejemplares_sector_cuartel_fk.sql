-- La columna apunta al catálogo. No borra filas ni reescribe ejemplares.

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'ejemplares_sector_cuartel_fk'
      AND conrelid = 'ejemplares'::regclass
  ) THEN
    ALTER TABLE ejemplares
      ADD CONSTRAINT ejemplares_sector_cuartel_fk
      FOREIGN KEY (sector_cuartel_id) REFERENCES catalogos (id);
  END IF;
END $$;
