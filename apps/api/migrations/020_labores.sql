-- Frente 2A. Extiende actividades (no crea un segundo monitoreo).
-- geom puede ser NULL si hay lugar o zona de supervisión.

ALTER TABLE actividades ADD COLUMN IF NOT EXISTS fecha_solicitud DATE;
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS fecha_atencion DATE;
-- lugar_id y zona_supervision_id ya son BIGINT con FK (016). No se recrean como TEXT.
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS cuadrilla_id TEXT;
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS clase_codigo TEXT;
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS tipo_codigo TEXT;
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS comentario TEXT NOT NULL DEFAULT '';
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS lugar_libre TEXT NOT NULL DEFAULT '';
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS origen_ref TEXT;
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS origen TEXT NOT NULL DEFAULT 'interna';

ALTER TABLE actividades ALTER COLUMN geom DROP NOT NULL;

ALTER TABLE actividades DROP CONSTRAINT IF EXISTS actividades_ubicacion_chk;
ALTER TABLE actividades DROP CONSTRAINT IF EXISTS actividades_geom_lugar_chk;
ALTER TABLE actividades ADD CONSTRAINT actividades_ubicacion_chk CHECK (
  geom IS NOT NULL
  OR lugar_id IS NOT NULL
  OR zona_supervision_id IS NOT NULL
  OR (origen_ref IS NOT NULL AND origen_ref <> '')
);

ALTER TABLE actividades DROP CONSTRAINT IF EXISTS actividades_fechas_chk;
ALTER TABLE actividades ADD CONSTRAINT actividades_fechas_chk CHECK (
  fecha_solicitud IS NULL
  OR fecha_atencion IS NULL
  OR fecha_atencion >= fecha_solicitud
);

CREATE UNIQUE INDEX IF NOT EXISTS actividades_origen_ref_uidx
  ON actividades (origen_ref) WHERE origen_ref IS NOT NULL AND origen_ref <> '';

COMMENT ON COLUMN actividades.origen_ref IS
  'Id de la fila fuente para reimportar. No guarda datos personales.';
COMMENT ON COLUMN actividades.estado IS
  'Incluye sin_estado cuando la hoja de monitoreo trae el estado vacío.';

INSERT INTO catalogos (clase, codigo, nombre, orden) VALUES
  ('estado', 'sin_estado', 'Sin estado', 0),
  ('clase_actividad', 'habilitacion', 'Habilitación de jardines', 1),
  ('clase_actividad', 'rehabilitacion', 'Rehabilitación y rediseño de jardines', 2),
  ('clase_actividad', 'mantenimiento', 'Mantenimiento de jardines', 3),
  ('clase_actividad', 'poda', 'Poda', 4),
  ('clase_actividad', 'propagacion', 'Propagación y plantación', 5),
  ('clase_actividad', 'riego', 'Riego', 6),
  ('clase_actividad', 'residuos', 'Manejo de residuos vegetales', 7)
ON CONFLICT (clase, codigo) DO NOTHING;

CREATE TABLE IF NOT EXISTS cuadrillas (
  id               TEXT PRIMARY KEY,
  nombre_ficticio  TEXT NOT NULL,
  turno            TEXT NOT NULL DEFAULT 'manana',
  activo           BOOLEAN NOT NULL DEFAULT TRUE,
  CONSTRAINT cuadrillas_turno_chk CHECK (turno IN ('manana', 'tarde'))
);

COMMENT ON TABLE cuadrillas IS
  'Responsables ficticios. No hay personas reales.';

INSERT INTO cuadrillas (id, nombre_ficticio, turno) VALUES
  ('cua-valeria', 'Valeria Quispe', 'manana'),
  ('cua-mateo', 'Mateo Salazar', 'manana'),
  ('cua-renato', 'Renato Cárdenas', 'tarde'),
  ('cua-nora', 'Nora Beltrán', 'manana'),
  ('cua-ivan', 'Iván Paredes', 'tarde'),
  ('cua-lucia', 'Lucía Mendoza', 'manana')
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS personal_labor (
  id            UUID PRIMARY KEY,
  actividad_id  UUID NOT NULL REFERENCES actividades (id),
  nombre_ficticio TEXT NOT NULL,
  rol_campo     TEXT NOT NULL DEFAULT 'operario de cuadrilla',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE personal_labor IS
  'Participantes de la labor. Solo nombres ficticios.';

CREATE TABLE IF NOT EXISTS actividad_avances (
  id               UUID PRIMARY KEY,
  actividad_id     UUID NOT NULL REFERENCES actividades (id),
  fecha            DATE NOT NULL,
  nota             TEXT NOT NULL DEFAULT '',
  area_feature_id  TEXT,
  ejemplar_ref     TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE actividad_avances IS
  'Avance de varios días ligado al área o al ejemplar. No hay borrado físico.';

CREATE INDEX IF NOT EXISTS actividad_avances_act_idx ON actividad_avances (actividad_id, fecha);
