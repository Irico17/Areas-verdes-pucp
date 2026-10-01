package usecases_test

import (
	"context"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type mockPodaRepo struct {
	podas     []*entities.Poda
	guardada  *entities.Poda
	archivada string
	err       error
}

func (m *mockPodaRepo) Listar(_ context.Context) ([]*entities.Poda, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.podas, nil
}

func (m *mockPodaRepo) Guardar(_ context.Context, in entities.GuardarPoda) (*entities.Poda, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.guardada = &entities.Poda{
		ID:                in.ID,
		Codigo:            in.Codigo,
		CodigoExterno:     &in.CodigoExterno,
		Tipo:              in.Tipo,
		TipoActividad:     in.TipoActividad,
		Personal:          in.Personal,
		Ubicacion:         in.Ubicacion,
		Unidad:            in.Unidad,
		CantidadPedida:    in.CantidadPedida,
		CantidadEjecutada: in.CantidadEjecutada,
		Prioridad:         in.Prioridad,
		Comentario:        in.Comentario,
	}
	return m.guardada, nil
}

func (m *mockPodaRepo) Archivar(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	m.archivada = id
	return nil
}

func TestPodaUseCase_Validaciones(t *testing.T) {
	ctx := context.Background()
	repo := &mockPodaRepo{}
	uc := usecases.NewPodaUseCase(repo)

	// 1. Código inválido
	_, err := uc.Crear(ctx, dto.GuardarPodaDTO{
		ID:        "11111111-1111-1111-1111-111111111111",
		Codigo:    "ABC",
		Prioridad: "media",
	})
	if err == nil || err.Error() != "el código de poda es PO-n" {
		t.Fatalf("esperado error de código, obtenido: %v", err)
	}

	// 2. ID inválido
	_, err = uc.Crear(ctx, dto.GuardarPodaDTO{
		ID:        "invalido",
		Codigo:    "PO-001",
		Prioridad: "media",
	})
	if err == nil || err.Error() != "id debe ser un UUID" {
		t.Fatalf("esperado error de UUID, obtenido: %v", err)
	}

	// 3. Prioridad no reconocida
	_, err = uc.Crear(ctx, dto.GuardarPodaDTO{
		ID:        "11111111-1111-1111-1111-111111111111",
		Codigo:    "PO-001",
		Prioridad: "urgente",
	})
	if err == nil || err.Error() != "prioridad no reconocida" {
		t.Fatalf("esperado error de prioridad, obtenido: %v", err)
	}

	// 4. Cantidad pedida negativa
	_, err = uc.Crear(ctx, dto.GuardarPodaDTO{
		ID:             "11111111-1111-1111-1111-111111111111",
		Codigo:         "PO-001",
		Prioridad:      "media",
		CantidadPedida: -1,
	})
	if err == nil || err.Error() != "las cantidades son cero o más" {
		t.Fatalf("esperado error de cantidades, obtenido: %v", err)
	}

	// 5. Cantidad ejecutada negativa
	_, err = uc.Crear(ctx, dto.GuardarPodaDTO{
		ID:                "11111111-1111-1111-1111-111111111111",
		Codigo:            "PO-001",
		Prioridad:         "alta",
		CantidadEjecutada: -0.5,
	})
	if err == nil || err.Error() != "las cantidades son cero o más" {
		t.Fatalf("esperado error de cantidades, obtenido: %v", err)
	}

	// 6. Archivar con ID inválido
	err = uc.Archivar(ctx, "no-uuid")
	if err == nil || err.Error() != "id debe ser un UUID" {
		t.Fatalf("esperado error de UUID en archivar, obtenido: %v", err)
	}
}

func TestPodaUseCase_CrearYEditar(t *testing.T) {
	ctx := context.Background()
	repo := &mockPodaRepo{}
	uc := usecases.NewPodaUseCase(repo)

	in := dto.GuardarPodaDTO{
		ID:                "11111111-1111-1111-1111-111111111111",
		Codigo:            "PO-001",
		CodigoExterno:     "osg-999",
		Tipo:              "Mantenimiento",
		TipoActividad:     "Descope",
		Personal:          "Cuadrilla 1",
		Ubicacion:         "Pabellón A",
		Unidad:            "árboles",
		CantidadPedida:    5,
		CantidadEjecutada: 5,
		Prioridad:         "alta",
		Comentario:        "Finalizado",
	}

	res, err := uc.Crear(ctx, in)
	if err != nil {
		t.Fatalf("error inesperado en Crear: %v", err)
	}
	if res.Codigo != "PO-001" || res.CodigoExterno != "OSG-999" || res.Prioridad != "alta" {
		t.Fatalf("resultado inesperado en Crear: %+v", res)
	}

	// Editar
	in.Prioridad = "media"
	in.Comentario = "Editado"
	resEdit, err := uc.Editar(ctx, in)
	if err != nil {
		t.Fatalf("error inesperado en Editar: %v", err)
	}
	if resEdit.Prioridad != "media" || resEdit.Comentario != "Editado" {
		t.Fatalf("resultado inesperado en Editar: %+v", resEdit)
	}

	// Archivar
	if err := uc.Archivar(ctx, "11111111-1111-1111-1111-111111111111"); err != nil {
		t.Fatalf("error inesperado en Archivar: %v", err)
	}
	if repo.archivada != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("id archivado inesperado: %s", repo.archivada)
	}
}

func TestPodaUseCase_Listar(t *testing.T) {
	ctx := context.Background()
	codExt := "OSG-100"
	fRep := "2026-05-01"
	repo := &mockPodaRepo{
		podas: []*entities.Poda{
			{
				ID:            "11111111-1111-1111-1111-111111111111",
				Codigo:        "PO-001",
				CodigoExterno: &codExt,
				FechaReporte:  &fRep,
				Prioridad:     "media",
			},
		},
	}
	uc := usecases.NewPodaUseCase(repo)

	res, err := uc.Listar(ctx)
	if err != nil {
		t.Fatalf("error inesperado en Listar: %v", err)
	}
	if len(res.Podas) != 1 {
		t.Fatalf("se esperaba 1 poda, obtenidas %d", len(res.Podas))
	}
	if res.Podas[0].Codigo != "PO-001" || res.Podas[0].CodigoExterno != "OSG-100" || res.Podas[0].FechaReporte != "2026-05-01" {
		t.Fatalf("poda inesperada: %+v", res.Podas[0])
	}
}
