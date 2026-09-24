-- Limpia filas de prueba y deja una orden de servicio estable para la demo.
-- Idempotente: se puede correr sobre una base que ya tuvo capturas manuales.

DELETE FROM evidencias
WHERE actividad_id IN (SELECT id FROM actividades WHERE titulo = 'Prueba de pin');

DELETE FROM ordenes_servicio
WHERE actividad_id IN (SELECT id FROM actividades WHERE titulo = 'Prueba de pin');

UPDATE solicitudes
SET actividad_id = NULL
WHERE actividad_id IN (SELECT id FROM actividades WHERE titulo = 'Prueba de pin');

DELETE FROM actividad_eventos
WHERE actividad_id IN (SELECT id FROM actividades WHERE titulo = 'Prueba de pin');

DELETE FROM actividades WHERE titulo = 'Prueba de pin';

INSERT INTO actividades (
  id, tipo, estado, titulo, detalle, assigned_capataz_id, geom, ejecutor
) VALUES (
  '99999999-9999-4999-8999-999999999999',
  'poda', 'pendiente',
  'Poda contratada del borde sur',
  'Servicio externo de setos. No se cierra sin una orden.',
  'cap-sur',
  ST_SetSRID(ST_MakePoint(-77.08120, -12.07080), 4326),
  'tercerizada'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO actividad_eventos (actividad_id, tipo, estado, capataz_id, actor_rol, nota)
SELECT '99999999-9999-4999-8999-999999999999', 'creada', 'pendiente', 'cap-sur', 'coordinacion', 'Alta de demostración'
WHERE NOT EXISTS (
  SELECT 1 FROM actividad_eventos
  WHERE actividad_id = '99999999-9999-4999-8999-999999999999' AND tipo = 'creada'
);

INSERT INTO ordenes_servicio (id, actividad_id, empresa, referencia, frecuencia, estado)
VALUES (
  '99999999-9999-4999-8999-999999999991',
  '99999999-9999-4999-8999-999999999999',
  'Jardines del Rímac S.A.C.',
  'OS-2026-018',
  'quincenal',
  'en_proceso'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO ordenes_servicio (id, actividad_id, empresa, referencia, frecuencia, estado)
SELECT
  '99999999-9999-4999-8999-999999999992',
  a.id,
  'Jardines del Rímac S.A.C.',
  'OS-2026-017',
  'quincenal',
  'en_proceso'
FROM actividades a
WHERE a.titulo = 'Poda contratada'
  AND a.ejecutor = 'tercerizada'
  AND NOT EXISTS (SELECT 1 FROM ordenes_servicio o WHERE o.actividad_id = a.id)
LIMIT 1
ON CONFLICT (id) DO NOTHING;

INSERT INTO riego_registros (id, sector, turno, capataz_id, fecha, nota) VALUES
  ('88888888-8888-4888-8888-888888888881', 'Eje central', 'manana', 'cap-norte', DATE '2026-09-24', 'Aspersores del eje.'),
  ('88888888-8888-4888-8888-888888888882', 'Bosque húmedo', 'tarde', 'cap-riego', DATE '2026-09-24', 'Ronda de la tarde.')
ON CONFLICT (id) DO NOTHING;
