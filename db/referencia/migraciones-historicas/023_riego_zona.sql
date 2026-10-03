-- Frente 2A. El riego guarda la zona por FK a la tabla de 010.
-- No recrea zonas_supervision: el frente 1A ya la definió con id numérico.

ALTER TABLE riego_registros ADD COLUMN IF NOT EXISTS zona_supervision_id BIGINT REFERENCES zonas_supervision (id);
ALTER TABLE riego_registros ADD COLUMN IF NOT EXISTS ciclo TEXT NOT NULL DEFAULT '';
ALTER TABLE riego_registros ADD COLUMN IF NOT EXISTS superficie_m2 NUMERIC;

ALTER TABLE riego_registros DROP CONSTRAINT IF EXISTS riego_superficie_chk;
ALTER TABLE riego_registros ADD CONSTRAINT riego_superficie_chk CHECK (
  superficie_m2 IS NULL OR superficie_m2 >= 0
);

CREATE UNIQUE INDEX IF NOT EXISTS riego_zona_turno_fecha_uidx
  ON riego_registros (zona_supervision_id, turno, fecha)
  WHERE zona_supervision_id IS NOT NULL;

COMMENT ON COLUMN riego_registros.zona_supervision_id IS
  'Sector de supervisión. El alta nueva exige esta FK. El indicador oficial queda pendiente.';

ALTER TABLE actividades DROP CONSTRAINT IF EXISTS actividades_cuadrilla_fk;
ALTER TABLE actividades ADD CONSTRAINT actividades_cuadrilla_fk
  FOREIGN KEY (cuadrilla_id) REFERENCES cuadrillas (id);
