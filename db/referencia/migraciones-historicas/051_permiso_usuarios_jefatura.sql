-- La jefatura administra cuentas. El administrador del sistema conserva el acceso que ya tenía.
-- No borra filas. Si el permiso ya está, no lo reescribe.

INSERT INTO permisos (rol, accion)
VALUES ('jefatura', 'usuarios')
ON CONFLICT (rol, accion) DO NOTHING;

INSERT INTO permisos (rol, accion)
VALUES ('admin', 'usuarios')
ON CONFLICT (rol, accion) DO NOTHING;
