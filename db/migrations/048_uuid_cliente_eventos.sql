-- Idempotencia offline de eventos (RF-12).
-- Varios NULL siguen permitidos: el índice único es parcial.

ALTER TABLE actividad_eventos
  ADD COLUMN IF NOT EXISTS uuid_cliente UUID;

CREATE UNIQUE INDEX IF NOT EXISTS actividad_eventos_uuid_cliente_uidx
  ON actividad_eventos (uuid_cliente)
  WHERE uuid_cliente IS NOT NULL;

COMMENT ON COLUMN actividad_eventos.uuid_cliente IS
  'Identificador generado por el cliente para reintentos offline. Nulo en eventos anteriores.';
