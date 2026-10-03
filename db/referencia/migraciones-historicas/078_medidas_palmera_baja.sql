-- Baja lógica de la medida de palmera (RNF-14, DEC-05).
-- La fila ya cargada no se borra ni se reescribe. baja_en nace nulo.

ALTER TABLE medidas_palmera
  ADD COLUMN IF NOT EXISTS baja_en TIMESTAMPTZ;

COMMENT ON COLUMN medidas_palmera.baja_en IS
  'Baja lógica al revertir el lote que cargó la medida. La fila permanece. No hay plazo de retención: no se borra por antigüedad.';
