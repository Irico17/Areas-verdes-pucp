package auditoria

import (
	"errors"
	"testing"
)

func TestAltaDeLaborExigeTituloYEstado(t *testing.T) {
	_, _, err := camposAltaActividad(snap{Titulo: "", Estado: "pendiente"})
	if err == nil || !errors.As(err, new(InputError)) {
		t.Fatalf("sin título: %v", err)
	}
	_, _, err = camposAltaActividad(snap{Titulo: "Poda", Estado: "  "})
	if err == nil {
		t.Fatal("sin estado debía rechazarse")
	}
	titulo, estado, err := camposAltaActividad(snap{Titulo: " Poda ", Estado: "cerrada"})
	if err != nil || titulo != "Poda" || estado != "cerrada" {
		t.Fatalf("alta válida: %q %q %v", titulo, estado, err)
	}
}
