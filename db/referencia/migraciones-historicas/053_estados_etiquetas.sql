-- Etiquetas del libro para la clase estado. Los códigos no se renombran.
-- Bloqueada no se borra: queda inactiva y provisional.
-- ejecutado y archivada se agregan. Idempotente: un segundo paso no duplica cambios.

DO $$
DECLARE
  r RECORD;
BEGIN
  FOR r IN
    SELECT c.id,
           c.codigo,
           c.nombre AS antes_nombre,
           c.activo AS antes_activo,
           c.provisional AS antes_prov,
           c.orden AS antes_orden,
           v.nuevo_nombre,
           v.nuevo_activo,
           v.nuevo_prov,
           v.nuevo_orden
    FROM catalogos c
    JOIN (VALUES
      ('sin_estado', 'Sin estado', TRUE, FALSE, 0),
      ('pendiente', 'Por iniciar', TRUE, FALSE, 1),
      ('en_proceso', 'En proceso', TRUE, FALSE, 2),
      ('cerrada', 'Cerrado', TRUE, FALSE, 4),
      ('cancelada', 'Cancelado', TRUE, FALSE, 5),
      ('bloqueada', 'Bloqueada', FALSE, TRUE, 90)
    ) AS v(codigo, nuevo_nombre, nuevo_activo, nuevo_prov, nuevo_orden)
      ON c.clase = 'estado' AND c.codigo = v.codigo
    WHERE c.nombre IS DISTINCT FROM v.nuevo_nombre
       OR c.activo IS DISTINCT FROM v.nuevo_activo
       OR c.provisional IS DISTINCT FROM v.nuevo_prov
       OR c.orden IS DISTINCT FROM v.nuevo_orden
    ORDER BY c.id
  LOOP
    INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
    VALUES (
      'catalogos',
      r.id::text,
      'edicion',
      jsonb_build_object(
        'clase', 'estado',
        'codigo', r.codigo,
        'nombre', r.antes_nombre,
        'activo', r.antes_activo,
        'provisional', r.antes_prov,
        'orden', r.antes_orden
      ),
      jsonb_build_object(
        'clase', 'estado',
        'codigo', r.codigo,
        'nombre', r.nuevo_nombre,
        'activo', r.nuevo_activo,
        'provisional', r.nuevo_prov,
        'orden', r.nuevo_orden
      )
    );

    UPDATE catalogos
    SET nombre = r.nuevo_nombre,
        activo = r.nuevo_activo,
        provisional = r.nuevo_prov,
        orden = r.nuevo_orden
    WHERE id = r.id;
  END LOOP;
END $$;

DO $$
DECLARE
  v RECORD;
  new_id BIGINT;
BEGIN
  FOR v IN
    SELECT *
    FROM (VALUES
      ('ejecutado', 'Ejecutado', 3),
      ('archivada', 'Archivado', 6)
    ) AS t(codigo, nombre, orden)
  LOOP
    IF NOT EXISTS (
      SELECT 1 FROM catalogos c WHERE c.clase = 'estado' AND c.codigo = v.codigo
    ) THEN
      INSERT INTO catalogos (clase, codigo, nombre, orden, activo, provisional)
      VALUES ('estado', v.codigo, v.nombre, v.orden, TRUE, FALSE)
      RETURNING id INTO new_id;

      INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
      VALUES (
        'catalogos',
        new_id::text,
        'alta',
        NULL,
        jsonb_build_object(
          'clase', 'estado',
          'codigo', v.codigo,
          'nombre', v.nombre,
          'activo', TRUE,
          'provisional', FALSE
        )
      );
    END IF;
  END LOOP;
END $$;
