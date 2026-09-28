package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type mockIntervencionRepo struct {
	capatacesFunc    func(ctx context.Context) ([]entities.Capataz, error)
	listFunc         func(ctx context.Context, f dto.FiltroIntervencionesDTO) (entities.FeatureCollection, error)
	createFunc       func(ctx context.Context, in dto.CrearIntervencionDTO) (entities.Feature, bool, error)
	assignFunc       func(ctx context.Context, in dto.AsignarIntervencionDTO) (entities.Feature, error)
	setEstadoFunc    func(ctx context.Context, in dto.CambiarEstadoDTO) (entities.Feature, error)
	archiveFunc      func(ctx context.Context, in dto.ArchivarIntervencionDTO) error
	timelineFunc     func(ctx context.Context, id string) (dto.TimelineResponseDTO, error)
	guardarFichaFunc func(ctx context.Context, in dto.FichaIntervencionDTO) error
	crearAvanceFunc  func(ctx context.Context, in dto.CrearAvanceDTO) error
	oneFunc          func(ctx context.Context, id string) (entities.Feature, error)
}

func (m *mockIntervencionRepo) Capataces(ctx context.Context) ([]entities.Capataz, error) {
	if m.capatacesFunc != nil {
		return m.capatacesFunc(ctx)
	}
	return nil, nil
}

func (m *mockIntervencionRepo) List(ctx context.Context, f dto.FiltroIntervencionesDTO) (entities.FeatureCollection, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return entities.Collection("actividades"), nil
}

func (m *mockIntervencionRepo) Create(ctx context.Context, in dto.CrearIntervencionDTO) (entities.Feature, bool, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, in)
	}
	return entities.Feature{ID: in.ID}, true, nil
}

func (m *mockIntervencionRepo) Assign(ctx context.Context, in dto.AsignarIntervencionDTO) (entities.Feature, error) {
	if m.assignFunc != nil {
		return m.assignFunc(ctx, in)
	}
	return entities.Feature{ID: in.ID}, nil
}

func (m *mockIntervencionRepo) SetEstado(ctx context.Context, in dto.CambiarEstadoDTO) (entities.Feature, error) {
	if m.setEstadoFunc != nil {
		return m.setEstadoFunc(ctx, in)
	}
	return entities.Feature{ID: in.ID}, nil
}

func (m *mockIntervencionRepo) Archive(ctx context.Context, in dto.ArchivarIntervencionDTO) error {
	if m.archiveFunc != nil {
		return m.archiveFunc(ctx, in)
	}
	return nil
}

func (m *mockIntervencionRepo) Timeline(ctx context.Context, id string) (dto.TimelineResponseDTO, error) {
	if m.timelineFunc != nil {
		return m.timelineFunc(ctx, id)
	}
	return dto.TimelineResponseDTO{ActividadID: id}, nil
}

func (m *mockIntervencionRepo) GuardarFicha(ctx context.Context, in dto.FichaIntervencionDTO) error {
	if m.guardarFichaFunc != nil {
		return m.guardarFichaFunc(ctx, in)
	}
	return nil
}

func (m *mockIntervencionRepo) CrearAvance(ctx context.Context, in dto.CrearAvanceDTO) error {
	if m.crearAvanceFunc != nil {
		return m.crearAvanceFunc(ctx, in)
	}
	return nil
}

func (m *mockIntervencionRepo) One(ctx context.Context, id string) (entities.Feature, error) {
	if m.oneFunc != nil {
		return m.oneFunc(ctx, id)
	}
	return entities.Feature{ID: id}, nil
}

func TestValidateCreateCapatazProhibido(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{})
	_, err := uc.CrearActividad(context.Background(), dto.CrearIntervencionDTO{
		ActorRol: "capataz",
		ID:       "11111111-1111-4111-8111-111111111111",
		Tipo:     "riego",
		Titulo:   "Riego",
		Lon:      -77.08,
		Lat:      -12.07,
	})
	if !errors.Is(err, domainErrors.ErrOperacionProhibido) {
		t.Fatalf("esperado prohibido, obtuve %v", err)
	}
}

func TestValidateQueryCapatazSinEquipo(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{})
	_, err := uc.ListarActividades(context.Background(), dto.FiltroIntervencionesDTO{
		Rol:          "capataz",
		SoloAbiertas: true,
	})
	if !errors.Is(err, domainErrors.ErrValidacionOperacion) {
		t.Fatalf("esperado validacion, obtuve %v", err)
	}
}

func TestSamePayload(t *testing.T) {
	in := dto.CrearIntervencionDTO{
		Tipo:              "poda",
		Titulo:            " Poda ",
		Detalle:           "borde",
		Lon:               -77.0808,
		Lat:               -12.0704,
		AssignedCapatazID: "cap-sur",
	}
	saved := SavedPayload{
		Tipo:              "poda",
		Titulo:            "Poda",
		Detalle:           "borde",
		Lon:               -77.0808,
		Lat:               -12.0704,
		AssignedCapatazID: "cap-sur",
	}
	if !SamePayload(saved, in) {
		t.Fatal("debía coincidir")
	}
	saved.Titulo = "Otra"
	if SamePayload(saved, in) {
		t.Fatal("no debía coincidir")
	}
}

func TestAltaPorLugarSinPin(t *testing.T) {
	repo := &mockIntervencionRepo{}
	uc := NewIntervencionUseCase(repo)
	_, err := uc.CrearActividad(context.Background(), dto.CrearIntervencionDTO{
		ActorRol: "coordinacion",
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
	if PuedeCerrar(string(enums.EjecutorTercerizada), false, true) == nil {
		t.Fatal("sin orden no cierra")
	}
	if PuedeCerrar(string(enums.EjecutorPropia), false, true) != nil {
		t.Fatal("la propia con ejecución sí cierra")
	}
}

func TestPuntoFuera(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{})
	_, err := uc.CrearActividad(context.Background(), dto.CrearIntervencionDTO{
		ActorRol: "coordinacion",
		ID:       "11111111-1111-4111-8111-111111111111",
		Tipo:     "riego",
		Titulo:   "Lejos",
		Lon:      -70,
		Lat:      -12.07,
	})
	var input domainErrors.InputError
	if !errors.As(err, &input) {
		t.Fatalf("esperado input error, obtuve %v", err)
	}
}

func TestListarCapataces(t *testing.T) {
	repo := &mockIntervencionRepo{
		capatacesFunc: func(_ context.Context) ([]entities.Capataz, error) {
			return []entities.Capataz{
				{ID: "cap-norte", Equipo: "Equipo Norte", Turno: "mañana", Activo: true},
			}, nil
		},
	}
	uc := NewIntervencionUseCase(repo)
	res, err := uc.ListarCapataces(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 || res[0].ID != "cap-norte" {
		t.Fatalf("resultado inesperado: %+v", res)
	}
}

func TestAsignarActividadValidacion(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{})
	// Capataz actor cannot assign
	_, err := uc.AsignarActividad(context.Background(), dto.AsignarIntervencionDTO{
		ID:        "11111111-1111-4111-8111-111111111111",
		CapatazID: "cap-sur",
		ActorRol:  "capataz",
	})
	if !errors.Is(err, domainErrors.ErrOperacionProhibido) {
		t.Fatalf("esperado prohibido, obtuve %v", err)
	}

	// Empty capataz_id fails
	_, err = uc.AsignarActividad(context.Background(), dto.AsignarIntervencionDTO{
		ID:        "11111111-1111-4111-8111-111111111111",
		CapatazID: "",
		ActorRol:  "coordinacion",
	})
	var input domainErrors.InputError
	if !errors.As(err, &input) {
		t.Fatalf("esperado input error, obtuve %v", err)
	}
}

func TestCambiarEstadoValidacion(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{})
	// Capataz cannot close or cancel
	for _, st := range []string{"cerrada", "cancelada"} {
		_, err := uc.CambiarEstado(context.Background(), dto.CambiarEstadoDTO{
			ID:        "11111111-1111-4111-8111-111111111111",
			Estado:    st,
			ActorRol:  "capataz",
			CapatazID: "cap-norte",
		})
		var forb domainErrors.ForbiddenError
		if !errors.As(err, &forb) {
			t.Fatalf("esperado ForbiddenError para estado %s, obtuve %v", st, err)
		}
	}

	// Capataz without capataz_id fails
	_, err := uc.CambiarEstado(context.Background(), dto.CambiarEstadoDTO{
		ID:        "11111111-1111-4111-8111-111111111111",
		Estado:    "en_proceso",
		ActorRol:  "capataz",
		CapatazID: "",
	})
	var input domainErrors.InputError
	if !errors.As(err, &input) {
		t.Fatalf("esperado input error, obtuve %v", err)
	}
}

func TestArchivarValidacion(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{})
	// Capataz cannot archive
	_, err := uc.ArchivarActividad(context.Background(), dto.ArchivarIntervencionDTO{
		ID:       "11111111-1111-4111-8111-111111111111",
		ActorRol: "capataz",
	})
	if !errors.Is(err, domainErrors.ErrOperacionProhibido) {
		t.Fatalf("esperado prohibido, obtuve %v", err)
	}

	// Coordinación can archive
	res, err := uc.ArchivarActividad(context.Background(), dto.ArchivarIntervencionDTO{
		ID:       "11111111-1111-4111-8111-111111111111",
		ActorRol: "coordinacion",
		Motivo:   "duplicada",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Archivada || res.ID != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("respuesta inesperada: %+v", res)
	}
}

func TestTimelineValidacion(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{})
	// Invalid UUID fails
	_, err := uc.Timeline(context.Background(), "no-uuid")
	var input domainErrors.InputError
	if !errors.As(err, &input) {
		t.Fatalf("esperado input error, obtuve %v", err)
	}
}

func TestGuardarFichaValidacion(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{})
	// Invalid UUID fails
	err := uc.GuardarFicha(context.Background(), dto.FichaIntervencionDTO{
		ID: "no-uuid",
	})
	var input domainErrors.InputError
	if !errors.As(err, &input) {
		t.Fatalf("esperado input error, obtuve %v", err)
	}

	// Fecha atención anterior a solicitud fails
	err = uc.GuardarFicha(context.Background(), dto.FichaIntervencionDTO{
		ID:             "11111111-1111-4111-8111-111111111111",
		FechaSolicitud: "2026-09-20",
		FechaAtencion:  "2026-09-10",
	})
	if !errors.As(err, &input) {
		t.Fatalf("esperado input error, obtuve %v", err)
	}
}

func TestCrearAvanceValidacion(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{})
	// Invalid UUID fails
	err := uc.CrearAvance(context.Background(), dto.CrearAvanceDTO{
		ActividadID:   "no-uuid",
		ID:            "22222222-2222-4222-8222-222222222222",
		Fecha:         "2026-09-28",
		AreaFeatureID: "AV-0001",
	})
	var input domainErrors.InputError
	if !errors.As(err, &input) {
		t.Fatalf("esperado input error, obtuve %v", err)
	}

	// Invalid date format fails
	err = uc.CrearAvance(context.Background(), dto.CrearAvanceDTO{
		ActividadID:   "11111111-1111-4111-8111-111111111111",
		ID:            "22222222-2222-4222-8222-222222222222",
		Fecha:         "28/09/2026",
		AreaFeatureID: "AV-0001",
	})
	if !errors.As(err, &input) {
		t.Fatalf("esperado input error, obtuve %v", err)
	}

	// Sin area ni ejemplar fails
	err = uc.CrearAvance(context.Background(), dto.CrearAvanceDTO{
		ActividadID: "11111111-1111-4111-8111-111111111111",
		ID:          "22222222-2222-4222-8222-222222222222",
		Fecha:       "2026-09-28",
	})
	if !errors.As(err, &input) {
		t.Fatalf("esperado input error, obtuve %v", err)
	}
}
