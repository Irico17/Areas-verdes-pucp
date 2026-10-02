-- Puntos PUCP (sección 4.15). Sin teléfono, placeId, website ni imagen de sesión.

CREATE TABLE IF NOT EXISTS puntos_pucp (
  id          BIGSERIAL PRIMARY KEY,
  titulo      TEXT NOT NULL,
  lat         DOUBLE PRECISION NOT NULL,
  lon         DOUBLE PRECISION NOT NULL,
  url         TEXT,
  geom        geometry(Point, 4326) NOT NULL,
  origen_ref  TEXT NOT NULL,
  activo      BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT puntos_pucp_lat_chk CHECK (lat BETWEEN -12.20 AND -11.90),
  CONSTRAINT puntos_pucp_lon_chk CHECK (lon BETWEEN -77.30 AND -76.90),
  CONSTRAINT puntos_pucp_url_chk CHECK (
    url IS NULL OR (url NOT LIKE '%place_id=%' AND url !~* 'phone' AND position('placeId' in url) = 0)
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS puntos_pucp_origen_ref_uidx ON puntos_pucp (origen_ref);
CREATE INDEX IF NOT EXISTS puntos_pucp_geom_gix ON puntos_pucp USING GIST (geom);
CREATE INDEX IF NOT EXISTS puntos_pucp_titulo_idx ON puntos_pucp (lower(titulo));

COMMENT ON TABLE puntos_pucp IS
  'Lugar público del campus. Solo título, coordenada y URL de mapa. No hay columnas de contacto.';
