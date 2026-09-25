-- Upsert de áreas verdes desde etl-lote. No trunca. origen_ref único para ON CONFLICT.

CREATE UNIQUE INDEX IF NOT EXISTS areas_verdes_origen_ref_uidx
  ON areas_verdes (origen_ref);

COMMENT ON COLUMN areas_verdes.origen_ref IS
  'Identificador de la feature de catastro (AV-NNNN). etl-lote hace upsert por esta clave, sin TRUNCATE.';
