package enums_test

import (
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
)

func TestRolEnumValoresYMetodos(t *testing.T) {
	tests := []struct {
		rol      enums.Rol
		expected string
		valido   bool
	}{
		{enums.RolCapataz, "capataz", true},
		{enums.RolCoordinacion, "coordinacion", true},
		{enums.RolJefatura, "jefatura", true},
		{enums.RolAdmin, "admin", true},
		{enums.Rol("operario"), "operario", false},
		{enums.Rol("desconocido"), "desconocido", false},
		{enums.Rol(""), "", false},
	}

	for _, tt := range tests {
		if tt.rol.String() != tt.expected {
			t.Errorf("Rol.String() = %q, esperado %q", tt.rol.String(), tt.expected)
		}
		if tt.rol.EsValido() != tt.valido {
			t.Errorf("Rol.EsValido() para %q = %v, esperado %v", tt.rol, tt.rol.EsValido(), tt.valido)
		}
	}
}

func TestRolesValidos(t *testing.T) {
	roles := enums.RolesValidos()
	if len(roles) != 4 {
		t.Fatalf("RolesValidos() len = %d, esperado 4", len(roles))
	}

	encontrados := make(map[enums.Rol]bool)
	for _, r := range roles {
		encontrados[r] = true
		if !r.EsValido() {
			t.Errorf("rol %q devuelto por RolesValidos no es válido", r)
		}
	}

	esperados := []enums.Rol{
		enums.RolCapataz,
		enums.RolCoordinacion,
		enums.RolJefatura,
		enums.RolAdmin,
	}
	for _, exp := range esperados {
		if !encontrados[exp] {
			t.Errorf("falta rol esperado %q en RolesValidos()", exp)
		}
	}
}
