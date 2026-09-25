-- Frente 2A. El riego guarda la zona por FK. No calcula cobertura oficial.

CREATE TABLE IF NOT EXISTS zonas_supervision (
  id      TEXT PRIMARY KEY,
  codigo  TEXT NOT NULL UNIQUE,
  nombre  TEXT NOT NULL
);

COMMENT ON TABLE zonas_supervision IS
  'Z1–Z4. Si el frente de catastro ya creó la tabla, esta sentencia no la pisa.';

INSERT INTO zonas_supervision (id, codigo, nombre) VALUES
  ('Z1', 'Z1', 'Zona 1'),
  ('Z2', 'Z2', 'Zona 2'),
  ('Z3', 'Z3', 'Zona 3'),
  ('Z4', 'Z4', 'Zona 4')
ON CONFLICT (id) DO NOTHING;

ALTER TABLE riego_registros ADD COLUMN IF NOT EXISTS zona_supervision_id TEXT REFERENCES zonas_supervision (id);
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

ALTER TABLE actividades DROP CONSTRAINT IF EXISTS actividades_zona_fk;
ALTER TABLE actividades ADD CONSTRAINT actividades_zona_fk
  FOREIGN KEY (zona_supervision_id) REFERENCES zonas_supervision (id);

ALTER TABLE actividades DROP CONSTRAINT IF EXISTS actividades_cuadrilla_fk;
ALTER TABLE actividades ADD CONSTRAINT actividades_cuadrilla_fk
  FOREIGN KEY (cuadrilla_id) REFERENCES cuadrillas (id);
