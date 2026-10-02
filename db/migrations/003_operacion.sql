-- Operación Fase D. Labores de campo sobre el mismo Postgres/PostGIS.
-- Los capataces son equipos ficticios: no hay personas ni datos reales.

CREATE TABLE IF NOT EXISTS capataces (
  id      TEXT PRIMARY KEY,
  equipo  TEXT NOT NULL,
  turno   TEXT NOT NULL,
  activo  BOOLEAN NOT NULL DEFAULT TRUE
);

COMMENT ON TABLE capataces IS
  'Equipos de campo de demostración. No identifican a una persona.';

INSERT INTO capataces (id, equipo, turno) VALUES
  ('cap-norte', 'Equipo Norte', 'mañana'),
  ('cap-sur', 'Equipo Sur', 'mañana'),
  ('cap-riego', 'Equipo Riego', 'tarde')
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS actividades (
  id                   UUID PRIMARY KEY,
  tipo                 TEXT NOT NULL,
  estado               TEXT NOT NULL,
  titulo               TEXT NOT NULL,
  detalle              TEXT NOT NULL DEFAULT '',
  area_feature_id      TEXT,
  zona_feature_id      TEXT,
  assigned_capataz_id  TEXT REFERENCES capataces (id),
  geom                 geometry(Point, 4326) NOT NULL,
  archivada_en         TIMESTAMPTZ,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT actividades_tipo_chk CHECK (tipo IN ('riego', 'poda', 'limpieza', 'incidencia', 'inspeccion')),
  CONSTRAINT actividades_estado_chk CHECK (estado IN ('pendiente', 'en_proceso', 'bloqueada', 'cerrada', 'cancelada'))
);

CREATE INDEX IF NOT EXISTS actividades_geom_gix ON actividades USING GIST (geom);
CREATE INDEX IF NOT EXISTS actividades_asignacion_idx ON actividades (assigned_capataz_id);

COMMENT ON TABLE actividades IS
  'Labores de supervisión. geom es Point EPSG:4326. archivada_en es la baja lógica.';
COMMENT ON COLUMN actividades.assigned_capataz_id IS
  'Equipo stub (capataces.id), no un usuario institucional.';

CREATE TABLE IF NOT EXISTS actividad_eventos (
  id            BIGSERIAL PRIMARY KEY,
  actividad_id  UUID NOT NULL REFERENCES actividades (id),
  tipo          TEXT NOT NULL,
  estado        TEXT,
  capataz_id    TEXT,
  actor_rol     TEXT NOT NULL,
  nota          TEXT NOT NULL DEFAULT '',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT actividad_eventos_tipo_chk CHECK (
    tipo IN ('creada', 'asignada', 'reasignada', 'estado', 'cancelada', 'archivada')
  )
);

CREATE INDEX IF NOT EXISTS actividad_eventos_act_idx ON actividad_eventos (actividad_id, id);

-- Semilla fija para la demo. Coordenadas dentro del campus Pando.
INSERT INTO actividades (
  id, tipo, estado, titulo, detalle, assigned_capataz_id, geom, created_at, updated_at
) VALUES
  (
    '11111111-1111-4111-8111-111111111111',
    'riego', 'pendiente',
    'Riego del eje central',
    'Revisar aspersores de la explanada central.',
    'cap-norte',
    ST_SetSRID(ST_MakePoint(-77.07955, -12.06955), 4326),
    TIMESTAMPTZ '2026-09-24 08:10:00-05',
    TIMESTAMPTZ '2026-09-24 08:10:00-05'
  ),
  (
    '22222222-2222-4222-8222-222222222222',
    'poda', 'en_proceso',
    'Poda de setos sur',
    'Setos del borde sur, altura de mantenimiento.',
    'cap-sur',
    ST_SetSRID(ST_MakePoint(-77.08080, -12.07040), 4326),
    TIMESTAMPTZ '2026-09-24 08:20:00-05',
    TIMESTAMPTZ '2026-09-24 09:05:00-05'
  ),
  (
    '33333333-3333-4333-8333-333333333333',
    'limpieza', 'pendiente',
    'Limpieza de caminería norte',
    'Hojas y residuos en la caminería del sector norte.',
    'cap-norte',
    ST_SetSRID(ST_MakePoint(-77.07890, -12.06820), 4326),
    TIMESTAMPTZ '2026-09-24 08:30:00-05',
    TIMESTAMPTZ '2026-09-24 08:30:00-05'
  ),
  (
    '44444444-4444-4444-8444-444444444444',
    'incidencia', 'bloqueada',
    'Fuga en línea de riego',
    'Espera de corte de agua antes de intervenir.',
    'cap-riego',
    ST_SetSRID(ST_MakePoint(-77.08120, -12.07150), 4326),
    TIMESTAMPTZ '2026-09-24 08:40:00-05',
    TIMESTAMPTZ '2026-09-24 10:00:00-05'
  ),
  (
    '55555555-5555-4555-8555-555555555555',
    'inspeccion', 'pendiente',
    'Inspección bosque húmedo',
    'Recorrido de control del sector bosque húmedo.',
    'cap-sur',
    ST_SetSRID(ST_MakePoint(-77.07920, -12.07280), 4326),
    TIMESTAMPTZ '2026-09-24 08:50:00-05',
    TIMESTAMPTZ '2026-09-24 08:50:00-05'
  ),
  (
    '66666666-6666-4666-8666-666666666666',
    'riego', 'cerrada',
    'Riego de losas, turno temprano',
    'Cerrada en la semilla: no debe salir en el listado abierto.',
    'cap-riego',
    ST_SetSRID(ST_MakePoint(-77.08020, -12.06680), 4326),
    TIMESTAMPTZ '2026-09-23 07:00:00-05',
    TIMESTAMPTZ '2026-09-23 11:00:00-05'
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO actividad_eventos (actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at)
SELECT id, 'creada', estado, assigned_capataz_id, 'coordinacion', 'Carga de demostración', created_at
FROM actividades
WHERE id IN (
  '11111111-1111-4111-8111-111111111111',
  '22222222-2222-4222-8222-222222222222',
  '33333333-3333-4333-8333-333333333333',
  '44444444-4444-4444-8444-444444444444',
  '55555555-5555-4555-8555-555555555555',
  '66666666-6666-4666-8666-666666666666'
)
AND NOT EXISTS (
  SELECT 1 FROM actividad_eventos e WHERE e.actividad_id = actividades.id AND e.tipo = 'creada'
);

INSERT INTO actividad_eventos (actividad_id, tipo, estado, capataz_id, actor_rol, nota, created_at)
SELECT id, 'asignada', estado, assigned_capataz_id, 'coordinacion', 'Asignación inicial de demostración', created_at
FROM actividades
WHERE assigned_capataz_id IS NOT NULL
  AND id IN (
    '11111111-1111-4111-8111-111111111111',
    '22222222-2222-4222-8222-222222222222',
    '33333333-3333-4333-8333-333333333333',
    '44444444-4444-4444-8444-444444444444',
    '55555555-5555-4555-8555-555555555555',
    '66666666-6666-4666-8666-666666666666'
  )
  AND NOT EXISTS (
    SELECT 1 FROM actividad_eventos e WHERE e.actividad_id = actividades.id AND e.tipo = 'asignada'
  );
