-- Catálogo semilla de roles y FK de permisos.rol.
-- Los códigos son los que ya usa el CHECK de usuarios. No cambia lo que la API muestra.

CREATE TABLE IF NOT EXISTS roles (
  codigo TEXT PRIMARY KEY,
  nombre TEXT NOT NULL,
  orden  INTEGER NOT NULL DEFAULT 0
);

COMMENT ON TABLE roles IS
  'Catálogo semilla de roles. No es un editor de políticas ni el SSO.';

INSERT INTO roles (codigo, nombre, orden) VALUES
  ('capataz', 'Capataz', 1),
  ('coordinacion', 'Coordinación', 2),
  ('jefatura', 'Jefatura', 3),
  ('admin', 'Administración', 4)
ON CONFLICT (codigo) DO NOTHING;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'permisos_rol_fkey'
  ) THEN
    ALTER TABLE permisos
      ADD CONSTRAINT permisos_rol_fkey
      FOREIGN KEY (rol) REFERENCES roles (codigo);
  END IF;
END $$;
