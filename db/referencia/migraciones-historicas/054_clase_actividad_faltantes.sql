-- Las dos clases de actividad que el libro pide y la base no tenía.
-- No toca las siete ya cargadas en 020.

DO $$
DECLARE
  v RECORD;
  new_id BIGINT;
BEGIN
  FOR v IN
    SELECT *
    FROM (VALUES
      ('fitosanitario', 'Manejo fitosanitario', 8),
      ('inspeccion_monitoreo', 'Inspección y monitoreo', 9)
    ) AS t(codigo, nombre, orden)
  LOOP
    IF NOT EXISTS (
      SELECT 1 FROM catalogos c
      WHERE c.clase = 'clase_actividad' AND c.codigo = v.codigo
    ) THEN
      INSERT INTO catalogos (clase, codigo, nombre, orden, activo, provisional)
      VALUES ('clase_actividad', v.codigo, v.nombre, v.orden, TRUE, FALSE)
      RETURNING id INTO new_id;

      INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
      VALUES (
        'catalogos',
        new_id::text,
        'alta',
        NULL,
        jsonb_build_object(
          'clase', 'clase_actividad',
          'codigo', v.codigo,
          'nombre', v.nombre,
          'activo', TRUE,
          'provisional', FALSE
        )
      );
    END IF;
  END LOOP;
END $$;
