-- Una labor puede registrarse sin GPS si hay lugar o zona de supervisión.

ALTER TABLE actividades ADD COLUMN IF NOT EXISTS lugar_id BIGINT REFERENCES lugares (id);
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS zona_supervision_id BIGINT REFERENCES zonas_supervision (id);

ALTER TABLE actividades ALTER COLUMN geom DROP NOT NULL;

ALTER TABLE actividades DROP CONSTRAINT IF EXISTS actividades_ubicacion_chk;
ALTER TABLE actividades ADD CONSTRAINT actividades_ubicacion_chk CHECK (
  geom IS NOT NULL
  OR lugar_id IS NOT NULL
  OR zona_supervision_id IS NOT NULL
);

CREATE INDEX IF NOT EXISTS actividades_lugar_idx ON actividades (lugar_id);
CREATE INDEX IF NOT EXISTS actividades_zona_sup_idx ON actividades (zona_supervision_id);

COMMENT ON COLUMN actividades.geom IS
  'Point EPSG:4326. NULL solo si hay lugar_id o zona_supervision_id.';
