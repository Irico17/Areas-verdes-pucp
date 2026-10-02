package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type mockIntervencionRepo struct {
	capatacesFunc    func(ctx context.Context) ([]entities.Capataz, error)
	listFunc         func(ctx context.Context, f entities.FiltroIntervenciones) (entities.FeatureCollection, error)
	createFunc       func(ctx context.Context, in entities.NuevaIntervencion) (entities.Feature, bool, error)
	assignFunc       func(ctx context.Context, in entities.AsignarIntervencion) (entities.Feature, error)
	setEstadoFunc    func(ctx context.Context, in entities.CambiarEstadoIntervencion) (entities.Feature, error)
	estadoActualFunc func(ctx context.Context, id string) (string, error)
	archiveFunc      func(ctx context.Context, in entities.ArchivarIntervencion) error
	timelineFunc     func(ctx context.Context, id string) ([]entities.ActividadEvento, error)
	guardarFichaFunc func(ctx context.Context, in entities.FichaIntervencion) error
	crearAvanceFunc  func(ctx context.Context, in entities.NuevoAvance) error
	oneFunc          func(ctx context.Context, id string) (entities.Feature, error)
}

func (m *mockIntervencionRepo) Capataces(ctx context.Context) ([]entities.Capataz, error) {
	if m.capatacesFunc != nil {
		return m.capatacesFunc(ctx)
	}
	return nil, nil
}

func (m *mockIntervencionRepo) List(ctx context.Context, f entities.FiltroIntervenciones) (entities.FeatureCollection, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return entities.Collection("actividades"), nil
}

func (m *mockIntervencionRepo) Create(ctx context.Context, in entities.NuevaIntervencion) (entities.Feature, bool, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, in)
	}
	return entities.Feature{ID: in.ID}, true, nil
}

func (m *mockIntervencionRepo) Assign(ctx context.Context, in entities.AsignarIntervencion) (entities.Feature, error) {
	if m.assignFunc != nil {
		return m.assignFunc(ctx, in)
	}
	return entities.Feature{ID: in.ID}, nil
}

func (m *mockIntervencionRepo) SetEstado(ctx context.Context, in entities.CambiarEstadoIntervencion) (entities.Feature, error) {
	if m.setEstadoFunc != nil {
		return m.setEstadoFunc(ctx, in)
	}
	return entities.Feature{ID: in.ID}, nil
}

func (m *mockIntervencionRepo) EstadoActual(ctx context.Context, id string) (string, error) {
	if m.estadoActualFunc != nil {
		return m.estadoActualFunc(ctx, id)
	}
	return "pendiente", nil
}

func (m *mockIntervencionRepo) Archive(ctx context.Context, in entities.ArchivarIntervencion) error {
	if m.archiveFunc != nil {
		return m.archiveFunc(ctx, in)
	}
	return nil
}

func (m *mockIntervencionRepo) Timeline(ctx context.Context, id string) ([]entities.ActividadEvento, error) {
	if m.timelineFunc != nil {
		return m.timelineFunc(ctx, id)
	}
	return []entities.ActividadEvento{}, nil
}

func (m *mockIntervencionRepo) GuardarFicha(ctx context.Context, in entities.FichaIntervencion) error {
	if m.guardarFichaFunc != nil {
		return m.guardarFichaFunc(ctx, in)
	}
	return nil
}

func (m *mockIntervencionRepo) CrearAvance(ctx context.Context, in entities.NuevoAvance) error {
	if m.crearAvanceFunc != nil {
		return m.crearAvanceFunc(ctx, in)
	}
	return nil
}

type catalogoSinItems struct{}

func (catalogoSinItems) List(context.Context, string, bool) ([]entities.CatalogoItem, error) {
	return nil, nil
}
func (catalogoSinItems) Activo(context.Context, string, string) (bool, error) { return false, nil }
func (catalogoSinItems) Create(context.Context, string, string, string) (*entities.CatalogoItem, error) {
	return nil, nil
}
func (catalogoSinItems) Deactivate(context.Context, int64, int64) error { return nil }
func (catalogoSinItems) Renombrar(context.Context, int64, string, int64) (*entities.CatalogoItem, error) {
	return nil, nil
}

func (m *mockIntervencionRepo) Taxonomia(context.Context) (entities.TaxonomiaActividad, error) {
	return entities.TaxonomiaActividad{}, nil
}

func (m *mockIntervencionRepo) ListarPersonal(context.Context, string) ([]entities.PersonalLabor, error) {
	return nil, nil
}

func (m *mockIntervencionRepo) RegistrarPersonal(context.Context, string, []string) ([]entities.PersonalLabor, error) {
	return nil, nil
}

func (m *mockIntervencionRepo) One(ctx context.Context, id string) (entities.Feature, error) {
	if m.oneFunc != nil {
		return m.oneFunc(ctx, id)
	}
	return entities.Feature{ID: id}, nil
}

func TestValidateCreateCapatazProhibido(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{}, catalogoSinItems{})
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
	uc := NewIntervencionUseCase(&mockIntervencionRepo{}, catalogoSinItems{})
	_, err := uc.ListarActividades(context.Background(), dto.FiltroIntervencionesDTO{
		Rol:          "capataz",
		SoloAbiertas: true,
	})
	if !errors.Is(err, domainErrors.ErrValidacionOperacion) {
		t.Fatalf("esperado validacion, obtuve %v", err)
	}
}

func TestSamePayload(t *testing.T) {
	in := entities.NuevaIntervencion{
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
	uc := NewIntervencionUseCase(repo, catalogoSinItems{})
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

type catalogoMapa struct {
	ok map[string]bool
}

func (c catalogoMapa) List(context.Context, string, bool) ([]entities.CatalogoItem, error) {
	return nil, nil
}
func (c catalogoMapa) Activo(_ context.Context, clase, codigo string) (bool, error) {
	return c.ok[clase+"/"+codigo], nil
}
func (c catalogoMapa) Create(context.Context, string, string, string) (*entities.CatalogoItem, error) {
	return nil, nil
}
func (c catalogoMapa) Deactivate(context.Context, int64, int64) error { return nil }
func (c catalogoMapa) Renombrar(context.Context, int64, string, int64) (*entities.CatalogoItem, error) {
	return nil, nil
}

func TestAltaRechazaRiesgoFueraDeCatalogoYLugarLibre(t *testing.T) {
	catalogo := catalogoMapa{ok: map[string]bool{
		"nivel_riesgo/bajo": true,
		"nivel_riesgo/alto": true,
	}}
	uc := NewIntervencionUseCase(&mockIntervencionRepo{}, catalogo)
	base := dto.CrearIntervencionDTO{
		ActorRol: "coordinacion",
		ID:       "11111111-1111-4111-8111-111111111111",
		Tipo:     "riego",
		Titulo:   "Riego",
		Lon:      -77.08,
		Lat:      -12.07,
	}
	medio := base
	medio.NivelRiesgo = "medio"
	if _, err := uc.CrearActividad(context.Background(), medio); err == nil {
		t.Fatal("medio no está en el catálogo")
	}
	libre := base
	libre.LugarTexto = "jardín inventado"
	if _, err := uc.CrearActividad(context.Background(), libre); err == nil {
		t.Fatal("el lugar libre no se acepta")
	}
}

func TestAltaEntregaLosCamposAlRepositorio(t *testing.T) {
	var got entities.NuevaIntervencion
	repo := &mockIntervencionRepo{
		createFunc: func(_ context.Context, in entities.NuevaIntervencion) (entities.Feature, bool, error) {
			got = in
			return entities.Feature{ID: in.ID}, true, nil
		},
	}
	catalogo := catalogoMapa{ok: map[string]bool{
		"fuente/interna":                 true,
		"nivel_riesgo/alto":              true,
		"clase_actividad/riego":          true,
		"subtipo_actividad/riego_manual": true,
	}}
	cantidad := 4.0
	uc := NewIntervencionUseCase(repo, catalogo)
	_, err := uc.CrearActividad(context.Background(), dto.CrearIntervencionDTO{
		ActorRol:          "coordinacion",
		ID:                "11111111-1111-4111-8111-111111111111",
		Tipo:              "riego",
		Titulo:            "Riego del eje",
		Lon:               -77.08,
		Lat:               -12.07,
		Origen:            "interna",
		CodigoExterno:     "CENT-2026-0001",
		UnidadSolicitante: "Oficina de campus",
		NivelRiesgo:       "alto",
		FechaProgramada:   "2026-10-15",
		Cantidad:          &cantidad,
		Clase:             "riego",
		Subtipo:           "riego_manual",
		Personal:          []string{"Elsa Mamani"},
		LugarID:           "12",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Origen != "interna" || got.CodigoExterno != "CENT-2026-0001" || got.UnidadSolicitante != "Oficina de campus" {
		t.Fatalf("texto %+v", got)
	}
	if got.NivelRiesgo != "alto" || got.FechaProgramada != "2026-10-15" || got.Cantidad == nil || *got.Cantidad != 4 {
		t.Fatalf("pedido %+v", got)
	}
	if got.Clase != "riego" || got.Subtipo != "riego_manual" || len(got.Personal) != 1 {
		t.Fatalf("taxonomía %+v", got)
	}
}

func TestPuntoFuera(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{}, catalogoSinItems{})
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
	uc := NewIntervencionUseCase(repo, catalogoSinItems{})
	res, err := uc.ListarCapataces(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 || res[0].ID != "cap-norte" {
		t.Fatalf("resultado inesperado: %+v", res)
	}
}

func TestAsignarActividadValidacion(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{}, catalogoSinItems{})
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
	uc := NewIntervencionUseCase(&mockIntervencionRepo{}, catalogoSinItems{})
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

type catalogoEstadosFijos struct{}

func (catalogoEstadosFijos) List(context.Context, string, bool) ([]entities.CatalogoItem, error) {
	return []entities.CatalogoItem{
		{Codigo: "pendiente", Nombre: "Por iniciar", Activo: true},
		{Codigo: "en_proceso", Nombre: "En proceso", Activo: true},
		{Codigo: "ejecutado", Nombre: "Ejecutado", Activo: true},
		{Codigo: "cerrada", Nombre: "Cerrado", Activo: true},
		{Codigo: "bloqueada", Nombre: "Bloqueada", Activo: false},
	}, nil
}
func (catalogoEstadosFijos) Activo(context.Context, string, string) (bool, error) {
	return false, nil
}
func (catalogoEstadosFijos) Create(context.Context, string, string, string) (*entities.CatalogoItem, error) {
	return nil, nil
}
func (catalogoEstadosFijos) Deactivate(context.Context, int64, int64) error { return nil }
func (catalogoEstadosFijos) Renombrar(context.Context, int64, string, int64) (*entities.CatalogoItem, error) {
	return nil, nil
}

func TestCambiarEstadoContraCatalogo(t *testing.T) {
	id := "11111111-1111-4111-8111-111111111111"
	catalogo := catalogoEstadosFijos{}

	t.Run("desconocido", func(t *testing.T) {
		uc := NewIntervencionUseCase(&mockIntervencionRepo{}, catalogo)
		_, err := uc.CambiarEstado(context.Background(), dto.CambiarEstadoDTO{
			ID: id, Estado: "inventado", ActorRol: "coordinacion",
		})
		var input domainErrors.InputError
		if !errors.As(err, &input) || input.Reason != "el estado no está activo en el catálogo" {
			t.Fatalf("se esperaba 400 de catálogo, obtuve %v", err)
		}
	})

	t.Run("no salta el cierre desde por iniciar", func(t *testing.T) {
		uc := NewIntervencionUseCase(&mockIntervencionRepo{
			estadoActualFunc: func(context.Context, string) (string, error) { return "pendiente", nil },
		}, catalogo)
		_, err := uc.CambiarEstado(context.Background(), dto.CambiarEstadoDTO{
			ID: id, Estado: "cerrada", ActorRol: "coordinacion",
		})
		var input domainErrors.InputError
		if !errors.As(err, &input) || input.Reason != "esa transición de estado no está permitida" {
			t.Fatalf("se esperaba transición rechazada, obtuve %v", err)
		}
	})

	t.Run("catalogo y transicion", func(t *testing.T) {
		uc := NewIntervencionUseCase(&mockIntervencionRepo{
			estadoActualFunc: func(context.Context, string) (string, error) { return "en_proceso", nil },
			setEstadoFunc: func(_ context.Context, in entities.CambiarEstadoIntervencion) (entities.Feature, error) {
				return entities.Feature{
					Type: "Feature",
					ID:   in.ID,
					Properties: entities.ActividadProperties{
						ID:     in.ID,
						Estado: in.Estado,
					},
				}, nil
			},
		}, catalogo)
		feat, err := uc.CambiarEstado(context.Background(), dto.CambiarEstadoDTO{
			ID: id, Estado: "ejecutado", ActorRol: "coordinacion",
		})
		if err != nil {
			t.Fatal(err)
		}
		props, ok := feat.Properties.(entities.ActividadProperties)
		if !ok || props.EstadoEtiqueta != "Ejecutado" || props.Estado != "ejecutado" {
			t.Fatalf("se esperaba etiqueta Ejecutado, obtuve %+v", feat.Properties)
		}
	})
}

func TestArchivarValidacion(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{}, catalogoSinItems{})
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
	uc := NewIntervencionUseCase(&mockIntervencionRepo{}, catalogoSinItems{})
	// Invalid UUID fails
	_, err := uc.Timeline(context.Background(), "no-uuid")
	var input domainErrors.InputError
	if !errors.As(err, &input) {
		t.Fatalf("esperado input error, obtuve %v", err)
	}
}

func TestGuardarFichaValidacion(t *testing.T) {
	uc := NewIntervencionUseCase(&mockIntervencionRepo{}, catalogoSinItems{})
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
	uc := NewIntervencionUseCase(&mockIntervencionRepo{}, catalogoSinItems{})
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

func TestTimelineMapeo(t *testing.T) {
	tm := time.Date(2026, 9, 30, 15, 30, 0, 0, time.UTC)
	st := "en_proceso"
	cap := "cap-01"
	eq := "Equipo 1"
	uID := int64(42)
	repo := &mockIntervencionRepo{
		timelineFunc: func(_ context.Context, id string) ([]entities.ActividadEvento, error) {
			return []entities.ActividadEvento{
				{
					ID:          1,
					ActividadID: id,
					Tipo:        "asignacion",
					Estado:      &st,
					CapatazID:   &cap,
					Equipo:      &eq,
					ActorRol:    "coordinacion",
					UsuarioID:   &uID,
					Usuario:     "coord1",
					Nombre:      "Coord Uno",
					Nota:        "Asignado a capataz",
					CreatedAt:   tm,
				},
			}, nil
		},
	}
	uc := NewIntervencionUseCase(repo, catalogoSinItems{})
	res, err := uc.Timeline(context.Background(), "11111111-1111-4111-8111-111111111111")
	if err != nil {
		t.Fatalf("error inesperado en Timeline: %v", err)
	}
	if res.ActividadID != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("actividad_id inesperado: %s", res.ActividadID)
	}
	if len(res.Eventos) != 1 {
		t.Fatalf("esperado 1 evento, obtuve %d", len(res.Eventos))
	}
	ev := res.Eventos[0]
	if ev.CreatedAt != "2026-09-30T15:30:00Z" {
		t.Fatalf("formato de fecha inesperado: %s", ev.CreatedAt)
	}
	if ev.Nombre != "Coord Uno" || ev.Usuario != "coord1" {
		t.Fatalf("datos de usuario inesperados: %+v", ev)
	}
}
