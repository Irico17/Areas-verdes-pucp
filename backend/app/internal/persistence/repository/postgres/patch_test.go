package postgres_test

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
)

func TestClavesPatchNoMarcaLoOmitido(t *testing.T) {
	claves, doc, err := postgres.ClavesPatch(json.RawMessage(`{"nota":"junto al comedor"}`))
	if err != nil {
		t.Fatal(err)
	}
	if !claves["nota"] || claves["codigo"] || claves["_todo"] || strings.Contains(doc, "_todo") {
		t.Fatalf("claves %#v doc %s", claves, doc)
	}
	for _, cuerpo := range []string{"", "null", "{}", `{"_todo":false}`, `{"_todo":true}`} {
		if _, _, err := postgres.ClavesPatch(json.RawMessage(cuerpo)); !errors.Is(err, domainErrors.ErrPatchVacio) {
			t.Fatalf("cuerpo %q debía ser patch vacío, fue %v", cuerpo, err)
		}
	}
	claves, doc, err = postgres.ClavesPatch(json.RawMessage(`{"_todo":true,"lat":-12.07}`))
	if err != nil {
		t.Fatal(err)
	}
	if claves["_todo"] || !claves["lat"] || strings.Contains(doc, "_todo") {
		t.Fatalf("el cliente no puede activar _todo: %#v %s", claves, doc)
	}
}

func TestGeomParcialUsaLaCoordenadaGuardada(t *testing.T) {
	body, err := os.ReadFile("inventario_campo.repository.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(body)
	if strings.Contains(src, "? '_todo'") {
		t.Fatal("el SQL no puede obedecer _todo enviado por el cliente")
	}
	for _, pieza := range []string{
		"CASE WHEN $23::jsonb ? 'lon' THEN $3 ELSE lon END",
		"CASE WHEN $23::jsonb ? 'lat' THEN $2 ELSE lat END",
		"CASE WHEN $8::jsonb ? 'lon' THEN $6 ELSE lon END",
		"CASE WHEN $6::jsonb ? 'lat' THEN $2 ELSE lat END",
	} {
		if !strings.Contains(src, pieza) {
			t.Fatalf("falta coordenada guardada: %s", pieza)
		}
	}
}
