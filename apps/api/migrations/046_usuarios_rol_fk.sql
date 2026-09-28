-- Migración 046: Reemplaza el CHECK usuarios_rol_chk con una FK hacia roles(codigo).
-- Permite que los roles sean configurables (RNF-05, RF-02).
-- Migración aditiva e idempotente sin cambios de datos.

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'usuarios_rol_fkey'
  ) THEN
    ALTER TABLE usuarios
      ADD CONSTRAINT usuarios_rol_fkey
      FOREIGN KEY (rol) REFERENCES roles (codigo) NOT VALID;

    ALTER TABLE usuarios
      VALIDATE CONSTRAINT usuarios_rol_fkey;
  END IF;
END $$;

ALTER TABLE usuarios DROP CONSTRAINT IF EXISTS usuarios_rol_chk;
