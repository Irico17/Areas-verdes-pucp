-- Sector operativo de cada polígono de cuadrilla: equivalente anonimizado del
-- color por grupo del mapa original (por jefe de grupo).
-- Aditiva. No trunca ni borra. Solo completa sector donde es NULL.
-- Los pares (source_index, sector) salen de data/v1/zonas_sector.json
-- (make sectores, luego go run ./cmd/sectores -sql). Sin nombres ni hashes.

INSERT INTO cuadrillas (id, nombre_ficticio, turno) VALUES
  ('cua-valeria', 'Valeria Quispe', 'manana'),
  ('cua-mateo', 'Mateo Salazar', 'manana'),
  ('cua-renato', 'Renato Cárdenas', 'tarde')
ON CONFLICT (id) DO NOTHING;

ALTER TABLE poligonos_cuadrilla ADD COLUMN IF NOT EXISTS sector TEXT;

ALTER TABLE poligonos_cuadrilla DROP CONSTRAINT IF EXISTS poligonos_sector_chk;
ALTER TABLE poligonos_cuadrilla ADD CONSTRAINT poligonos_sector_chk
  CHECK (sector IS NULL OR sector IN ('cua-valeria', 'cua-mateo', 'cua-renato', 'campo-deportivo', 'bosque-humedo'));

CREATE INDEX IF NOT EXISTS poligonos_cuadrilla_sector_idx ON poligonos_cuadrilla (sector);

COMMENT ON COLUMN poligonos_cuadrilla.sector IS
  'Sector operativo ficticio o rótulo de lugar. NULL = sin sector. No identifica personas.';

CREATE TABLE IF NOT EXISTS poligonos_sector_ref (
  source_index  INTEGER PRIMARY KEY,
  sector        TEXT NOT NULL,
  CONSTRAINT poligonos_sector_ref_chk
    CHECK (sector IN ('cua-valeria', 'cua-mateo', 'cua-renato', 'campo-deportivo', 'bosque-humedo'))
);

COMMENT ON TABLE poligonos_sector_ref IS
  'Sector por source_index de la fuente de polígonos. Sin FK: el TRUNCATE del ETL no la vacía.';

INSERT INTO poligonos_sector_ref (source_index, sector)
SELECT unnest('{2,4,12,13,14,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,100,101,102,103,104,105,106,107,108,109,110,111,112,113,114,115,116,117,118,119,120,121,122,123,124,125,126,127,128,129,130,131,132,133,134,135,136,182,183,191,193,204,205,206,207,208,231,255,270,275,276,277,278,279,280,281,288,292,293,294,295,305,306,307,333,334,335,336,337,338,339,340,341,342,343,344,345,346,347,348,349,350,351,352,353,354,355,356,384,386,387,388,389,390,391,392,393,394,395,398,399,400,401,402,403,404,405,406,407,408,409,410,412,422,423,424,425,426,427,428,429,430,431,432,433,434,435,436,437,438,439,440,441,442,443,444,445,446,447,450,451,452,453,455,456,457,458,459,460,461,462,463,464,465,466,467,468,484,485,486,487,488,489,490,491,495,496,497,503,504,505,506,507,508,509,510,514,515,518,519,520,521,522,529,530,532,533}'::int[]), 'cua-valeria'
ON CONFLICT (source_index) DO NOTHING;

INSERT INTO poligonos_sector_ref (source_index, sector)
SELECT unnest('{1,3,5,6,8,9,10,11,15,16,17,18,19,20,21,22,47,48,49,50,52,53,54,55,137,138,139,140,141,142,143,144,145,146,147,148,149,150,151,152,153,154,155,156,157,158,159,160,161,162,163,164,165,166,167,168,169,170,171,172,173,174,175,176,177,211,212,213,214,215,216,217,232,233,234,235,236,237,238,239,240,241,242,243,266,271,274,282,283,284,296,297,298,299,300,301,302,303,304,357,358,359,360,361,362,363,364,365,366,367,368,369,370,371,372,373,374,375,376,377,378,379,381,382,383,385,396,397,411,413,414,415,416,417,418,419,420,421,448,449,454,469,470,471,472,473,474,475,476,477,478,479,480,481,482,483,492,493,494,512,513,516,517,523,524,525,526,527}'::int[]), 'cua-mateo'
ON CONFLICT (source_index) DO NOTHING;

INSERT INTO poligonos_sector_ref (source_index, sector)
SELECT unnest('{7,51,56,178,179,180,181,184,185,186,187,188,189,190,192,194,195,196,197,198,199,200,201,202,203,209,210,218,219,220,221,222,223,224,225,226,227,228,229,230,244,245,246,247,248,249,250,251,252,253,254,256,257,258,259,260,261,262,263,264,265,267,268,269,272,273,285,286,287,289,290,291,308,309,310,311,312,313,314,315,316,317,318,319,320,321,322,323,324,325,326,327,328,329,330,331,332,380,498,499,500,511,528,531}'::int[]), 'cua-renato'
ON CONFLICT (source_index) DO NOTHING;

INSERT INTO poligonos_sector_ref (source_index, sector)
SELECT unnest('{501,502}'::int[]), 'campo-deportivo'
ON CONFLICT (source_index) DO NOTHING;

INSERT INTO poligonos_sector_ref (source_index, sector)
SELECT unnest('{0}'::int[]), 'bosque-humedo'
ON CONFLICT (source_index) DO NOTHING;

WITH hechos AS (
  UPDATE poligonos_cuadrilla p
  SET sector = r.sector, updated_at = now()
  FROM poligonos_sector_ref r
  WHERE p.source_index = r.source_index
    AND p.sector IS NULL
    AND p.feature_id IN (
      'Z-' || lpad((r.source_index + 1)::text, 4, '0'),
      'PC-' || lpad((r.source_index + 1)::text, 4, '0'))
  RETURNING p.id, p.sector
)
INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
SELECT 'poligonos_cuadrilla', id::text, 'edicion',
       jsonb_build_object('sector', NULL::text),
       jsonb_build_object('sector', sector, 'motivo', 'sector operativo desde data/v1/zonas_sector.json')
FROM hechos;

CREATE OR REPLACE VIEW zonas AS
SELECT
  id, feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act,
  referencia, perimetro_m, area_m2, geom, created_at, updated_at, sector
FROM poligonos_cuadrilla;
