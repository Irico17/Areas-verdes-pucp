-- Catálogos del catastro. No inventa usos de área: esos se dan de alta con la carga.

INSERT INTO catalogos (clase, codigo, nombre, orden) VALUES
  ('tipo_vegetacion', 'arbol', 'Árbol', 1),
  ('tipo_vegetacion', 'palmera', 'Palmera', 2),
  ('tipo_vegetacion', 'arbusto', 'Arbusto', 3),
  ('tipo_vegetacion', 'herbacea', 'Herbácea', 4),
  ('tipo_vegetacion', 'trepadora', 'Trepadora', 5),
  ('tipo_vegetacion', 'suculenta', 'Suculenta', 6),
  ('tipo_vegetacion', 'cafeto', 'cafeto', 7)
ON CONFLICT (clase, codigo) DO NOTHING;
