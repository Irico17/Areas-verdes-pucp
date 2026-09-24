package migrate

import (
	"strings"
	"testing"
)

func TestSplitSQLRespetaFuncion(t *testing.T) {
	src := `
-- comentario; no es sentencia
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE OR REPLACE FUNCTION catastro_geom_4326(geojson text)
RETURNS geometry
LANGUAGE plpgsql
AS $$
BEGIN
  IF geojson IS NULL THEN
    RETURN NULL;
  END IF;
  RETURN NULL;
END;
$$;

CREATE TABLE IF NOT EXISTS zonas (id bigint);
`
	parts := splitSQL(src)
	if len(parts) != 3 {
		t.Fatalf("sentencias = %d: %#v", len(parts), parts)
	}
	if !strings.Contains(parts[1], "RETURN NULL;") {
		t.Fatalf("la función se partió: %s", parts[1])
	}
	if strings.Contains(parts[0], "comentario") {
		t.Fatalf("el comentario entró en la sentencia: %s", parts[0])
	}
}
