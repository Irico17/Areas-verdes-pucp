-- Cuadrillas. El seed Norte/Sur/Riego sigue siendo de demostración.
-- Los tres ficticios del campus los carga el ETL, no esta migración.

CREATE TABLE IF NOT EXISTS cuadrillas (
  id               TEXT PRIMARY KEY,
  nombre_ficticio  TEXT NOT NULL,
  turno            TEXT NOT NULL,
  activo           BOOLEAN NOT NULL DEFAULT TRUE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT cuadrillas_turno_chk CHECK (turno IN ('manana', 'tarde'))
);

COMMENT ON TABLE cuadrillas IS
  'Equipos de campo. nombre_ficticio no es una persona real.';

INSERT INTO cuadrillas (id, nombre_ficticio, turno, activo)
SELECT
  id,
  equipo,
  CASE WHEN turno IN ('mañana', 'manana') THEN 'manana' ELSE 'tarde' END,
  activo
FROM capataces
ON CONFLICT (id) DO NOTHING;

ALTER TABLE usuarios ADD COLUMN IF NOT EXISTS cuadrilla_id TEXT REFERENCES cuadrillas (id);

COMMENT ON COLUMN usuarios.cuadrilla_id IS
  'Cuadrilla que opera la cuenta, si aplica. No identifica a una persona real.';
