-- Frente 3A. Vista previa del importador. La confirmación pasa el lote a confirmado
-- para que la reversión de 041 pueda deshacerlo.

ALTER TABLE lotes_importacion ADD COLUMN IF NOT EXISTS contenido BYTEA;
ALTER TABLE lotes_importacion ADD COLUMN IF NOT EXISTS nombre_archivo TEXT NOT NULL DEFAULT '';

ALTER TABLE lotes_importacion DROP CONSTRAINT IF EXISTS lotes_estado_chk;
ALTER TABLE lotes_importacion ADD CONSTRAINT lotes_estado_chk
  CHECK (estado IN ('vista_previa', 'confirmado', 'revertido'));

COMMENT ON COLUMN lotes_importacion.contenido IS
  'Archivo de la vista previa. Confirmar lo vuelve a leer. Revertir no lo necesita.';
