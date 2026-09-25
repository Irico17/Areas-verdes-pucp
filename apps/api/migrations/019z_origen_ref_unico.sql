-- Upsert del frente 1E. origen_ref ya existía en 010-015, sin índice único.
-- PostgreSQL admite varios NULL; las filas del lote siempre traen origen_ref.
-- Si un valor no nulo está repetido, el menor id lo conserva y el resto lleva « ·<id>».

DO $$
DECLARE
  r RECORD;
  nuevo TEXT;
  sufijo INT;
  ocupado BOOLEAN;
  tbl TEXT;
BEGIN
  FOREACH tbl IN ARRAY ARRAY['zonas_supervision', 'poligonos_cuadrilla', 'lugares', 'ejemplares']
  LOOP
    FOR r IN EXECUTE format(
      'SELECT id, origen_ref FROM %I a
       WHERE origen_ref IS NOT NULL
         AND id <> (SELECT min(id) FROM %I b WHERE b.origen_ref = a.origen_ref)
       ORDER BY id', tbl, tbl)
    LOOP
      sufijo := 0;
      LOOP
        IF sufijo = 0 THEN
          nuevo := r.origen_ref || ' ·' || r.id::text;
        ELSE
          nuevo := r.origen_ref || ' ·' || r.id::text || '-' || sufijo::text;
        END IF;
        EXECUTE format(
          'SELECT EXISTS (SELECT 1 FROM %I WHERE origen_ref = $1 AND id <> $2)', tbl)
          INTO ocupado
          USING nuevo, r.id;
        EXIT WHEN NOT ocupado;
        sufijo := sufijo + 1;
      END LOOP;
      INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
      VALUES (
        tbl, r.id::text, 'edicion',
        jsonb_build_object('origen_ref', r.origen_ref),
        jsonb_build_object(
          'origen_ref', nuevo,
          'motivo', 'origen_ref duplicado; se conservo el de menor id. Corrija el identificador desde la ficha'
        )
      );
      EXECUTE format('UPDATE %I SET origen_ref = $1, updated_at = now() WHERE id = $2', tbl)
        USING nuevo, r.id;
    END LOOP;
  END LOOP;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS zonas_supervision_origen_ref_uidx
  ON zonas_supervision (origen_ref);

CREATE UNIQUE INDEX IF NOT EXISTS poligonos_cuadrilla_origen_ref_uidx
  ON poligonos_cuadrilla (origen_ref);

CREATE UNIQUE INDEX IF NOT EXISTS lugares_origen_ref_uidx
  ON lugares (origen_ref);

CREATE UNIQUE INDEX IF NOT EXISTS ejemplares_origen_ref_uidx
  ON ejemplares (origen_ref);
