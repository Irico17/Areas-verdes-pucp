-- Etiquetas visibles de roles v2 (Backlog v2) y ajuste de permisos de jefatura.
-- Actualización aditiva e idempotente. No borra filas.

ALTER TABLE roles ADD COLUMN IF NOT EXISTS descripcion TEXT;
ALTER TABLE roles ADD COLUMN IF NOT EXISTS activo BOOLEAN NOT NULL DEFAULT TRUE;

CREATE UNIQUE INDEX IF NOT EXISTS permisos_rol_accion_uidx ON permisos (rol, accion);

-- Actualizar nombres visibles de roles registrando antes y después en cambios
DO $$
DECLARE
  r RECORD;
BEGIN
  FOR r IN
    SELECT rol_fila.codigo, rol_fila.nombre AS antes_nombre, v.nuevo_nombre
    FROM roles rol_fila
    JOIN (VALUES
      ('capataz', 'Capataz'),
      ('coordinacion', 'Ingeniería/Coordinación'),
      ('jefatura', 'Jefatura / Jefe de sección'),
      ('admin', 'Administrador del sistema')
    ) AS v(codigo, nuevo_nombre) ON rol_fila.codigo = v.codigo
    WHERE rol_fila.nombre IS DISTINCT FROM v.nuevo_nombre
    ORDER BY rol_fila.codigo
  LOOP
    INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
    VALUES (
      'roles',
      r.codigo,
      'edicion',
      jsonb_build_object('nombre', r.antes_nombre),
      jsonb_build_object('nombre', r.nuevo_nombre)
    );

    UPDATE roles
    SET nombre = r.nuevo_nombre
    WHERE codigo = r.codigo;
  END LOOP;
END $$;

-- Permiso para que jefatura pueda registrar evidencias fotográficas
INSERT INTO permisos (rol, accion)
VALUES ('jefatura', 'evidencias')
ON CONFLICT (rol, accion) DO NOTHING;
