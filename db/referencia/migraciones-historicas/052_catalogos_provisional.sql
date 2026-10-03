-- Marca de catálogo aún no confirmada por el cliente.
-- Aditiva: no borra filas ni reescribe ítems.

ALTER TABLE catalogos ADD COLUMN IF NOT EXISTS provisional BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN catalogos.provisional IS
  'El cliente todavía no cierra este valor. Se muestra, no se trata como decisión oficial.';
