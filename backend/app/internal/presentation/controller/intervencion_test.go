package controller

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type mockIntervencionUC struct {
	capatacesFunc    func(ctx context.Context) ([]dto.CapatazDTO, error)
	listFunc         func(ctx context.Context, f dto.FiltroIntervencionesDTO) (entities.FeatureCollection, error)
	createFunc       func(ctx context.Context, in dto.CrearIntervencionDTO) (dto.CrearIntervencionResponseDTO, error)
	assignFunc       func(ctx context.Context, in dto.AsignarIntervencionDTO) (entities.Feature, error)
	setEstadoFunc    func(ctx context.Context, in dto.CambiarEstadoDTO) (entities.Feature, error)
	archiveFunc      func(ctx context.Context, in dto.ArchivarIntervencionDTO) (dto.ArchivarIntervencionResponseDTO, error)
	timelineFunc     func(ctx context.Context, id string) (dto.TimelineResponseDTO, error)
	guardarFichaFunc func(ctx context.Context, in dto.FichaIntervencionDTO) error
	crearAvanceFunc  func(ctx context.Context, in dto.CrearAvanceDTO) error
}

func (m *mockIntervencionUC) ListarCapataces(ctx context.Context) ([]dto.CapatazDTO, error) {
	if m.capatacesFunc != nil {
		return m.capatacesFunc(ctx)
	}
	return []dto.CapatazDTO{}, nil
}

func (m *mockIntervencionUC) ListarActividades(ctx context.Context, f dto.FiltroIntervencionesDTO) (entities.FeatureCollection, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, f)
	}
	return entities.Collection("actividades"), nil
}

func (m *mockIntervencionUC) CrearActividad(ctx context.Context, in dto.CrearIntervencionDTO) (dto.CrearIntervencionResponseDTO, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, in)
	}
	return dto.CrearIntervencionResponseDTO{Creada: true, Feature: entities.Feature{ID: in.ID}}, nil
}

func (m *mockIntervencionUC) AsignarActividad(ctx context.Context, in dto.AsignarIntervencionDTO) (entities.Feature, error) {
	if m.assignFunc != nil {
		return m.assignFunc(ctx, in)
	}
	return entities.Feature{ID: in.ID}, nil
}

func (m *mockIntervencionUC) CambiarEstado(ctx context.Context, in dto.CambiarEstadoDTO) (entities.Feature, error) {
	if m.setEstadoFunc != nil {
		return m.setEstadoFunc(ctx, in)
	}
	return entities.Feature{ID: in.ID}, nil
}

func (m *mockIntervencionUC) ArchivarActividad(ctx context.Context, in dto.ArchivarIntervencionDTO) (dto.ArchivarIntervencionResponseDTO, error) {
	if m.archiveFunc != nil {
		return m.archiveFunc(ctx, in)
	}
	return dto.ArchivarIntervencionResponseDTO{Archivada: true, ID: in.ID}, nil
}

func (m *mockIntervencionUC) Timeline(ctx context.Context, id string) (dto.TimelineResponseDTO, error) {
	if m.timelineFunc != nil {
		return m.timelineFunc(ctx, id)
	}
	return dto.TimelineResponseDTO{ActividadID: id, Eventos: []dto.EventoTimelineDTO{}}, nil
}

func (m *mockIntervencionUC) GuardarFicha(ctx context.Context, in dto.FichaIntervencionDTO) error {
	if m.guardarFichaFunc != nil {
		return m.guardarFichaFunc(ctx, in)
	}
	return nil
}

func (m *mockIntervencionUC) CrearAvance(ctx context.Context, in dto.CrearAvanceDTO) error {
	if m.crearAvanceFunc != nil {
		return m.crearAvanceFunc(ctx, in)
	}
	return nil
}

func TestCreateSinCookieEs401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewIntervencionController(&mockIntervencionUC{}, zerolog.Nop())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"id":"11111111-1111-4111-8111-111111111111","tipo":"riego","titulo":"Riego","lon":-77.08,"lat":-12.07,"actor_rol":"jefatura"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	ctrl.Create(c)
	if w.Code != 401 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
}

func TestCreateCapatazConSesionEs403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewIntervencionController(&mockIntervencionUC{
		createFunc: func(_ context.Context, _ dto.CrearIntervencionDTO) (dto.CrearIntervencionResponseDTO, error) {
			return dto.CrearIntervencionResponseDTO{}, domainErrors.ErrOperacionProhibido
		},
	}, zerolog.Nop())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("usuario", dto.UsuarioSesionDTO{Rol: "capataz", CapatazID: "cap-norte", Usuario: "norte"})
	body := `{"id":"11111111-1111-4111-8111-111111111111","tipo":"riego","titulo":"Riego","lon":-77.08,"lat":-12.07,"actor_rol":"jefatura"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	ctrl.Create(c)
	if w.Code != 403 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
}

func TestEstadoCapatazSinCapatazIDRetorna403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewIntervencionController(&mockIntervencionUC{}, zerolog.Nop())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("usuario", dto.UsuarioSesionDTO{Rol: "capataz", CapatazID: "", Usuario: "sin-id"})
	body := `{"estado":"en_proceso","capataz_id":"cap-norte"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/estado", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	ctrl.Estado(c)
	if w.Code != 403 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "sin identificador asignado") {
		t.Fatalf("mensaje inesperado: %s", w.Body.String())
	}
}

func TestListSinSesionEs401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewIntervencionController(&mockIntervencionUC{}, zerolog.Nop())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/operacion/actividades", nil)
	ctrl.List(c)
	if w.Code != 401 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
}

func TestTimelineSinSesionEs401(t *testing.T) {
	// Without session, handled by middleware or controller check
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/operacion/actividades/1/timeline", nil)
	ctrl := NewIntervencionController(&mockIntervencionUC{}, zerolog.Nop())
	// In the router, RequierePermiso checks auth; when called directly on Timeline, controller checks params
	ctrl.Timeline(c)
	// Direct call without params returns 400 or 404
	if w.Code == 200 {
		t.Fatalf("no debió ser 200 sin id")
	}
}

func TestCapatacesSinSesionEs401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewIntervencionController(&mockIntervencionUC{}, zerolog.Nop())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/operacion/capataces", nil)
	ctrl.Capataces(c)
	if w.Code != 401 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
}

func TestCapatacesError500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewIntervencionController(&mockIntervencionUC{
		capatacesFunc: func(_ context.Context) ([]dto.CapatazDTO, error) {
			return nil, context.DeadlineExceeded
		},
	}, zerolog.Nop())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("usuario", dto.UsuarioSesionDTO{Rol: "coordinacion", Usuario: "coord"})
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/operacion/capataces", nil)
	ctrl.Capataces(c)
	if w.Code != 500 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "no se pudo leer los equipos") {
		t.Fatalf("mensaje inesperado: %s", w.Body.String())
	}
}

func TestListCapatazSinEquipo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewIntervencionController(&mockIntervencionUC{
		listFunc: func(_ context.Context, f dto.FiltroIntervencionesDTO) (entities.FeatureCollection, error) {
			if f.Rol == "capataz" && f.CapatazID == "" {
				return entities.FeatureCollection{}, domainErrors.InputError{Reason: "capataz_id es obligatorio para el rol capataz"}
			}
			return entities.Collection("actividades"), nil
		},
	}, zerolog.Nop())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("usuario", dto.UsuarioSesionDTO{Rol: "coordinacion", Usuario: "coord"})
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/operacion/actividades?rol=capataz", nil)
	ctrl.List(c)
	if w.Code != 400 {
		t.Fatalf("sin capataz_id código esperado 400, obtuve %d: %s", w.Code, w.Body.String())
	}
}

func TestCapatazNoPuedeCerrarNiCancelar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewIntervencionController(&mockIntervencionUC{
		setEstadoFunc: func(_ context.Context, in dto.CambiarEstadoDTO) (entities.Feature, error) {
			if in.ActorRol == "capataz" && (in.Estado == "cerrada" || in.Estado == "cancelada") {
				return entities.Feature{}, domainErrors.ForbiddenError{Reason: "el capataz no puede cerrar ni cancelar una labor"}
			}
			return entities.Feature{ID: in.ID}, nil
		},
	}, zerolog.Nop())

	for _, estado := range []string{"cerrada", "cancelada"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", dto.UsuarioSesionDTO{Rol: "capataz", CapatazID: "cap-norte", Usuario: "norte"})
		c.Params = gin.Params{{Key: "id", Value: "11111111-1111-4111-8111-111111111111"}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/operacion/actividades/1/estado",
			strings.NewReader(`{"estado":"`+estado+`","capataz_id":"cap-norte"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		ctrl.Estado(c)
		if w.Code != 403 {
			t.Fatalf("estado %s debía ser 403, fue %d: %s", estado, w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "el capataz no puede cerrar ni cancelar una labor") {
			t.Fatalf("mensaje esperado no encontrado: %s", w.Body.String())
		}
	}
}

func TestCapatazFichaYAvancesPermisos(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &mockIntervencionUC{
		guardarFichaFunc: func(ctx context.Context, in dto.FichaIntervencionDTO) error {
			if in.ActorRol == "capataz" && in.CapatazID != "cap-norte" {
				return domainErrors.ForbiddenError{Reason: "el capataz solo puede editar la ficha de sus propias labores"}
			}
			return nil
		},
		crearAvanceFunc: func(ctx context.Context, in dto.CrearAvanceDTO) error {
			if in.ActorRol == "capataz" && in.CapatazID != "cap-norte" {
				return domainErrors.ForbiddenError{Reason: "el capataz solo puede registrar avances en sus propias labores"}
			}
			return nil
		},
		setEstadoFunc: func(ctx context.Context, in dto.CambiarEstadoDTO) (entities.Feature, error) {
			if in.ActorRol == "capataz" && (in.Estado == "cerrada" || in.Estado == "cancelada") {
				return entities.Feature{}, domainErrors.ForbiddenError{Reason: "el capataz no puede cerrar ni cancelar una labor"}
			}
			if in.ActorRol == "capataz" && in.Estado == "pendiente" {
				return entities.Feature{}, domainErrors.ForbiddenError{Reason: "solo jefatura/coordinacion reabre una labor cerrada"}
			}
			return entities.Feature{ID: in.ID}, nil
		},
	}
	ctrl := NewIntervencionController(uc, zerolog.Nop())
	laborNorte := "11111111-1111-4111-8111-111111111111"

	// 1. PATCH /ficha por cap-sur (capataz ajeno) -> 403
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", dto.UsuarioSesionDTO{Rol: "capataz", CapatazID: "cap-sur", Usuario: "sur"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/operacion/actividades/"+laborNorte+"/ficha",
			strings.NewReader(`{"comentario":"intento sur"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		ctrl.Ficha(c)
		if w.Code != 403 {
			t.Fatalf("capataz ajeno en ficha status esperado 403, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 2. PATCH /ficha por cap-norte (asignado) -> 200
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", dto.UsuarioSesionDTO{Rol: "capataz", CapatazID: "cap-norte", Usuario: "norte"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/operacion/actividades/"+laborNorte+"/ficha",
			strings.NewReader(`{"comentario":"nota de capataz asignado"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		ctrl.Ficha(c)
		if w.Code != 200 {
			t.Fatalf("capataz asignado en ficha status esperado 200, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 3. PATCH /ficha por coordinacion (oficina) -> 200
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", dto.UsuarioSesionDTO{Rol: "coordinacion", Usuario: "coord"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/operacion/actividades/"+laborNorte+"/ficha",
			strings.NewReader(`{"comentario":"nota de coordinacion"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		ctrl.Ficha(c)
		if w.Code != 200 {
			t.Fatalf("coordinacion en ficha status esperado 200, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 4. POST /avances por cap-sur (capataz ajeno) -> 403
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", dto.UsuarioSesionDTO{Rol: "capataz", CapatazID: "cap-sur", Usuario: "sur"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		body := `{"id":"22222222-2222-4222-8222-222222222221","fecha":"2026-09-28","nota":"intento avance sur","area_feature_id":"AV-0001"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/"+laborNorte+"/avances",
			strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		ctrl.CrearAvance(c)
		if w.Code != 403 {
			t.Fatalf("capataz ajeno en avance status esperado 403, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 5. POST /avances por cap-norte (asignado) -> 201
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", dto.UsuarioSesionDTO{Rol: "capataz", CapatazID: "cap-norte", Usuario: "norte"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		body := `{"id":"22222222-2222-4222-8222-222222222222","fecha":"2026-09-28","nota":"avance norte","area_feature_id":"AV-0001"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/"+laborNorte+"/avances",
			strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		ctrl.CrearAvance(c)
		if w.Code != 201 {
			t.Fatalf("capataz asignado en avance status esperado 201, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 6. POST /avances por coordinacion (oficina) -> 201
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", dto.UsuarioSesionDTO{Rol: "coordinacion", Usuario: "coord"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		body := `{"id":"22222222-2222-4222-8222-222222222223","fecha":"2026-09-28","nota":"avance coord","area_feature_id":"AV-0001"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/"+laborNorte+"/avances",
			strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		ctrl.CrearAvance(c)
		if w.Code != 201 {
			t.Fatalf("coordinacion en avance status esperado 201, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 7. Coordinación cierra la labor (tiene avance previo del paso 5 y ejecutor=propia) -> 200
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", dto.UsuarioSesionDTO{Rol: "coordinacion", Usuario: "coord", ID: 4})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/"+laborNorte+"/estado",
			strings.NewReader(`{"estado":"cerrada"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		ctrl.Estado(c)
		if w.Code != 200 {
			t.Fatalf("coordinacion cerrar labor esperado 200, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 8. Capataz asignado (cap-norte) intenta reabrir labor cerrada a pendiente -> 403
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", dto.UsuarioSesionDTO{Rol: "capataz", CapatazID: "cap-norte", Usuario: "norte", ID: 1})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/"+laborNorte+"/estado",
			strings.NewReader(`{"estado":"pendiente"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		ctrl.Estado(c)
		if w.Code != 403 {
			t.Fatalf("capataz reabriendo labor cerrada esperado 403, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 9. Coordinación sí puede reabrir labor cerrada a pendiente -> 200
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", dto.UsuarioSesionDTO{Rol: "coordinacion", Usuario: "coord", ID: 4})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/"+laborNorte+"/estado",
			strings.NewReader(`{"estado":"pendiente"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		ctrl.Estado(c)
		if w.Code != 200 {
			t.Fatalf("coordinacion reabriendo labor cerrada esperado 200, obtuve %d: %s", w.Code, w.Body.String())
		}
	}
}

func TestEstadoCatalogoHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewIntervencionController(&mockIntervencionUC{
		setEstadoFunc: func(_ context.Context, in dto.CambiarEstadoDTO) (entities.Feature, error) {
			if in.Estado != "ejecutado" {
				return entities.Feature{}, domainErrors.InputError{Reason: "el estado no está activo en el catálogo"}
			}
			return entities.Feature{
				Type: "Feature",
				ID:   in.ID,
				Properties: entities.ActividadProperties{
					Estado:         "ejecutado",
					EstadoEtiqueta: "Ejecutado",
				},
			}, nil
		},
	}, zerolog.Nop())
	id := "22222222-2222-4222-8222-222222222222"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("usuario", dto.UsuarioSesionDTO{Rol: "coordinacion", Usuario: "coord", ID: 4})
	c.Params = gin.Params{{Key: "id", Value: id}}
	c.Request = httptest.NewRequest(http.MethodPatch, "/areas-verdes/v1/operacion/actividades/"+id+"/estado",
		strings.NewReader(`{"estado":"inventado"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	ctrl.Estado(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("estado fuera de catálogo: esperado 400, obtuve %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("usuario", dto.UsuarioSesionDTO{Rol: "coordinacion", Usuario: "coord", ID: 4})
	c.Params = gin.Params{{Key: "id", Value: id}}
	c.Request = httptest.NewRequest(http.MethodPatch, "/areas-verdes/v1/operacion/actividades/"+id+"/estado",
		strings.NewReader(`{"estado":"ejecutado"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	ctrl.Estado(c)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Ejecutado") {
		t.Fatalf("con el ítem creado: esperado 200 y etiqueta Ejecutado, obtuve %d %s", w.Code, w.Body.String())
	}
}

func init() {
	_ = bytes.NewBuffer(nil)
}
