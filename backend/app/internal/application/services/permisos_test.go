package services_test

import (
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/services"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
)

func TestPermiteCapatazNoAdministraCatalogos(t *testing.T) {
	svc := services.NewPermisosService()

	if svc.Permite(enums.RolCapataz.String(), "catalogos") {
		t.Fatal("el capataz no administra catálogos")
	}
	if !svc.Permite(enums.RolAdmin.String(), "catalogos") {
		t.Fatal("admin sí administra catálogos")
	}
	if !svc.Permite(enums.RolCapataz.String(), "registrar") {
		t.Fatal("el capataz registra en campo")
	}
	if svc.Permite(enums.RolJefatura.String(), "registrar") {
		t.Fatal("jefatura no da de alta labores en esta matriz")
	}
	if !svc.Permite(enums.RolJefatura.String(), "evidencias") {
		t.Fatal("jefatura debe tener permiso para registrar evidencias")
	}
}

func TestPermiteAlguno(t *testing.T) {
	svc := services.NewPermisosService()

	if !svc.PermiteAlguno(enums.RolCapataz.String(), "catalogos", "registrar") {
		t.Fatal("capataz debe permitir al menos registrar")
	}
	if svc.PermiteAlguno(enums.RolCapataz.String(), "catalogos", "validar") {
		t.Fatal("capataz no tiene ni catalogos ni validar")
	}
}
