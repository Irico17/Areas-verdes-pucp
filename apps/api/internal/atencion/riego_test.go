package atencion

import "strings"
import "testing"

func TestConsultaRiegoSoloElEquipoDelCapataz(t *testing.T) {
	q, args := consultaRiego("cap-norte")
	if !strings.Contains(q, "WHERE r.capataz_id = $1") {
		t.Fatalf("el capataz debe filtrar por equipo: %s", q)
	}
	if len(args) != 1 || args[0] != "cap-norte" {
		t.Fatalf("argumento: %#v", args)
	}
}

func TestConsultaRiegoOficinaVeTodos(t *testing.T) {
	q, args := consultaRiego("  ")
	if strings.Contains(q, "WHERE") {
		t.Fatalf("oficina no debe filtrar: %s", q)
	}
	if len(args) != 0 {
		t.Fatalf("sin argumentos: %#v", args)
	}
}
