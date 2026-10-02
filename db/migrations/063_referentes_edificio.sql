-- El edificio se guarda por id del extracto, no como texto suelto.
-- lugar_id apunta al catálogo de lugares. No se borra lugar_libre.

CREATE TABLE IF NOT EXISTS referentes_edificio (
  id           BIGSERIAL PRIMARY KEY,
  lugar_id     BIGINT NOT NULL REFERENCES lugares (id),
  edificio_id  TEXT NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT referentes_edificio_par_uidx UNIQUE (lugar_id, edificio_id)
);

CREATE INDEX IF NOT EXISTS referentes_edificio_lugar_idx
  ON referentes_edificio (lugar_id);

COMMENT ON TABLE referentes_edificio IS
  'Par lugar de catálogo + id de edificio. No acepta un nombre de edificio escrito a mano.';

COMMENT ON COLUMN referentes_edificio.edificio_id IS
  'id del GeoJSON de edificios (por ejemplo osm-way-…). No es un texto libre.';
