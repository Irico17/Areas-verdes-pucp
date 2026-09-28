package services_test

import (
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/services"
)

func TestPermiteCapatazNoAdministraCatalogos(t *testing.T) {
	svc := services.NewPermisosService()

	if svc.Permite("capataz", "catalogos") {
		t.Fatal("el capataz no administra catálogos")
	}
	if !svc.Permite("admin", "catalogos") {
		t.Fatal("admin sí administra catálogos")
	}
	if !svc.Permite("capataz", "registrar") {
		t.Fatal("el capataz registra en campo")
	}
	if svc.Permite("jefatura", "registrar") {
		t.Fatal("jefatura no da de alta labores en esta matriz")
	}
	if !svc.Permite("jefatura", "evidencias") {
		t.Fatal("jefatura debe tener permiso para registrar evidencias")
	}
}

func TestPermiteAlguno(t *testing.T) {
	svc := services.NewPermisosService()

	if !svc.PermiteAlguno("capataz", "catalogos", "registrar") {
		t.Fatal("capataz debe permitir al menos registrar")
	}
	if svc.PermiteAlguno("capataz", "catalogos", "validar") {
		t.Fatal("capataz no tiene ni catalogos ni validar")
	}
}
