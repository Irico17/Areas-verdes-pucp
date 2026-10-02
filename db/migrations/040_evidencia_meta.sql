-- Metadatos de evidencia de campo: hash, punto y EXIF básico.
-- El binario sigue en disco o S3 (internal/blobs). Postgres guarda la clave.

ALTER TABLE evidencias ADD COLUMN IF NOT EXISTS sha256 TEXT;
ALTER TABLE evidencias ADD COLUMN IF NOT EXISTS lat DOUBLE PRECISION;
ALTER TABLE evidencias ADD COLUMN IF NOT EXISTS lon DOUBLE PRECISION;
ALTER TABLE evidencias ADD COLUMN IF NOT EXISTS exif JSONB;

COMMENT ON COLUMN evidencias.sha256 IS
  'SHA-256 hexadecimal del archivo. El mismo id con otro hash responde 409.';
COMMENT ON COLUMN evidencias.lat IS
  'Latitud opcional (EXIF o permiso del teléfono). NULL si no hay punto.';
COMMENT ON COLUMN evidencias.lon IS
  'Longitud opcional, en el mismo sistema que lat.';
COMMENT ON COLUMN evidencias.exif IS
  'Solo fecha, lat, lon y orientacion. El resto se descarta.';

ALTER TABLE actividad_eventos ADD COLUMN IF NOT EXISTS usuario_id BIGINT REFERENCES usuarios (id);

COMMENT ON COLUMN actividad_eventos.usuario_id IS
  'Usuario de la sesión que registró el evento. NULL en filas anteriores a esta columna.';
