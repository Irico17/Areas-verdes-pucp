-- Base operativa: sesión local, catálogos, catastro editable, riego,
-- solicitudes, órdenes, evidencias. No hay SSO ni borrado físico.

ALTER TABLE actividades DROP CONSTRAINT IF EXISTS actividades_tipo_chk;
ALTER TABLE actividades DROP CONSTRAINT IF EXISTS actividades_estado_chk;

ALTER TABLE actividades ADD COLUMN IF NOT EXISTS ejecutor TEXT NOT NULL DEFAULT 'propia';
ALTER TABLE actividades ADD COLUMN IF NOT EXISTS motivo_archivo TEXT;

ALTER TABLE actividades DROP CONSTRAINT IF EXISTS actividades_ejecutor_chk;
ALTER TABLE actividades ADD CONSTRAINT actividades_ejecutor_chk
  CHECK (ejecutor IN ('propia', 'tercerizada'));

ALTER TABLE actividad_eventos DROP CONSTRAINT IF EXISTS actividad_eventos_tipo_chk;
ALTER TABLE actividad_eventos ADD CONSTRAINT actividad_eventos_tipo_chk CHECK (
  tipo IN ('creada', 'asignada', 'reasignada', 'estado', 'cancelada', 'archivada', 'evidencia')
);

CREATE TABLE IF NOT EXISTS usuarios (
  id            BIGSERIAL PRIMARY KEY,
  usuario       TEXT NOT NULL UNIQUE,
  nombre        TEXT NOT NULL,
  rol           TEXT NOT NULL,
  capataz_id    TEXT REFERENCES capataces (id),
  password_hash TEXT NOT NULL,
  activo        BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT usuarios_rol_chk CHECK (rol IN ('capataz', 'coordinacion', 'jefatura', 'admin'))
);

COMMENT ON TABLE usuarios IS
  'Cuentas locales de desarrollo. No son el SSO de la PUCP.';

CREATE TABLE IF NOT EXISTS sesiones (
  token_hash TEXT PRIMARY KEY,
  usuario_id BIGINT NOT NULL REFERENCES usuarios (id),
  expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS permisos (
  rol    TEXT NOT NULL,
  accion TEXT NOT NULL,
  PRIMARY KEY (rol, accion)
);

CREATE TABLE IF NOT EXISTS catalogos (
  id      BIGSERIAL PRIMARY KEY,
  clase   TEXT NOT NULL,
  codigo  TEXT NOT NULL,
  nombre  TEXT NOT NULL,
  activo  BOOLEAN NOT NULL DEFAULT TRUE,
  orden   INTEGER NOT NULL DEFAULT 0,
  UNIQUE (clase, codigo)
);

COMMENT ON TABLE catalogos IS
  'Valores parametrizables. La baja es lógica (activo = false).';

INSERT INTO catalogos (clase, codigo, nombre, orden) VALUES
  ('tipo_actividad', 'riego', 'Riego', 1),
  ('tipo_actividad', 'poda', 'Poda', 2),
  ('tipo_actividad', 'limpieza', 'Limpieza', 3),
  ('tipo_actividad', 'incidencia', 'Incidencia', 4),
  ('tipo_actividad', 'inspeccion', 'Inspección', 5),
  ('estado', 'pendiente', 'Pendiente', 1),
  ('estado', 'en_proceso', 'En proceso', 2),
  ('estado', 'bloqueada', 'Bloqueada', 3),
  ('estado', 'cerrada', 'Cerrada', 4),
  ('estado', 'cancelada', 'Cancelada', 5),
  ('prioridad', 'baja', 'Baja', 1),
  ('prioridad', 'media', 'Media', 2),
  ('prioridad', 'alta', 'Alta', 3),
  ('lugar', 'norte', 'Sector norte', 1),
  ('lugar', 'sur', 'Sector sur', 2),
  ('lugar', 'eje', 'Eje central', 3),
  ('lugar', 'bosque', 'Bosque húmedo', 4),
  ('especie', 'tipuana', 'Tipuana tipu', 1),
  ('especie', 'ficus', 'Ficus benjamina', 2),
  ('especie', 'schinus', 'Schinus molle', 3),
  ('motivo_archivo', 'error', 'Registrada por error', 1),
  ('motivo_archivo', 'duplicada', 'Duplicada', 2),
  ('motivo_archivo', 'no_corresponde', 'Ya no corresponde', 3),
  ('motivo_archivo', 'otro', 'Otro', 4),
  ('turno', 'manana', 'Mañana', 1),
  ('turno', 'tarde', 'Tarde', 2),
  ('fuente', 'centuria', 'Centuria', 1),
  ('fuente', 'osg', 'Matriz OSG', 2),
  ('fuente', 'correo', 'Correo', 3),
  ('fuente', 'interna', 'Interna', 4)
ON CONFLICT (clase, codigo) DO NOTHING;

CREATE TABLE IF NOT EXISTS solicitudes (
  id              UUID PRIMARY KEY,
  codigo_externo  TEXT,
  fuente          TEXT NOT NULL,
  titulo          TEXT NOT NULL,
  detalle         TEXT NOT NULL DEFAULT '',
  prioridad       TEXT NOT NULL DEFAULT 'media',
  estado          TEXT NOT NULL DEFAULT 'por_iniciar',
  lugar           TEXT,
  cantidad        INTEGER,
  actividad_id    UUID REFERENCES actividades (id),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT solicitudes_fuente_chk CHECK (fuente IN ('centuria', 'osg', 'correo', 'interna')),
  CONSTRAINT solicitudes_estado_chk CHECK (estado IN ('por_iniciar', 'en_proceso', 'ejecutado', 'cerrado', 'cancelado'))
);

CREATE UNIQUE INDEX IF NOT EXISTS solicitudes_codigo_externo_uidx
  ON solicitudes (codigo_externo) WHERE codigo_externo IS NOT NULL AND codigo_externo <> '';

COMMENT ON COLUMN solicitudes.codigo_externo IS
  'Código de Centuria u OSG si existe. El sistema no lo inventa.';

CREATE TABLE IF NOT EXISTS ordenes_servicio (
  id            UUID PRIMARY KEY,
  actividad_id  UUID NOT NULL REFERENCES actividades (id),
  empresa       TEXT NOT NULL,
  referencia    TEXT NOT NULL,
  frecuencia    TEXT NOT NULL DEFAULT '',
  estado        TEXT NOT NULL DEFAULT 'en_proceso',
  conformidad   TEXT NOT NULL DEFAULT '',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT ordenes_estado_chk CHECK (estado IN ('en_proceso', 'ejecutada', 'conforme'))
);

CREATE TABLE IF NOT EXISTS riego_registros (
  id          UUID PRIMARY KEY,
  sector      TEXT NOT NULL,
  turno       TEXT NOT NULL,
  capataz_id  TEXT REFERENCES capataces (id),
  fecha       DATE NOT NULL,
  nota        TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE riego_registros IS
  'Cobertura mínima de riego. No calcula un indicador oficial.';

CREATE TABLE IF NOT EXISTS evidencias (
  id            UUID PRIMARY KEY,
  actividad_id  UUID REFERENCES actividades (id),
  solicitud_id  UUID REFERENCES solicitudes (id),
  orden_id      UUID REFERENCES ordenes_servicio (id),
  nombre        TEXT NOT NULL,
  mime          TEXT NOT NULL,
  bytes         INTEGER NOT NULL,
  ruta          TEXT NOT NULL,
  nota          TEXT NOT NULL DEFAULT '',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO solicitudes (
  id, codigo_externo, fuente, titulo, detalle, prioridad, estado, lugar, actividad_id
) VALUES (
  '77777777-7777-4777-8777-777777777777',
  'OSG-2026-0142',
  'osg',
  'Fuga reportada en línea de riego',
  'Aviso manual. No hay integración con Centuria.',
  'alta',
  'en_proceso',
  'Eje central',
  '44444444-4444-4444-8444-444444444444'
) ON CONFLICT (id) DO NOTHING;
