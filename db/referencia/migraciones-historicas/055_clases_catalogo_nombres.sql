-- Clases que la API aún no aceptaba, como catálogo de nombres.
-- Los ítems de ejemplo son ficticios o dicen que el archivo no llegó.
-- Los que el cliente no confirmó quedan provisional = true. No se borra nada.

DO $$
DECLARE
  v RECORD;
  new_id BIGINT;
BEGIN
  FOR v IN
    SELECT *
    FROM (VALUES
      ('plaga', 'demo_hoja', 'Plaga de demostración', 1, TRUE),
      ('producto_fitosanitario', 'demo_producto', 'Producto fitosanitario de demostración', 1, TRUE),
      ('frecuencia', 'semanal', 'Semanal', 1, FALSE),
      ('frecuencia', 'quincenal', 'Quincenal', 2, FALSE),
      ('sede', 'sede_demo', 'Sede de demostración', 1, TRUE),
      ('cuartel', 'sin_archivo', 'Sin archivo de cuarteles', 1, TRUE),
      ('sector_capataz', 'sector_norte', 'Sector norte (ficticio)', 1, FALSE),
      ('sector_capataz', 'campo_deportivo', 'Campo deportivo', 2, FALSE),
      ('sector_capataz', 'bosque_humedo', 'Bosque húmedo', 3, FALSE)
    ) AS t(clase, codigo, nombre, orden, provisional)
  LOOP
    IF NOT EXISTS (
      SELECT 1 FROM catalogos c WHERE c.clase = v.clase AND c.codigo = v.codigo
    ) THEN
      INSERT INTO catalogos (clase, codigo, nombre, orden, activo, provisional)
      VALUES (v.clase, v.codigo, v.nombre, v.orden, TRUE, v.provisional)
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
          'provisional', v.provisional
        )
      );
    END IF;
  END LOOP;
END $$;
