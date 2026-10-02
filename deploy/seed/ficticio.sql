-- Semilla ficticia aditiva de develop y qa.
-- Cada INSERT usa ON CONFLICT DO NOTHING: no borra, no trunca y no pisa filas ya cargadas.
-- Los polígonos 0101–0106 cubren el campus (EPSG:4326) y se ven a la escala del mapa.
-- Nombres de demostración. No hay personas.

INSERT INTO areas_verdes (
  feature_id, source_index, codigo, nombre, uso, referencia, perimetro_m, area_m2, geom, activo
) VALUES (
  'AV-FICT-0001',
  900001,
  'FICT-N',
  'Jardín Ficticio Norte',
  'jardin',
  'Polígono de demostración. No es un predio real.',
  40,
  100,
  ST_SetSRID(ST_Multi(ST_GeomFromText(
    'POLYGON((-77.0800 -12.0700, -77.0799 -12.0700, -77.0799 -12.0699, -77.0800 -12.0699, -77.0800 -12.0700))'
  )), 4326),
  TRUE
) ON CONFLICT (feature_id) DO NOTHING;

INSERT INTO poligonos_cuadrilla (
  feature_id, source_index, codigo, nombre, uso, referencia, perimetro_m, area_m2, geom, cuadrilla_id, activo
) VALUES (
  'Z-FICT-0001',
  900001,
  'FICT-Z',
  'Sector Ficticio Norte',
  'jardin',
  'Cuadrilla ficticia Equipo Norte. Sin personas reales.',
  40,
  100,
  ST_SetSRID(ST_Multi(ST_GeomFromText(
    'POLYGON((-77.0800 -12.0700, -77.0799 -12.0700, -77.0799 -12.0699, -77.0800 -12.0699, -77.0800 -12.0700))'
  )), 4326),
  'cap-norte',
  TRUE
) ON CONFLICT (feature_id) DO NOTHING;

INSERT INTO areas_verdes (
  feature_id, source_index, codigo, nombre, uso, referencia, perimetro_m, area_m2, geom, activo
) VALUES
  (
    'AV-FICT-0101', 910001, 'FICT-01', 'Prado Ficticio Institucional',
    'Uso Institucional', 'Demostración. No es un predio real.', 360, 8100,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08120 -12.06620, -77.08020 -12.06620, -77.08020 -12.06540, -77.08120 -12.06540, -77.08120 -12.06620))')), 4326),
    TRUE
  ),
  (
    'AV-FICT-0102', 910002, 'FICT-02', 'Patio Ficticio Administrativo',
    'Uso Administrativo', 'Demostración. No es un predio real.', 360, 8100,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.07980 -12.06980, -77.07880 -12.06980, -77.07880 -12.06900, -77.07980 -12.06900, -77.07980 -12.06980))')), 4326),
    TRUE
  ),
  (
    'AV-FICT-0103', 910003, 'FICT-03', 'Descanso Ficticio Este',
    'Áreas de descanso y recreación', 'Demostración. No es un predio real.', 360, 8100,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.07860 -12.06820, -77.07770 -12.06820, -77.07770 -12.06740, -77.07860 -12.06740, -77.07860 -12.06820))')), 4326),
    TRUE
  ),
  (
    'AV-FICT-0104', 910004, 'FICT-04', 'Manejo Ficticio Sur',
    'Manejo sostenible y reducción de consumo', 'Demostración. No es un predio real.', 360, 8100,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08100 -12.07280, -77.08000 -12.07280, -77.08000 -12.07200, -77.08100 -12.07200, -77.08100 -12.07280))')), 4326),
    TRUE
  ),
  (
    'AV-FICT-0105', 910005, 'FICT-05', 'Cancha Ficticia Oeste',
    'Áreas deportivas y recreación activa', 'Demostración. No es un predio real.', 360, 8100,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08280 -12.07050, -77.08180 -12.07050, -77.08180 -12.06970, -77.08280 -12.06970, -77.08280 -12.07050))')), 4326),
    TRUE
  ),
  (
    'AV-FICT-0106', 910006, 'FICT-06', 'Reserva Ficticia del Humedal',
    'Áreas de conservación', 'Demostración. No es un predio real.', 360, 8100,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.07920 -12.07220, -77.07820 -12.07220, -77.07820 -12.07140, -77.07920 -12.07140, -77.07920 -12.07220))')), 4326),
    TRUE
  )
ON CONFLICT (feature_id) DO NOTHING;

INSERT INTO poligonos_cuadrilla (
  feature_id, source_index, codigo, nombre, uso, referencia, perimetro_m, area_m2, geom, cuadrilla_id, sector, activo
) VALUES
  (
    'Z-FICT-0101', 910001, 'FICT-S1', 'Sector Ficticio Valeria',
    'Uso Institucional', 'Cuadrilla ficticia. Sin personas reales.', 400, 10000,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08140 -12.06640, -77.08000 -12.06640, -77.08000 -12.06520, -77.08140 -12.06520, -77.08140 -12.06640))')), 4326),
    'cap-norte', 'cua-valeria', TRUE
  ),
  (
    'Z-FICT-0102', 910002, 'FICT-S2', 'Sector Ficticio Mateo',
    'Uso Administrativo', 'Cuadrilla ficticia. Sin personas reales.', 400, 10000,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08000 -12.07000, -77.07860 -12.07000, -77.07860 -12.06880, -77.08000 -12.06880, -77.08000 -12.07000))')), 4326),
    'cap-norte', 'cua-mateo', TRUE
  ),
  (
    'Z-FICT-0103', 910003, 'FICT-S3', 'Sector Ficticio Renato',
    'Áreas de descanso y recreación', 'Cuadrilla ficticia. Sin personas reales.', 400, 10000,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.07880 -12.06840, -77.07760 -12.06840, -77.07760 -12.06720, -77.07880 -12.06720, -77.07880 -12.06840))')), 4326),
    'cap-norte', 'cua-renato', TRUE
  ),
  (
    'Z-FICT-0104', 910004, 'FICT-S4', 'Sector Ficticio Deportivo',
    'Áreas deportivas y recreación activa', 'Rótulo de lugar. Sin personas reales.', 400, 10000,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08300 -12.07070, -77.08160 -12.07070, -77.08160 -12.06950, -77.08300 -12.06950, -77.08300 -12.07070))')), 4326),
    'cap-norte', 'campo-deportivo', TRUE
  ),
  (
    'Z-FICT-0105', 910005, 'FICT-S5', 'Sector Ficticio Bosque Húmedo',
    'Áreas de conservación', 'Rótulo de lugar. Sin personas reales.', 400, 10000,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08120 -12.07300, -77.07980 -12.07300, -77.07980 -12.07180, -77.08120 -12.07180, -77.08120 -12.07300))')), 4326),
    'cap-norte', 'bosque-humedo', TRUE
  )
ON CONFLICT (feature_id) DO NOTHING;
