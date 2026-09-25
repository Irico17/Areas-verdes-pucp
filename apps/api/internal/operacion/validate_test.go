package operacion

import (
	"errors"
	"testing"
)

func TestValidateCreateCapatazProhibido(t *testing.T) {
	err := ValidateCreate(CreateInput{ActorRol: RolCapataz, ID: "11111111-1111-4111-8111-111111111111", Tipo: "riego", Titulo: "Riego", Lon: -77.08, Lat: -12.07})
	if !errors.Is(err, ErrProhibido) {
		t.Fatalf("esperado prohibido, obtuve %v", err)
	}
}

func TestValidateQueryCapatazSinEquipo(t *testing.T) {
	err := ValidateQuery(Query{Rol: RolCapataz, SoloAbiertas: true})
	if !errors.Is(err, ErrValidacion) {
		t.Fatalf("esperado validacion, obtuve %v", err)
	}
}

func TestSamePayload(t *testing.T) {
	in := CreateInput{Tipo: "poda", Titulo: " Poda ", Detalle: "borde", Lon: -77.0808, Lat: -12.0704, AssignedCapatazID: "cap-sur"}
	saved := Saved{Tipo: "poda", Titulo: "Poda", Detalle: "borde", Lon: -77.0808, Lat: -12.0704, AssignedCapatazID: "cap-sur"}
	if !SamePayload(saved, in) {
		t.Fatal("debía coincidir")
	}
	saved.Titulo = "Otra"
	if SamePayload(saved, in) {
		t.Fatal("no debía coincidir")
	}
}

func TestAltaPorLugarSinPin(t *testing.T) {
	err := ValidateCreate(CreateInput{
		ActorRol: RolCoordinacion,
		ID:       "11111111-1111-4111-8111-111111111111",
		Tipo:     "riego",
		Titulo:   "Por lugar",
		LugarID:  "lugar-1",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPuedeCerrarTercerizada(t *testing.T) {
	if PuedeCerrar("tercerizada", false, true) == nil {
		t.Fatal("sin orden no cierra")
	}
	if PuedeCerrar("propia", false, true) != nil {
		t.Fatal("la propia con ejecución sí cierra")
	}
}

func TestPuntoFuera(t *testing.T) {
	err := ValidateCreate(CreateInput{
		ActorRol: RolCoordinacion,
		ID:       "11111111-1111-4111-8111-111111111111",
		Tipo:     "riego",
		Titulo:   "Lejos",
		Lon:      -70,
		Lat:      -12.07,
	})
	var input InputError
	if !errors.As(err, &input) {
		t.Fatalf("esperado input error, obtuve %v", err)
	}
}
