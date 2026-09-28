package services_test

import (
	"context"
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/services"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type mockCambioRepository struct {
	creados []*entities.Cambio
}

func (m *mockCambioRepository) Crear(_ context.Context, cambio *entities.Cambio) error {
	m.creados = append(m.creados, cambio)
	return nil
}

func TestAuditoriaService_RegistrarCambio(t *testing.T) {
	repo := &mockCambioRepository{}
	svc := services.NewAuditoriaService(repo)

	var uid int64 = 42
	err := svc.RegistrarCambio(context.Background(), dto.RegistrarCambioDTO{
		Entidad:   "areas_verdes",
		EntidadID: "AV-0001",
		Accion:    "edicion",
		Antes:     map[string]string{"nombre": "Viejo"},
		Despues:   map[string]string{"nombre": "Nuevo"},
		UsuarioID: &uid,
	})
	if err != nil {
		t.Fatalf("error registrando cambio: %v", err)
	}

	if len(repo.creados) != 1 {
		t.Fatalf("se esperaba 1 cambio creado, obtenido: %d", len(repo.creados))
	}
	c := repo.creados[0]
	if c.Entidad != "areas_verdes" || c.EntidadID != "AV-0001" || c.Accion != "edicion" {
		t.Fatalf("datos incorrectos en cambio: %+v", c)
	}
	if c.UsuarioID == nil || *c.UsuarioID != 42 {
		t.Fatalf("usuario_id esperado 42, obtenido: %v", c.UsuarioID)
	}
	if c.Antes == nil || !strings.Contains(*c.Antes, "Viejo") {
		t.Fatalf("antes inesperado: %v", c.Antes)
	}
	if c.Despues == nil || !strings.Contains(*c.Despues, "Nuevo") {
		t.Fatalf("despues inesperado: %v", c.Despues)
	}
}
