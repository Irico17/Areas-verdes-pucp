-- Semilla ficticia aditiva de develop y qa.
-- INSERT ... ON CONFLICT actualiza solo las filas ficticias de este archivo
-- (su geometría y su nombre). No borra, no trunca y no toca el catastro del ETL.
-- Los polígonos 0101–0106 quedan en el parque oeste, lejos de los edificios.
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
) ON CONFLICT (feature_id) DO UPDATE SET
  geom = EXCLUDED.geom,
  nombre = EXCLUDED.nombre,
  uso = EXCLUDED.uso,
  referencia = EXCLUDED.referencia;

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
) ON CONFLICT (feature_id) DO UPDATE SET
  geom = EXCLUDED.geom,
  nombre = EXCLUDED.nombre,
  uso = EXCLUDED.uso,
  referencia = EXCLUDED.referencia;

INSERT INTO areas_verdes (
  feature_id, source_index, codigo, nombre, uso, referencia, perimetro_m, area_m2, geom, activo
) VALUES
  (
    'AV-FICT-0101', 910001, 'FICT-01', 'Prado Ficticio Institucional',
    'Uso Institucional', 'Demostración. No es un predio real.', 360, 8100,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08290 -12.07220, -77.08250 -12.07220, -77.08250 -12.07185, -77.08290 -12.07185, -77.08290 -12.07220))')), 4326),
    TRUE
  ),
  (
    'AV-FICT-0102', 910002, 'FICT-02', 'Patio Ficticio Administrativo',
    'Uso Administrativo', 'Demostración. No es un predio real.', 360, 8100,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08305 -12.07140, -77.08265 -12.07140, -77.08265 -12.07105, -77.08305 -12.07105, -77.08305 -12.07140))')), 4326),
    TRUE
  ),
  (
    'AV-FICT-0103', 910003, 'FICT-03', 'Descanso Ficticio Este',
    'Áreas de descanso y recreación', 'Demostración. No es un predio real.', 360, 8100,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08270 -12.07080, -77.08230 -12.07080, -77.08230 -12.07045, -77.08270 -12.07045, -77.08270 -12.07080))')), 4326),
    TRUE
  ),
  (
    'AV-FICT-0104', 910004, 'FICT-04', 'Manejo Ficticio Sur',
    'Manejo sostenible y reducción de consumo', 'Demostración. No es un predio real.', 360, 8100,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08310 -12.07000, -77.08270 -12.07000, -77.08270 -12.06965, -77.08310 -12.06965, -77.08310 -12.07000))')), 4326),
    TRUE
  ),
  (
    'AV-FICT-0105', 910005, 'FICT-05', 'Cancha Ficticia Oeste',
    'Áreas deportivas y recreación activa', 'Demostración. No es un predio real.', 360, 8100,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08320 -12.06920, -77.08280 -12.06920, -77.08280 -12.06885, -77.08320 -12.06885, -77.08320 -12.06920))')), 4326),
    TRUE
  ),
  (
    'AV-FICT-0106', 910006, 'FICT-06', 'Reserva Ficticia del Humedal',
    'Áreas de conservación', 'Demostración. No es un predio real.', 360, 8100,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08170 -12.06860, -77.08130 -12.06860, -77.08130 -12.06825, -77.08170 -12.06825, -77.08170 -12.06860))')), 4326),
    TRUE
  )
ON CONFLICT (feature_id) DO UPDATE SET
  geom = EXCLUDED.geom,
  nombre = EXCLUDED.nombre,
  uso = EXCLUDED.uso,
  referencia = EXCLUDED.referencia;

INSERT INTO poligonos_cuadrilla (
  feature_id, source_index, codigo, nombre, uso, referencia, perimetro_m, area_m2, geom, cuadrilla_id, sector, activo
) VALUES
  (
    'Z-FICT-0101', 910001, 'FICT-S1', 'Sector Ficticio Valeria',
    'Uso Institucional', 'Cuadrilla ficticia. Sin personas reales.', 400, 10000,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08290 -12.07220, -77.08250 -12.07220, -77.08250 -12.07185, -77.08290 -12.07185, -77.08290 -12.07220))')), 4326),
    'cap-norte', 'cua-valeria', TRUE
  ),
  (
    'Z-FICT-0102', 910002, 'FICT-S2', 'Sector Ficticio Mateo',
    'Uso Administrativo', 'Cuadrilla ficticia. Sin personas reales.', 400, 10000,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08305 -12.07140, -77.08265 -12.07140, -77.08265 -12.07105, -77.08305 -12.07105, -77.08305 -12.07140))')), 4326),
    'cap-norte', 'cua-mateo', TRUE
  ),
  (
    'Z-FICT-0103', 910003, 'FICT-S3', 'Sector Ficticio Renato',
    'Áreas de descanso y recreación', 'Cuadrilla ficticia. Sin personas reales.', 400, 10000,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08270 -12.07080, -77.08230 -12.07080, -77.08230 -12.07045, -77.08270 -12.07045, -77.08270 -12.07080))')), 4326),
    'cap-norte', 'cua-renato', TRUE
  ),
  (
    'Z-FICT-0104', 910004, 'FICT-S4', 'Sector Ficticio Deportivo',
    'Áreas deportivas y recreación activa', 'Rótulo de lugar. Sin personas reales.', 400, 10000,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08310 -12.07000, -77.08270 -12.07000, -77.08270 -12.06965, -77.08310 -12.06965, -77.08310 -12.07000))')), 4326),
    'cap-norte', 'campo-deportivo', TRUE
  ),
  (
    'Z-FICT-0105', 910005, 'FICT-S5', 'Sector Ficticio Bosque Húmedo',
    'Áreas de conservación', 'Rótulo de lugar. Sin personas reales.', 400, 10000,
    ST_SetSRID(ST_Multi(ST_GeomFromText('POLYGON((-77.08320 -12.06920, -77.08280 -12.06920, -77.08280 -12.06885, -77.08320 -12.06885, -77.08320 -12.06920))')), 4326),
    'cap-norte', 'bosque-humedo', TRUE
  )
ON CONFLICT (feature_id) DO UPDATE SET
  geom = EXCLUDED.geom,
  nombre = EXCLUDED.nombre,
  uso = EXCLUDED.uso,
  referencia = EXCLUDED.referencia;
