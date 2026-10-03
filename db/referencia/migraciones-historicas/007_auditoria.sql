-- Auditoría de cambios (ola 0). lote_id queda libre hasta la migración de lotes.
-- No borra filas ni recarga tablas.

CREATE TABLE IF NOT EXISTS cambios (
  id          BIGSERIAL PRIMARY KEY,
  entidad     TEXT NOT NULL,
  entidad_id  TEXT NOT NULL,
  accion      TEXT NOT NULL,
  antes       JSONB,
  despues     JSONB,
  usuario_id  BIGINT REFERENCES usuarios (id),
  lote_id     BIGINT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT cambios_accion_chk CHECK (
    accion IN ('alta', 'edicion', 'baja', 'importacion', 'reversion')
  )
);

COMMENT ON TABLE cambios IS
  'Antes y después de una edición. El actor es el usuario de sesión cuando exista.';

CREATE INDEX IF NOT EXISTS cambios_entidad_idx ON cambios (entidad, entidad_id, id);
CREATE INDEX IF NOT EXISTS cambios_lote_idx ON cambios (lote_id) WHERE lote_id IS NOT NULL;
