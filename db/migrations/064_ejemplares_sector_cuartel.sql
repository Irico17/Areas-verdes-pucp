-- Sector de capataz o cuartel del ejemplar. La columna nace vacía:
-- no se rellenan las filas ya importadas con un valor inventado.

ALTER TABLE ejemplares
  ADD COLUMN IF NOT EXISTS sector_cuartel_id BIGINT;

COMMENT ON COLUMN ejemplares.sector_cuartel_id IS
  'Catálogo de sector de capataz o de cuartel. NULL hasta que alguien lo asigne. No es texto libre.';

CREATE INDEX IF NOT EXISTS ejemplares_sector_cuartel_idx
  ON ejemplares (sector_cuartel_id)
  WHERE sector_cuartel_id IS NOT NULL;
