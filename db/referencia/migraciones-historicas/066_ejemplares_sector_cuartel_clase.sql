-- Solo se acepta un ítem activo de sector de capataz o de cuartel.
-- El resto de columnas del ejemplar no dispara esta regla, así el upsert
-- por origen_ref sigue igual.

CREATE OR REPLACE FUNCTION ejemplares_sector_cuartel_clase()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF NEW.sector_cuartel_id IS NULL THEN
    RETURN NEW;
  END IF;
  IF NOT EXISTS (
    SELECT 1
    FROM catalogos c
    WHERE c.id = NEW.sector_cuartel_id
      AND c.activo
      AND c.clase IN ('sector_capataz', 'cuartel')
  ) THEN
    RAISE EXCEPTION 'sector o cuartel desconocido'
      USING ERRCODE = '23514';
  END IF;
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS ejemplares_sector_cuartel_clase_trg ON ejemplares;

CREATE TRIGGER ejemplares_sector_cuartel_clase_trg
  BEFORE INSERT OR UPDATE OF sector_cuartel_id ON ejemplares
  FOR EACH ROW
  EXECUTE PROCEDURE ejemplares_sector_cuartel_clase();

COMMENT ON FUNCTION ejemplares_sector_cuartel_clase() IS
  'Rechaza un sector_cuartel_id que no sea sector de capataz o cuartel activo.';
