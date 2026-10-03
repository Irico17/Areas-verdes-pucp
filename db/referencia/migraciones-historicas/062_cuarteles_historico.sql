-- Cuarteles forestales, solo como historia. La geometría no es obligatoria
-- y esta migración no inserta ninguna fila: sin el archivo del cliente
-- la capa se lee vacía.

CREATE TABLE IF NOT EXISTS cuarteles_historico (
  id          BIGSERIAL PRIMARY KEY,
  codigo      TEXT NOT NULL,
  nombre      TEXT NOT NULL,
  geom        geometry(MultiPolygon, 4326),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS cuarteles_historico_codigo_uidx
  ON cuarteles_historico (codigo);

COMMENT ON TABLE cuarteles_historico IS
  'Referencia histórica de solo lectura. geom puede ser NULL. Sin archivo de cuarteles no hay filas.';

COMMENT ON COLUMN cuarteles_historico.geom IS
  'Polígono del shape del cliente, si llega. Nunca se rellena con una geometría de ejemplo.';
