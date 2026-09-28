package entities_test

import (
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

func TestNormalizarNombre(t *testing.T) {
	if got := entities.NormalizarNombre("  Café  Tostado  "); got != "cafe tostado" {
		t.Fatalf("normalizar = %q", got)
	}
	if got := entities.NormalizarNombre("Árbol"); got != "arbol" {
		t.Fatalf("arbol = %q", got)
	}
}

func TestValidacionesDeCatastro(t *testing.T) {
	if err := entities.ValidarCodigoZona("Z1"); err != nil {
		t.Fatal(err)
	}
	if err := entities.ValidarCodigoZona("Z5"); err != domainErrors.ErrEntrada {
		t.Fatalf("Z5: %v", err)
	}
	if err := entities.ValidarPuntoCampus(-12.07, -77.08); err != nil {
		t.Fatal(err)
	}
	if err := entities.ValidarPuntoCampus(-1.20, -77.08); err != domainErrors.ErrEntrada {
		t.Fatalf("latitud fuera: %v", err)
	}
	if err := enums.ValidarTipoVegetacion("Palmera"); err != nil {
		t.Fatal(err)
	}
	if err := enums.ValidarTipoVegetacion("helecho"); err != domainErrors.ErrEntrada {
		t.Fatalf("tipo: %v", err)
	}
	if err := entities.ValidarCantidad(0); err != domainErrors.ErrEntrada {
		t.Fatal(err)
	}
	if err := entities.ValidarReferencia(string(make([]rune, 501))); err != domainErrors.ErrEntrada {
		t.Fatal("referencia larga")
	}
}
