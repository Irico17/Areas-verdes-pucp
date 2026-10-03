-- Lotes de importación reversibles (frente 2C).
-- Completa cambios (007): lote_id apunta al lote. No borra áreas ni labores.

CREATE TABLE IF NOT EXISTS lotes_importacion (
  id           BIGSERIAL PRIMARY KEY,
  entidad      TEXT NOT NULL,
  estado       TEXT NOT NULL DEFAULT 'confirmado',
  usuario_id   BIGINT NOT NULL REFERENCES usuarios (id),
  filas        INTEGER NOT NULL DEFAULT 0,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  revertido_en TIMESTAMPTZ,
  CONSTRAINT lotes_estado_chk CHECK (estado IN ('confirmado', 'revertido'))
);

COMMENT ON TABLE lotes_importacion IS
  'Lote de importación. La reversión aplica el antes de cambios y no pisa una edición posterior.';

ALTER TABLE cambios DROP CONSTRAINT IF EXISTS cambios_lote_fkey;
ALTER TABLE cambios
  ADD CONSTRAINT cambios_lote_fkey
  FOREIGN KEY (lote_id) REFERENCES lotes_importacion (id);

ALTER TABLE actividad_eventos ADD COLUMN IF NOT EXISTS usuario_id BIGINT REFERENCES usuarios (id);

COMMENT ON COLUMN actividad_eventos.usuario_id IS
  'Usuario de sesión que registró el evento. El timeline muestra esta cuenta, no actor_rol.';

CREATE INDEX IF NOT EXISTS actividad_eventos_usuario_idx
  ON actividad_eventos (usuario_id) WHERE usuario_id IS NOT NULL;
