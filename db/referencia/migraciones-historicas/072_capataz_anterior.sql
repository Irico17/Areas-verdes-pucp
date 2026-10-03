-- RF-31-CA2. La reasignación guarda la cuadrilla anterior y la nueva.
-- La columna nace nula: los eventos ya escritos siguen válidos.

ALTER TABLE actividad_eventos
  ADD COLUMN IF NOT EXISTS capataz_anterior TEXT;

COMMENT ON COLUMN actividad_eventos.capataz_anterior IS
  'Cuadrilla responsable antes de una reasignación. La nueva queda en capataz_id. Nulo en el resto de hitos.';

CREATE INDEX IF NOT EXISTS actividad_eventos_capataz_anterior_idx
  ON actividad_eventos (capataz_anterior)
  WHERE capataz_anterior IS NOT NULL;
