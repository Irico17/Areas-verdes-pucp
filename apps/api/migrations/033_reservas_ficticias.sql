-- Reservas ficticias (sección 4.14). La hoja real sigue en HTTP 401: no se abre.
-- origen solo admite ficticio hasta que la universidad entregue el archivo.

CREATE TABLE IF NOT EXISTS reservas_jardin (
  id           BIGSERIAL PRIMARY KEY,
  jardin_id    BIGINT REFERENCES jardines_reserva (id),
  fecha        DATE NOT NULL,
  hora_inicio  TIME NOT NULL,
  hora_fin     TIME NOT NULL,
  estado       TEXT NOT NULL,
  evento       TEXT NOT NULL,
  unidad       TEXT,
  origen       TEXT NOT NULL DEFAULT 'ficticio',
  origen_ref   TEXT NOT NULL,
  activo       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT reservas_jardin_estado_chk CHECK (estado IN ('reservado', 'realizado', 'cancelado')),
  CONSTRAINT reservas_jardin_origen_chk CHECK (origen = 'ficticio'),
  CONSTRAINT reservas_jardin_hora_chk CHECK (hora_fin > hora_inicio)
);

CREATE UNIQUE INDEX IF NOT EXISTS reservas_jardin_origen_ref_uidx ON reservas_jardin (origen_ref);
CREATE INDEX IF NOT EXISTS reservas_jardin_fecha_idx ON reservas_jardin (fecha, hora_inicio);

COMMENT ON TABLE reservas_jardin IS
  'Agenda ficticia. No proviene de la hoja de reservas (HTTP 401).';
COMMENT ON COLUMN reservas_jardin.unidad IS
  'Unidad organizativa, no una persona.';
COMMENT ON COLUMN reservas_jardin.origen IS
  'Siempre ficticio mientras la hoja institucional responda 401.';
