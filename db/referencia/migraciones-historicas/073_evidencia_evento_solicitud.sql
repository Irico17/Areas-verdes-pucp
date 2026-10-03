-- RF-19-CA3 y RF-31-CA4. evento_id (047) y solicitud_id (005) ya existen.
-- Esta migración no las recrea ni borra evidencias: fija el comentario y el índice de lectura.

COMMENT ON COLUMN evidencias.evento_id IS
  'Hito de la cadena al que cuelga la evidencia. Nulo en archivos anteriores a este vínculo.';

COMMENT ON COLUMN evidencias.solicitud_id IS
  'Solicitud a la que puede colgar la evidencia, además de la actividad. Nulo si no aplica.';

CREATE INDEX IF NOT EXISTS evidencias_solicitud_id_idx
  ON evidencias (solicitud_id)
  WHERE solicitud_id IS NOT NULL;
