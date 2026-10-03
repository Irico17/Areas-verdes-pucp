-- Etiquetas visibles del backlog v2 (§4.4). Solo cambia roles.nombre.
-- Idempotente: no inserta, no borra y no toca el código del rol.

UPDATE roles
SET nombre = 'Ingeniería / Coordinación'
WHERE codigo = 'coordinacion'
  AND nombre IS DISTINCT FROM 'Ingeniería / Coordinación';

UPDATE roles
SET nombre = 'Jefatura de sección'
WHERE codigo = 'jefatura'
  AND nombre IS DISTINCT FROM 'Jefatura de sección';

UPDATE roles
SET nombre = 'Capataz'
WHERE codigo = 'capataz'
  AND nombre IS DISTINCT FROM 'Capataz';

UPDATE roles
SET nombre = 'Administrador del sistema'
WHERE codigo = 'admin'
  AND nombre IS DISTINCT FROM 'Administrador del sistema';
