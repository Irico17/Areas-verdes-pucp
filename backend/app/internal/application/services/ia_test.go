package services_test

import (
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/services"
)

func TestSugerirRiego(t *testing.T) {
	svc := services.NewSugeridorTipoService()
	s := svc.SugerirTipo("Revisar aspersores del eje")
	if s.Codigo != "riego" {
		t.Fatalf("código %q", s.Codigo)
	}
	if !s.RequiereHumano {
		t.Fatal("la sugerencia no se aplica sola")
	}
}

func TestSugerirSinPista(t *testing.T) {
	svc := services.NewSugeridorTipoService()
	s := svc.SugerirTipo("Turno de la mañana")
	if s.Codigo != "" {
		t.Fatalf("no debía sugerir %q", s.Codigo)
	}
	if !s.RequiereHumano {
		t.Fatal("sin pista también exige a una persona")
	}
}
