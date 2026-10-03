-- RF-31. Amplía el CHECK de tipos de evento. Los siete anteriores se conservan.
-- Idempotente: se suelta la restricción y se vuelve a crear. No borra filas.

ALTER TABLE actividad_eventos DROP CONSTRAINT IF EXISTS actividad_eventos_tipo_chk;

ALTER TABLE actividad_eventos ADD CONSTRAINT actividad_eventos_tipo_chk CHECK (
  tipo IN (
    'creada',
    'asignada',
    'reasignada',
    'estado',
    'cancelada',
    'archivada',
    'evidencia',
    'inicio',
    'supervision',
    'derivacion',
    'observacion',
    'conformidad',
    'avance'
  )
);

COMMENT ON CONSTRAINT actividad_eventos_tipo_chk ON actividad_eventos IS
  'Cadena de la actividad. Los siete tipos originales siguen valiendo y se suman los hitos del libro y el avance.';
