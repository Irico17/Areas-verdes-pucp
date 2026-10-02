-- La evidencia puede colgar de un evento de la labor (RF-19, RF-31).
-- La columna es nula: las filas ya cargadas siguen válidas y la API anterior no la lee.

ALTER TABLE evidencias
  ADD COLUMN IF NOT EXISTS evento_id BIGINT;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'evidencias_evento_id_fkey'
  ) THEN
    ALTER TABLE evidencias
      ADD CONSTRAINT evidencias_evento_id_fkey
      FOREIGN KEY (evento_id) REFERENCES actividad_eventos (id) NOT VALID;
    ALTER TABLE evidencias VALIDATE CONSTRAINT evidencias_evento_id_fkey;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS evidencias_evento_id_idx ON evidencias (evento_id);

COMMENT ON COLUMN evidencias.evento_id IS
  'Evento de la labor al que pertenece la evidencia. Nulo en filas anteriores a esta migración.';
