-- El mismo nombre ficticio no se repite en una actividad. No borra filas.

CREATE INDEX IF NOT EXISTS personal_labor_actividad_idx
  ON personal_labor (actividad_id);

CREATE UNIQUE INDEX IF NOT EXISTS personal_labor_nombre_uidx
  ON personal_labor (actividad_id, nombre_ficticio);
