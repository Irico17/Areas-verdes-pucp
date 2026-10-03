-- Clave inicial entregada por la jefatura. La persona la cambia antes de entrar al mapa.
-- Aditiva e idempotente. Las cuentas ya cargadas quedan en false y siguen entrando.

ALTER TABLE usuarios
  ADD COLUMN IF NOT EXISTS debe_cambiar_password BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN usuarios.debe_cambiar_password IS
  'True cuando la jefatura entregó una clave inicial y la persona todavía no la cambió.';
