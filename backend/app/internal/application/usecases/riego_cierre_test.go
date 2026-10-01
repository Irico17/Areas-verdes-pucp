package usecases_test

import (
	"errors"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

func TestRiegoExigeZona(t *testing.T) {
	if err := usecases.ValidarRiego("", "manana", 0); err == nil {
		t.Fatal("sin zona debía fallar")
	}
	if err := usecases.ValidarRiego("Z1", "manana", 12); err != nil {
		t.Fatal(err)
	}
	if err := usecases.ValidarRiego("sector norte", "manana", 0); err == nil {
		t.Fatal("un texto libre no es la FK")
	}
}

func TestCierreTercerizada(t *testing.T) {
	err := usecases.PuedeCerrar("tercerizada", false, true)
	var input domainErrors.InputError
	if !errors.As(err, &input) {
		t.Fatalf("se esperaba rechazo, obtuve %v", err)
	}
	if usecases.PuedeCerrar("tercerizada", true, true) != nil {
		t.Fatal("con orden y ejecución debía cerrar")
	}
	if usecases.PuedeCerrar("propia", true, false) == nil {
		t.Fatal("sin ejecución no cierra")
	}
}
