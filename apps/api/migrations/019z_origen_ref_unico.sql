-- Upsert del frente 1E. origen_ref ya existía en 010-015, sin índice único.
-- PostgreSQL admite varios NULL; las filas del lote siempre traen origen_ref.

CREATE UNIQUE INDEX IF NOT EXISTS zonas_supervision_origen_ref_uidx
  ON zonas_supervision (origen_ref);

CREATE UNIQUE INDEX IF NOT EXISTS poligonos_cuadrilla_origen_ref_uidx
  ON poligonos_cuadrilla (origen_ref);

CREATE UNIQUE INDEX IF NOT EXISTS lugares_origen_ref_uidx
  ON lugares (origen_ref);

CREATE UNIQUE INDEX IF NOT EXISTS ejemplares_origen_ref_uidx
  ON ejemplares (origen_ref);
