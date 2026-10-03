-- Alta de actividad (RF-30, DEC-07). Columnas nulas: las filas ya cargadas no se reescriben.
-- nivel_riesgo no tiene CHECK. «Medio» entra después como ítem de catálogo, sin otra migración.

ALTER TABLE actividades ADD COLUMN IF NOT EXISTS subtipo TEXT;
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS codigo_externo TEXT;
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS unidad_solicitante TEXT;
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS nivel_riesgo TEXT;
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS fecha_programada DATE;
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS cantidad NUMERIC;

COMMENT ON COLUMN actividades.subtipo IS
  'Segundo nivel de la actividad, código del catálogo subtipo_actividad. Nulo si el alta no lo trae.';
COMMENT ON COLUMN actividades.codigo_externo IS
  'Código de Centuria u OSG si ya viene. El sistema no lo inventa.';
COMMENT ON COLUMN actividades.unidad_solicitante IS
  'Unidad que pide la actividad. Nulo si no se informa.';
COMMENT ON COLUMN actividades.nivel_riesgo IS
  'Código de catalogos clase nivel_riesgo. Sin CHECK: un valor nuevo se agrega al catálogo.';
COMMENT ON COLUMN actividades.fecha_programada IS
  'Fecha pedida para la actividad. No es la fecha de atención de la ficha.';
COMMENT ON COLUMN actividades.cantidad IS
  'Cantidad pedida. Nula si el alta no la trae.';
