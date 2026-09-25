package atencion

import (
	"errors"
	"testing"

	"campusverde/api/internal/operacion"
)

func TestRiegoExigeZona(t *testing.T) {
	if err := ValidarRiego("", "manana", 0); err == nil {
		t.Fatal("sin zona debía fallar")
	}
	if err := ValidarRiego("Z1", "manana", 12); err != nil {
		t.Fatal(err)
	}
	if err := ValidarRiego("sector norte", "manana", 0); err == nil {
		t.Fatal("un texto libre no es la FK")
	}
}

func TestCierreTercerizada(t *testing.T) {
	err := operacion.PuedeCerrar("tercerizada", false, true)
	var input operacion.InputError
	if !errors.As(err, &input) {
		t.Fatalf("se esperaba rechazo, obtuve %v", err)
	}
	if operacion.PuedeCerrar("tercerizada", true, true) != nil {
		t.Fatal("con orden y ejecución debía cerrar")
	}
	if operacion.PuedeCerrar("propia", true, false) == nil {
		t.Fatal("sin ejecución no cierra")
	}
}
