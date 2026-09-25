package catastro

import "testing"

func TestNormalizarNombre(t *testing.T) {
	if got := NormalizarNombre("  Café  Tostado  "); got != "cafe tostado" {
		t.Fatalf("normalizar = %q", got)
	}
	if got := NormalizarNombre("Árbol"); got != "arbol" {
		t.Fatalf("arbol = %q", got)
	}
}

func TestValidacionesDeCatastro(t *testing.T) {
	if err := ValidarCodigoZona("Z1"); err != nil {
		t.Fatal(err)
	}
	if err := ValidarCodigoZona("Z5"); err != ErrEntrada {
		t.Fatalf("Z5: %v", err)
	}
	if err := ValidarPuntoCampus(-12.07, -77.08); err != nil {
		t.Fatal(err)
	}
	if err := ValidarPuntoCampus(-1.20, -77.08); err != ErrEntrada {
		t.Fatalf("latitud fuera: %v", err)
	}
	if err := ValidarTipoVegetacion("Palmera"); err != nil {
		t.Fatal(err)
	}
	if err := ValidarTipoVegetacion("helecho"); err != ErrEntrada {
		t.Fatalf("tipo: %v", err)
	}
	if err := ValidarCantidad(0); err != ErrEntrada {
		t.Fatal(err)
	}
	if err := ValidarReferencia(string(make([]rune, 501))); err != ErrEntrada {
		t.Fatal("referencia larga")
	}
}
