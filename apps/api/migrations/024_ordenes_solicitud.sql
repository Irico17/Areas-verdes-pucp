-- Frente 2A. Orden tercerizada con periodo y reporte. No cierra la solicitud.

ALTER TABLE ordenes_servicio ADD COLUMN IF NOT EXISTS periodo_inicio DATE;
ALTER TABLE ordenes_servicio ADD COLUMN IF NOT EXISTS periodo_fin DATE;
ALTER TABLE ordenes_servicio ADD COLUMN IF NOT EXISTS reporte_proveedor TEXT NOT NULL DEFAULT '';

ALTER TABLE ordenes_servicio DROP CONSTRAINT IF EXISTS ordenes_periodo_chk;
ALTER TABLE ordenes_servicio ADD CONSTRAINT ordenes_periodo_chk CHECK (
  periodo_inicio IS NULL OR periodo_fin IS NULL OR periodo_fin >= periodo_inicio
);

COMMENT ON TABLE ordenes_servicio IS
  'La orden no cierra la labor ni la solicitud. El cierre de una labor tercerizada exige que la orden exista, y además una ejecución registrada.';

ALTER TABLE solicitudes ADD COLUMN IF NOT EXISTS origen_ref TEXT;
ALTER TABLE solicitudes ADD COLUMN IF NOT EXISTS archivada_en TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS solicitudes_origen_ref_uidx
  ON solicitudes (origen_ref) WHERE origen_ref IS NOT NULL AND origen_ref <> '';

COMMENT ON COLUMN solicitudes.codigo_externo IS
  'Código de Centuria u OSG si la fuente lo trae. El sistema no lo inventa ni llama a Centuria.';
