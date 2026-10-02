-- RF-13. Empresa y frecuencia pasan a ser una fila de catálogo.
-- Las columnas de texto se conservan: lo ya escrito se sigue mostrando.
-- No se crea otra columna de texto. Los nombres de empresa son ficticios.
-- «Servicios Forestales del Perú» es el nombre del cuerpo de paridad ya versionado.

ALTER TABLE ordenes_servicio ADD COLUMN IF NOT EXISTS empresa_id BIGINT;
ALTER TABLE ordenes_servicio ADD COLUMN IF NOT EXISTS frecuencia_id BIGINT;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'ordenes_empresa_id_fkey'
  ) THEN
    ALTER TABLE ordenes_servicio
      ADD CONSTRAINT ordenes_empresa_id_fkey
      FOREIGN KEY (empresa_id) REFERENCES catalogos (id);
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'ordenes_frecuencia_id_fkey'
  ) THEN
    ALTER TABLE ordenes_servicio
      ADD CONSTRAINT ordenes_frecuencia_id_fkey
      FOREIGN KEY (frecuencia_id) REFERENCES catalogos (id);
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS ordenes_empresa_id_idx ON ordenes_servicio (empresa_id);
CREATE INDEX IF NOT EXISTS ordenes_frecuencia_id_idx ON ordenes_servicio (frecuencia_id);

DO $$
DECLARE
  v RECORD;
  new_id BIGINT;
BEGIN
  FOR v IN
    SELECT *
    FROM (VALUES
      ('empresa', 'taller_verde_andino', 'Taller Verde Andino (ficticia)', 1, FALSE),
      ('empresa', 'jardines_del_rimac', 'Jardines del Rímac (ficticia)', 2, FALSE),
      ('empresa', 'poda_litoral', 'Poda Litoral (ficticia)', 3, FALSE),
      ('empresa', 'servicios_forestales_demo', 'Servicios Forestales del Perú', 4, FALSE),
      ('frecuencia', 'mensual', 'Mensual', 3, FALSE),
      ('frecuencia', 'unica', 'Única', 4, FALSE)
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

COMMENT ON COLUMN ordenes_servicio.empresa IS
  'Texto ya cargado o nombre del catálogo al crear. No es la fuente de las altas nuevas: esa es empresa_id.';
COMMENT ON COLUMN ordenes_servicio.empresa_id IS
  'FK a catalogos clase empresa. Nula en filas anteriores al catálogo.';
COMMENT ON COLUMN ordenes_servicio.frecuencia IS
  'Texto ya cargado o nombre del catálogo al crear. No es la fuente de las altas nuevas: esa es frecuencia_id.';
COMMENT ON COLUMN ordenes_servicio.frecuencia_id IS
  'FK a catalogos clase frecuencia. Nula en filas anteriores al catálogo.';
COMMENT ON COLUMN ordenes_servicio.conformidad IS
  'Conformidad de la orden. Se sigue guardando. No cierra la solicitud ni adjunta el reporte del proveedor.';
