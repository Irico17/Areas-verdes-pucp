-- Semilla ficticia de develop y qa.
-- El entrypoint la aplica solo si el catastro está vacío.
-- No es un volcado de producción ni una copia sin enmascarar.
-- Nombres de demostración: Jardín Ficticio Norte, Sector Ficticio Norte, Equipo Norte.

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
