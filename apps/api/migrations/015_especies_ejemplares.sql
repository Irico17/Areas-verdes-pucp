-- Especie, ejemplar, historial de códigos (HUID 04 y 05) y medida de palmera.
-- salud queda NULL hasta que la universidad entregue el dato.

CREATE TABLE IF NOT EXISTS especies (
  id                  BIGSERIAL PRIMARY KEY,
  nombre_cientifico   TEXT NOT NULL UNIQUE,
  nombre_comun        TEXT,
  activo              BOOLEAN NOT NULL DEFAULT TRUE,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE especies IS
  'Especie de un ejemplar. Se da de alta cuando el nombre científico es nuevo.';

CREATE TABLE IF NOT EXISTS ejemplares (
  id                     BIGSERIAL PRIMARY KEY,
  numero_origen          INTEGER UNIQUE,
  codigo                 TEXT,
  especie_id             BIGINT REFERENCES especies (id),
  nombre_comun           TEXT,
  tipo_vegetacion        TEXT,
  cantidad               INTEGER NOT NULL DEFAULT 1,
  area_verde_id          BIGINT REFERENCES areas_verdes (id),
  ubicacion_lugar_id     BIGINT REFERENCES lugares (id),
  referencia             TEXT,
  lat                    DOUBLE PRECISION,
  lon                    DOUBLE PRECISION,
  observacion_fen_2026   TEXT,
  foto_id                UUID REFERENCES evidencias (id),
  salud                  TEXT,
  geom                   geometry(Point, 4326),
  origen_ref             TEXT,
  activo                 BOOLEAN NOT NULL DEFAULT TRUE,
  created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT ejemplares_cantidad_chk CHECK (cantidad >= 1),
  CONSTRAINT ejemplares_tipo_chk CHECK (
    tipo_vegetacion IS NULL OR tipo_vegetacion IN (
      'Árbol', 'Palmera', 'Arbusto', 'Herbácea', 'Trepadora', 'Suculenta', 'cafeto'
    )
  ),
  CONSTRAINT ejemplares_lat_chk CHECK (lat IS NULL OR lat BETWEEN -12.20 AND -11.90),
  CONSTRAINT ejemplares_lon_chk CHECK (lon IS NULL OR lon BETWEEN -77.30 AND -76.90)
);

CREATE INDEX IF NOT EXISTS ejemplares_geom_gix ON ejemplares USING GIST (geom);
CREATE INDEX IF NOT EXISTS ejemplares_especie_idx ON ejemplares (especie_id);
CREATE INDEX IF NOT EXISTS ejemplares_lugar_idx ON ejemplares (ubicacion_lugar_id);

COMMENT ON TABLE ejemplares IS
  'Ejemplar de flora. salud permanece NULL: la fuente no trae ese dato.';
COMMENT ON COLUMN ejemplares.numero_origen IS
  'N° de la hoja de origen. Duplicado = el mismo número.';
COMMENT ON COLUMN ejemplares.foto_id IS
  'Archivo propio. No guarda un id de Drive.';

CREATE TABLE IF NOT EXISTS codigos_historicos (
  id                BIGSERIAL PRIMARY KEY,
  ejemplar_id       BIGINT NOT NULL REFERENCES ejemplares (id),
  codigo_anterior   TEXT NOT NULL,
  codigo_nuevo      TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS codigos_historicos_ejemplar_idx
  ON codigos_historicos (ejemplar_id, id);

COMMENT ON TABLE codigos_historicos IS
  'Código anterior de un ejemplar. La recodificación masiva no vive aquí.';

CREATE TABLE IF NOT EXISTS medidas_palmera (
  ejemplar_id    BIGINT PRIMARY KEY REFERENCES ejemplares (id),
  utm_norte      DOUBLE PRECISION,
  utm_este       DOUBLE PRECISION,
  altura         DOUBLE PRECISION,
  altura_fuste   DOUBLE PRECISION,
  dap            DOUBLE PRECISION,
  radio          DOUBLE PRECISION,
  zunchado       BOOLEAN,
  CONSTRAINT medidas_palmera_altura_chk CHECK (altura IS NULL OR altura >= 0),
  CONSTRAINT medidas_palmera_fuste_chk CHECK (altura_fuste IS NULL OR altura_fuste >= 0),
  CONSTRAINT medidas_palmera_dap_chk CHECK (dap IS NULL OR dap >= 0),
  CONSTRAINT medidas_palmera_radio_chk CHECK (radio IS NULL OR radio >= 0)
);

COMMENT ON TABLE medidas_palmera IS
  'Medidas 1:1 de una palmera. La fila se rechaza en la carga si la coordenada no cae en el campus.';
