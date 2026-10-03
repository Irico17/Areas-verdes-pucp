-- RF-11. Cantidad pedida y cantidad ejecutada, y ubicación por lugar o por punto.
-- cantidad no se borra: se copia a cantidad_solicitada cuando esa columna aún está vacía.
-- El CHECK de fuente y de estado se deja. Lo retira otra migración cuando el catálogo
-- tenga exactamente los mismos códigos.

ALTER TABLE solicitudes ADD COLUMN IF NOT EXISTS cantidad_solicitada INTEGER;
ALTER TABLE solicitudes ADD COLUMN IF NOT EXISTS cantidad_ejecutada INTEGER;
ALTER TABLE solicitudes ADD COLUMN IF NOT EXISTS lugar_id BIGINT;
ALTER TABLE solicitudes ADD COLUMN IF NOT EXISTS lat DOUBLE PRECISION;
ALTER TABLE solicitudes ADD COLUMN IF NOT EXISTS lon DOUBLE PRECISION;

UPDATE solicitudes
SET cantidad_solicitada = cantidad
WHERE cantidad IS NOT NULL
  AND cantidad_solicitada IS NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'solicitudes_lugar_id_fkey'
  ) THEN
    ALTER TABLE solicitudes
      ADD CONSTRAINT solicitudes_lugar_id_fkey
      FOREIGN KEY (lugar_id) REFERENCES lugares (id);
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'solicitudes_punto_par_chk'
  ) THEN
    ALTER TABLE solicitudes
      ADD CONSTRAINT solicitudes_punto_par_chk
      CHECK ((lat IS NULL AND lon IS NULL) OR (lat IS NOT NULL AND lon IS NOT NULL));
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS solicitudes_lugar_id_idx ON solicitudes (lugar_id);

COMMENT ON COLUMN solicitudes.cantidad IS
  'Cantidad histórica. No se borra. cantidad_solicitada la copia si aún no tenía valor.';
COMMENT ON COLUMN solicitudes.cantidad_solicitada IS
  'Cantidad pedida. Puede diferir de cantidad_ejecutada.';
COMMENT ON COLUMN solicitudes.cantidad_ejecutada IS
  'Cantidad atendida. Nula si todavía no hay ejecución. Puede diferir de la pedida.';
COMMENT ON COLUMN solicitudes.lugar_id IS
  'Lugar del catálogo. Nulo si la ubicación es un punto o solo texto ya cargado.';
COMMENT ON COLUMN solicitudes.lat IS
  'Latitud del pin. Nula si no hay punto. Va junto con lon.';
COMMENT ON COLUMN solicitudes.lon IS
  'Longitud del pin. Nula si no hay punto. Va junto con lat.';
COMMENT ON COLUMN solicitudes.lugar IS
  'Texto de lugar ya cargado. Se conserva y se muestra. El alta nueva usa lugar_id o el punto.';
