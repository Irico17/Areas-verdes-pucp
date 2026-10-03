-- Frente 2A. Poda como incidencia propia, enlazada a solicitud si hay código externo.

CREATE TABLE IF NOT EXISTS podas (
  id                   UUID PRIMARY KEY,
  codigo               TEXT NOT NULL UNIQUE,
  codigo_externo       TEXT,
  tipo                 TEXT NOT NULL DEFAULT '',
  tipo_actividad       TEXT NOT NULL DEFAULT '',
  fecha_reporte        DATE,
  fecha_ejecucion      DATE,
  personal_ficticio    TEXT NOT NULL DEFAULT '',
  ubicacion            TEXT NOT NULL DEFAULT '',
  lugar_id             TEXT,
  unidad               TEXT NOT NULL DEFAULT '',
  cantidad_pedida      NUMERIC NOT NULL DEFAULT 0,
  cantidad_ejecutada   NUMERIC NOT NULL DEFAULT 0,
  tipo_vegetacion      TEXT NOT NULL DEFAULT '',
  nombre_comun         TEXT NOT NULL DEFAULT '',
  nombre_cientifico    TEXT NOT NULL DEFAULT '',
  especie_id           TEXT,
  prioridad            TEXT NOT NULL DEFAULT 'media',
  comentario           TEXT NOT NULL DEFAULT '',
  solicitud_id         UUID REFERENCES solicitudes (id),
  actividad_id         UUID REFERENCES actividades (id),
  origen_ref           TEXT,
  archivada_en         TIMESTAMPTZ,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT podas_prioridad_chk CHECK (prioridad IN ('baja', 'media', 'alta')),
  CONSTRAINT podas_cantidades_chk CHECK (cantidad_pedida >= 0 AND cantidad_ejecutada >= 0),
  CONSTRAINT podas_fechas_chk CHECK (
    fecha_reporte IS NULL OR fecha_ejecucion IS NULL OR fecha_ejecucion >= fecha_reporte
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS podas_origen_ref_uidx
  ON podas (origen_ref) WHERE origen_ref IS NOT NULL AND origen_ref <> '';

CREATE INDEX IF NOT EXISTS podas_codigo_externo_idx
  ON podas (codigo_externo) WHERE codigo_externo IS NOT NULL AND codigo_externo <> '';

COMMENT ON COLUMN podas.codigo_externo IS
  'OSG-… solo si la fuente lo trae. El sistema no lo genera.';
COMMENT ON TABLE podas IS
  'Registros de poda. La baja es lógica (archivada_en).';
