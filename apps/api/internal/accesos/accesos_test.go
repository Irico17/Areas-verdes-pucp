package accesos

import "testing"

func TestPermiteCapatazNoAdministraCatalogos(t *testing.T) {
	if Permite("capataz", "catalogos") {
		t.Fatal("el capataz no administra catálogos")
	}
	if !Permite("admin", "catalogos") {
		t.Fatal("admin sí administra catálogos")
	}
	if !Permite("capataz", "registrar") {
		t.Fatal("el capataz registra en campo")
	}
	if Permite("jefatura", "registrar") {
		t.Fatal("jefatura no da de alta labores en esta matriz")
	}
}
