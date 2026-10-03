-- Nombres visibles de las cuadrillas de demostración y del tipo incidencia.
-- Idempotente: solo escribe si el texto sigue siendo el anterior.
-- No borra filas, no trunca y no cambia códigos.

DO $$
DECLARE
  r RECORD;
BEGIN
  FOR r IN
    SELECT c.id AS entidad_id, c.equipo AS antes, v.nuevo AS despues
    FROM capataces c
    JOIN (VALUES
      ('Equipo Norte', 'Cuadrilla Norte'),
      ('Equipo Sur', 'Cuadrilla Sur'),
      ('Equipo Riego', 'Cuadrilla Riego')
    ) AS v(antes, nuevo) ON c.equipo = v.antes
  LOOP
    INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
    VALUES (
      'capataces',
      r.entidad_id,
      'edicion',
      jsonb_build_object('equipo', r.antes),
      jsonb_build_object('equipo', r.despues)
    );
    UPDATE capataces SET equipo = r.despues WHERE id = r.entidad_id;
  END LOOP;

  FOR r IN
    SELECT c.id AS entidad_id, c.nombre_ficticio AS antes, v.nuevo AS despues
    FROM cuadrillas c
    JOIN (VALUES
      ('Equipo Norte', 'Cuadrilla Norte'),
      ('Equipo Sur', 'Cuadrilla Sur'),
      ('Equipo Riego', 'Cuadrilla Riego')
    ) AS v(antes, nuevo) ON c.nombre_ficticio = v.antes
  LOOP
    INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
    VALUES (
      'cuadrillas',
      r.entidad_id,
      'edicion',
      jsonb_build_object('nombre_ficticio', r.antes),
      jsonb_build_object('nombre_ficticio', r.despues)
    );
    UPDATE cuadrillas SET nombre_ficticio = r.despues WHERE id = r.entidad_id;
  END LOOP;

  FOR r IN
    SELECT u.id::text AS entidad_id, u.nombre AS antes, v.nuevo AS despues
    FROM usuarios u
    JOIN (VALUES
      ('norte', 'Equipo Norte', 'Elsa Quispe'),
      ('sur', 'Equipo Sur', 'Iván Paredes'),
      ('riego', 'Equipo Riego', 'Nora Beltrán')
    ) AS v(usuario, antes, nuevo) ON u.usuario = v.usuario AND u.nombre = v.antes
  LOOP
    INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
    VALUES (
      'usuarios',
      r.entidad_id,
      'edicion',
      jsonb_build_object('nombre', r.antes),
      jsonb_build_object('nombre', r.despues)
    );
    UPDATE usuarios SET nombre = r.despues WHERE id = r.entidad_id::bigint;
  END LOOP;

  FOR r IN
    SELECT c.id::text AS entidad_id, c.nombre AS antes
    FROM catalogos c
    WHERE c.clase = 'tipo_actividad'
      AND c.codigo = 'incidencia'
      AND c.nombre = 'Incidencia'
  LOOP
    INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
    VALUES (
      'catalogos',
      r.entidad_id,
      'edicion',
      jsonb_build_object('clase', 'tipo_actividad', 'codigo', 'incidencia', 'nombre', r.antes),
      jsonb_build_object('clase', 'tipo_actividad', 'codigo', 'incidencia', 'nombre', 'Novedad de campo')
    );
    UPDATE catalogos SET nombre = 'Novedad de campo' WHERE id = r.entidad_id::bigint;
  END LOOP;
END $$;
