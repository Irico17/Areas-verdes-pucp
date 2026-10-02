package enums_test

import (
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
)

func TestTransicionEstado_NoSaltaElCierre(t *testing.T) {
	if enums.TransicionEstadoPermitida("pendiente", "cerrada") {
		t.Fatal("por iniciar no puede cerrarse saltando la tabla")
	}
	if enums.TransicionEstadoPermitida("pendiente", "ejecutado") {
		t.Fatal("por iniciar no salta a ejecutado")
	}
	pasos := [][2]string{
		{"pendiente", "en_proceso"},
		{"en_proceso", "ejecutado"},
		{"ejecutado", "cerrada"},
		{"pendiente", "cancelada"},
		{"en_proceso", "archivada"},
		{"bloqueada", "en_proceso"},
	}
	for _, paso := range pasos {
		if !enums.TransicionEstadoPermitida(paso[0], paso[1]) {
			t.Fatalf("se esperaba permitir %s → %s", paso[0], paso[1])
		}
	}
	if enums.TransicionEstadoPermitida("pendiente", "bloqueada") {
		t.Fatal("bloqueada no es un destino nuevo")
	}
	if !enums.TransicionEstadoPermitida("en_proceso", "en_proceso") {
		t.Fatal("el mismo estado no es un salto")
	}
}
