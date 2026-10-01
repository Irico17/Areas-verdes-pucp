package usecases_test

import (
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
)

func TestConsultaRiegoSoloElEquipoDelCapataz(t *testing.T) {
	q, args := usecases.ConsultaRiego("cap-norte")
	if !strings.Contains(q, "WHERE r.capataz_id = $1") {
		t.Fatalf("el capataz debe filtrar por equipo: %s", q)
	}
	if len(args) != 1 || args[0] != "cap-norte" {
		t.Fatalf("argumento: %#v", args)
	}
}

func TestConsultaRiegoOficinaVeTodos(t *testing.T) {
	q, args := usecases.ConsultaRiego("  ")
	if strings.Contains(q, "WHERE") {
		t.Fatalf("oficina no debe filtrar: %s", q)
	}
	if len(args) != 0 {
		t.Fatalf("sin argumentos: %#v", args)
	}
}
