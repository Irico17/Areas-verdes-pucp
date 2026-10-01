package usecases_test

import (
	"context"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type mockViveroRepo struct {
	registros []*entities.Vivero
	guardado  *entities.Vivero
	archivado string
	mesListar string
	err       error
}

func (m *mockViveroRepo) Listar(_ context.Context, mes string) ([]*entities.Vivero, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.mesListar = mes
	return m.registros, nil
}

func (m *mockViveroRepo) Guardar(_ context.Context, in dto.GuardarViveroDTO) (*entities.Vivero, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.guardado = &entities.Vivero{
		ID:            in.ID,
		Fecha:         &in.Fecha,
		Area:          in.Area,
		Subproceso:    in.Subproceso,
		Etapa:         in.Etapa,
		Descripcion:   in.Descripcion,
		Observaciones: in.Observaciones,
		Responsables:  in.Responsables,
		LugarID:       &in.LugarID,
		LugarLibre:    in.LugarLibre,
	}
	return m.guardado, nil
}

func (m *mockViveroRepo) Archivar(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	m.archivado = id
	return nil
}

func TestViveroUseCase_Validaciones(t *testing.T) {
	ctx := context.Background()
	repo := &mockViveroRepo{}
	uc := usecases.NewViveroUseCase(repo)

	// 1. Listar con mes inválido
	_, err := uc.Listar(ctx, "202605")
	if err == nil || err.Error() != "el mes usa AAAA-MM" {
		t.Fatalf("esperado error de mes, obtenido: %v", err)
	}
	_, err = uc.Listar(ctx, "invalido")
	if err == nil || err.Error() != "el mes usa AAAA-MM" {
		t.Fatalf("esperado error de mes, obtenido: %v", err)
	}

	// 2. ID inválido
	_, err = uc.Crear(ctx, dto.GuardarViveroDTO{
		ID:   "invalido",
		Area: "Fauna",
	})
	if err == nil || err.Error() != "id debe ser un UUID" {
		t.Fatalf("esperado error de UUID, obtenido: %v", err)
	}

	// 3. Área no reconocida
	_, err = uc.Crear(ctx, dto.GuardarViveroDTO{
		ID:   "11111111-1111-1111-1111-111111111111",
		Area: "Invalida",
	})
	if err == nil || err.Error() != "el área debe estar en el catálogo" {
		t.Fatalf("esperado error de área de catálogo, obtenido: %v", err)
	}

	// 4. Archivar con ID inválido
	err = uc.Archivar(ctx, "no-uuid")
	if err == nil || err.Error() != "id debe ser un UUID" {
		t.Fatalf("esperado error de UUID en archivar, obtenido: %v", err)
	}
}

func TestViveroUseCase_CrearYEditar(t *testing.T) {
	ctx := context.Background()
	repo := &mockViveroRepo{}
	uc := usecases.NewViveroUseCase(repo)

	in := dto.GuardarViveroDTO{
		ID:            "11111111-1111-1111-1111-111111111111",
		Fecha:         "2026-05-10",
		Area:          "Flora",
		Subproceso:    "Siembra",
		Etapa:         "Germinación",
		Descripcion:   "Siembra de 20 plantones",
		Observaciones: "En invernadero",
		Responsables:  "Juan Perez",
		LugarID:       "VIV-01",
		LugarLibre:    "Mesa 2",
	}

	res, err := uc.Crear(ctx, in)
	if err != nil {
		t.Fatalf("error inesperado en Crear: %v", err)
	}
	if res.ID != in.ID || res.Area != "Flora" || res.Fecha != "2026-05-10" {
		t.Fatalf("resultado inesperado en Crear: %+v", res)
	}

	// Editar
	in.Etapa = "Crecimiento"
	resEdit, err := uc.Editar(ctx, in)
	if err != nil {
		t.Fatalf("error inesperado en Editar: %v", err)
	}
	if resEdit.Etapa != "Crecimiento" {
		t.Fatalf("resultado inesperado en Editar: %+v", resEdit)
	}

	// Archivar
	if err := uc.Archivar(ctx, "11111111-1111-1111-1111-111111111111"); err != nil {
		t.Fatalf("error inesperado en Archivar: %v", err)
	}
	if repo.archivado != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("id archivado inesperado: %s", repo.archivado)
	}
}

func TestViveroUseCase_Listar(t *testing.T) {
	ctx := context.Background()
	fecha := "2026-05-15"
	lugarID := "LUG-01"
	repo := &mockViveroRepo{
		registros: []*entities.Vivero{
			{
				ID:          "11111111-1111-1111-1111-111111111111",
				Fecha:       &fecha,
				Area:        "Fauna",
				Subproceso:  "Monitoreo",
				Etapa:       "Observación",
				Descripcion: "Registro de aves",
				LugarID:     &lugarID,
			},
		},
	}
	uc := usecases.NewViveroUseCase(repo)

	res, err := uc.Listar(ctx, "2026-05")
	if err != nil {
		t.Fatalf("error inesperado en Listar: %v", err)
	}
	if repo.mesListar != "2026-05" {
		t.Fatalf("mes pasado al repo inesperado: %s", repo.mesListar)
	}
	if len(res.Registros) != 1 {
		t.Fatalf("se esperaba 1 registro, obtenidos %d", len(res.Registros))
	}
	if res.Registros[0].Area != "Fauna" || res.Registros[0].Fecha != "2026-05-15" || res.Registros[0].LugarID != "LUG-01" {
		t.Fatalf("registro inesperado: %+v", res.Registros[0])
	}
}
