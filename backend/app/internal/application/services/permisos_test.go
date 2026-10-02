package services_test

import (
	"context"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/services"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
)

func TestPermiteCapatazNoAdministraCatalogos(t *testing.T) {
	svc := services.NewPermisosMemoria()

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
	if !svc.Permite(enums.RolJefatura.String(), "usuarios") {
		t.Fatal("jefatura administra las cuentas")
	}
	if svc.Permite(enums.RolCapataz.String(), "usuarios") {
		t.Fatal("el capataz no administra cuentas")
	}
}

func TestPermiteRecargaSinReiniciar(t *testing.T) {
	repo := services.NuevaMemoriaPermisos(services.MatrizPermisos)
	svc := services.NewPermisosService(repo)
	if !svc.Permite(enums.RolCapataz.String(), "registrar") {
		t.Fatal("la semilla deja registrar al capataz")
	}
	if err := repo.Establecer(context.Background(), enums.RolCapataz.String(), "registrar", false); err != nil {
		t.Fatal(err)
	}
	if svc.Permite(enums.RolCapataz.String(), "registrar") {
		t.Fatal("Permite debe ver el cambio sin crear otro servicio")
	}
}
