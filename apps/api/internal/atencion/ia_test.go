package atencion

import "testing"

func TestSugerirRiego(t *testing.T) {
	s := SugerirTipo("Revisar aspersores del eje")
	if s.Codigo != "riego" {
		t.Fatalf("código %q", s.Codigo)
	}
	if !s.RequiereHumano {
		t.Fatal("la sugerencia no se aplica sola")
	}
}

func TestSugerirSinPista(t *testing.T) {
	s := SugerirTipo("Turno de la mañana")
	if s.Codigo != "" {
		t.Fatalf("no debía sugerir %q", s.Codigo)
	}
	if !s.RequiereHumano {
		t.Fatal("sin pista también exige a una persona")
	}
}
