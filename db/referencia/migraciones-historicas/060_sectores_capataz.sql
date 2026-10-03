-- Catálogo de sector de capataz. Sin CHECK de códigos: el alta no queda
-- atrapada en una lista fija. No borra polígonos ni lugar_libre.
-- Los nombres son ficticios, los mismos que ya colorea el catastro.

CREATE TABLE IF NOT EXISTS sectores_capataz (
  id          BIGSERIAL PRIMARY KEY,
  codigo      TEXT NOT NULL,
  nombre      TEXT NOT NULL,
  color       TEXT NOT NULL,
  activo      BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS sectores_capataz_codigo_uidx
  ON sectores_capataz (codigo);

COMMENT ON TABLE sectores_capataz IS
  'Catálogo editable del sector de capataz. El color del mapa sale de aquí. activo = false es la baja y no se borra la fila.';

COMMENT ON COLUMN sectores_capataz.color IS
  'Hexadecimal #rrggbb que pinta el polígono. No es una lista cerrada en código.';

-- El CHECK de 044 fijaba cinco códigos. Se retira para que un sector nuevo
-- pueda guardarse en el polígono. Los valores ya cargados no se tocan.
ALTER TABLE poligonos_cuadrilla DROP CONSTRAINT IF EXISTS poligonos_sector_chk;

DO $$
DECLARE
  v RECORD;
  new_id BIGINT;
BEGIN
  FOR v IN
    SELECT *
    FROM (VALUES
      ('cua-valeria', 'Sector de capataz — Valeria Quispe (ficticio)', '#6b5596'),
      ('cua-mateo', 'Sector de capataz — Mateo Salazar (ficticio)', '#3f73b0'),
      ('cua-renato', 'Sector de capataz — Renato Cárdenas (ficticio)', '#c27c2c'),
      ('campo-deportivo', 'Campo deportivo', '#9aab3e'),
      ('bosque-humedo', 'Bosque húmedo', '#3f9a82')
    ) AS t(codigo, nombre, color)
  LOOP
    IF NOT EXISTS (
      SELECT 1 FROM sectores_capataz s WHERE s.codigo = v.codigo
    ) THEN
      INSERT INTO sectores_capataz (codigo, nombre, color, activo)
      VALUES (v.codigo, v.nombre, v.color, TRUE)
      RETURNING id INTO new_id;

      INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
      VALUES (
        'sectores_capataz',
        new_id::text,
        'alta',
        NULL,
        jsonb_build_object(
          'codigo', v.codigo,
          'nombre', v.nombre,
          'color', v.color,
          'activo', TRUE
        )
      );
    END IF;
  END LOOP;
END $$;
