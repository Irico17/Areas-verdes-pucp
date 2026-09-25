package capas

import (
	"encoding/json"
	"testing"
)

func TestClavesPatchNoMarcaLoOmitido(t *testing.T) {
	claves := clavesPatch(json.RawMessage(`{"nota":"junto al comedor"}`))
	if !claves["nota"] || claves["codigo"] || claves["_todo"] {
		t.Fatalf("claves %#v", claves)
	}
	if !clavesPatch(nil)["_todo"] {
		t.Fatal("sin cuerpo se actualiza el registro completo")
	}
}
