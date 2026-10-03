-- RF-11. La fuente y el estado de la solicitud se validan contra catálogo.
-- solicitudes_fuente_chk y solicitudes_estado_chk se dejan: el catálogo de estado
-- de la labor no usa los mismos códigos (pendiente, cerrada, cancelada).
-- estado_solicitud repite los códigos del CHECK. No se borra ningún ítem.

DO $$
DECLARE
  v RECORD;
  new_id BIGINT;
BEGIN
  FOR v IN
    SELECT *
    FROM (VALUES
      ('estado_solicitud', 'por_iniciar', 'Por iniciar', 1),
      ('estado_solicitud', 'en_proceso', 'En proceso', 2),
      ('estado_solicitud', 'ejecutado', 'Ejecutado', 3),
      ('estado_solicitud', 'cerrado', 'Cerrado', 4),
      ('estado_solicitud', 'cancelado', 'Cancelado', 5)
    ) AS t(clase, codigo, nombre, orden)
  LOOP
    IF NOT EXISTS (
      SELECT 1 FROM catalogos c WHERE c.clase = v.clase AND c.codigo = v.codigo
    ) THEN
      INSERT INTO catalogos (clase, codigo, nombre, orden, activo, provisional)
      VALUES (v.clase, v.codigo, v.nombre, v.orden, TRUE, FALSE)
      RETURNING id INTO new_id;

      INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
      VALUES (
        'catalogos',
        new_id::text,
        'alta',
        NULL,
        jsonb_build_object(
          'clase', v.clase,
          'codigo', v.codigo,
          'nombre', v.nombre,
          'activo', TRUE,
          'provisional', FALSE
        )
      );
    END IF;
  END LOOP;
END $$;

COMMENT ON CONSTRAINT solicitudes_fuente_chk ON solicitudes IS
  'Se deja hasta que el catálogo fuente tenga exactamente estos códigos. La API ya valida contra catalogos (clase fuente).';
COMMENT ON CONSTRAINT solicitudes_estado_chk ON solicitudes IS
  'Se deja hasta que el catálogo tenga exactamente estos códigos. La API valida contra catalogos (clase estado_solicitud).';
