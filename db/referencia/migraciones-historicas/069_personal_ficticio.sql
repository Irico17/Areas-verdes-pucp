-- Personal de la actividad. Nombres ficticios. No crea cuentas en usuarios.

CREATE TABLE IF NOT EXISTS personal_ficticio (
  id               TEXT PRIMARY KEY,
  nombre_ficticio  TEXT NOT NULL,
  activo           BOOLEAN NOT NULL DEFAULT TRUE
);

COMMENT ON TABLE personal_ficticio IS
  'Nombres ficticios que se pueden asignar a una actividad. No son cuentas ni personas reales.';

INSERT INTO personal_ficticio (id, nombre_ficticio) VALUES
  ('pf-elsa', 'Elsa Mamani'),
  ('pf-marco', 'Marco Huamán'),
  ('pf-nelida', 'Nélida Choque'),
  ('pf-pedro', 'Pedro Salas'),
  ('pf-rita', 'Rita Ccopa'),
  ('pf-tomas', 'Tomás Alegre')
ON CONFLICT (id) DO NOTHING;
