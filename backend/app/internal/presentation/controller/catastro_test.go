package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

type mockZonaUC struct {
	items []dto.ZonaSupervisionDTO
	err   error
}

func (m *mockZonaUC) Listar(_ context.Context) ([]dto.ZonaSupervisionDTO, error) {
	return m.items, m.err
}

func (m *mockZonaUC) Crear(_ context.Context, req dto.CrearZonaSupervisionDTO) (dto.ZonaSupervisionDTO, error) {
	if m.err != nil {
		return dto.ZonaSupervisionDTO{}, m.err
	}
	return dto.ZonaSupervisionDTO{ID: 1, Codigo: req.Codigo, Nombre: req.Nombre, AreaM2: req.AreaM2, ConGeom: true, Activo: true}, nil
}

func (m *mockZonaUC) Actualizar(_ context.Context, codigo string, req dto.ActualizarZonaSupervisionDTO, _ *int64) (dto.ZonaSupervisionDTO, error) {
	if m.err != nil {
		return dto.ZonaSupervisionDTO{}, m.err
	}
	return dto.ZonaSupervisionDTO{ID: 1, Codigo: codigo, Nombre: req.Nombre, AreaM2: req.AreaM2, ConGeom: true, Activo: true}, nil
}

func (m *mockZonaUC) Baja(_ context.Context, codigo string, _ *int64) (dto.ZonaSupervisionDTO, error) {
	if m.err != nil {
		return dto.ZonaSupervisionDTO{}, m.err
	}
	return dto.ZonaSupervisionDTO{ID: 1, Codigo: codigo, Nombre: "Zona", Activo: false, ConGeom: true}, nil
}

type mockCuadrillaUC struct {
	items []dto.CuadrillaDTO
	err   error
}

func (m *mockCuadrillaUC) Listar(_ context.Context) ([]dto.CuadrillaDTO, error) {
	return m.items, m.err
}

func (m *mockCuadrillaUC) Crear(_ context.Context, req dto.CrearCuadrillaDTO) (dto.CuadrillaDTO, error) {
	if m.err != nil {
		return dto.CuadrillaDTO{}, m.err
	}
	return dto.CuadrillaDTO{ID: req.ID, NombreFicticio: req.Nombre, Turno: req.Turno, Activo: true}, nil
}

type mockLugarUC struct {
	items []dto.LugarDTO
	err   error
}

func (m *mockLugarUC) Listar(_ context.Context) ([]dto.LugarDTO, error) {
	return m.items, m.err
}

func (m *mockLugarUC) Crear(_ context.Context, req dto.CrearLugarDTO) (dto.LugarDTO, error) {
	if m.err != nil {
		return dto.LugarDTO{}, m.err
	}
	return dto.LugarDTO{ID: 1, Nombre: req.Nombre, NombreNorm: req.Nombre, Lat: req.Lat, Lon: req.Lon, ZonaSupervisionID: req.ZonaID, Activo: true}, nil
}

type mockEspecieUC struct {
	items []dto.EspecieDTO
	err   error
}

func (m *mockEspecieUC) Listar(_ context.Context) ([]dto.EspecieDTO, error) {
	return m.items, m.err
}

func (m *mockEspecieUC) Crear(_ context.Context, req dto.CrearEspecieDTO) (dto.EspecieDTO, error) {
	if m.err != nil {
		return dto.EspecieDTO{}, m.err
	}
	return dto.EspecieDTO{ID: 1, NombreCientifico: req.Cientifico, NombreComun: req.Comun, Activo: true}, nil
}

type mockEjemplarUC struct {
	paginated dto.EjemplaresPaginadosDTO
	codigos   []dto.CodigoHistoricoDTO
	err       error
}

func (m *mockEjemplarUC) Listar(_ context.Context, limit, offset int, _ string) (dto.EjemplaresPaginadosDTO, error) {
	if m.err != nil {
		return dto.EjemplaresPaginadosDTO{}, m.err
	}
	res := m.paginated
	res.Limit = limit
	res.Offset = offset
	return res, nil
}

func (m *mockEjemplarUC) Crear(_ context.Context, e dto.EjemplarDTO) (dto.EjemplarDTO, error) {
	if m.err != nil {
		return dto.EjemplarDTO{}, m.err
	}
	e.ID = 10
	e.Activo = true
	return e, nil
}

func (m *mockEjemplarUC) Actualizar(_ context.Context, id int64, req dto.ActualizarEjemplarDTO, _ *int64) (dto.EjemplarDTO, error) {
	if m.err != nil {
		return dto.EjemplarDTO{}, m.err
	}
	return dto.EjemplarDTO{ID: id, Codigo: "EJ-1", Salud: req.Salud, Activo: true}, nil
}

func (m *mockEjemplarUC) Recodificar(_ context.Context, id int64, req dto.RecodificarDTO, _ *int64) (dto.CodigoHistoricoDTO, error) {
	if m.err != nil {
		return dto.CodigoHistoricoDTO{}, m.err
	}
	return dto.CodigoHistoricoDTO{ID: 1, EjemplarID: id, CodigoAnterior: "OLD", CodigoNuevo: req.Codigo}, nil
}

func (m *mockEjemplarUC) ListarCodigos(_ context.Context, _ int64) ([]dto.CodigoHistoricoDTO, error) {
	return m.codigos, m.err
}

type mockRefUC struct {
	poligonos []dto.PoligonoCuadrillaDTO
	capa      []dto.CapaFichaDTO
	err       error
}

func (m *mockRefUC) ListarPoligonos(_ context.Context) ([]dto.PoligonoCuadrillaDTO, error) {
	return m.poligonos, m.err
}

func (m *mockRefUC) ListarCapa(_ context.Context, _ string) ([]dto.CapaFichaDTO, error) {
	return m.capa, m.err
}

func setupCatastroTest(
	zonaUC *mockZonaUC,
	cuadUC *mockCuadrillaUC,
	lugarUC *mockLugarUC,
	espUC *mockEspecieUC,
	ejUC *mockEjemplarUC,
	refUC *mockRefUC,
) (*gin.Engine, controller.ICatastroController) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	ctrl := controller.NewCatastroController(zonaUC, cuadUC, lugarUC, espUC, ejUC, refUC, zerolog.Nop())
	return engine, ctrl
}

func TestCatastroController_Endpoints(t *testing.T) {
	zonaUC := &mockZonaUC{items: []dto.ZonaSupervisionDTO{{ID: 1, Codigo: "Z1", Nombre: "Zona 1"}}}
	cuadUC := &mockCuadrillaUC{items: []dto.CuadrillaDTO{{ID: "C1", NombreFicticio: "Cuadrilla 1"}}}
	lugarUC := &mockLugarUC{items: []dto.LugarDTO{{ID: 1, Nombre: "Lugar 1"}}}
	espUC := &mockEspecieUC{items: []dto.EspecieDTO{{ID: 1, NombreCientifico: "Species 1"}}}
	ejUC := &mockEjemplarUC{
		paginated: dto.EjemplaresPaginadosDTO{Ejemplares: []dto.EjemplarDTO{{ID: 1, Codigo: "EJ-1"}}, Total: 1},
		codigos:   []dto.CodigoHistoricoDTO{{ID: 1, EjemplarID: 1, CodigoAnterior: "OLD", CodigoNuevo: "EJ-1"}},
	}
	refUC := &mockRefUC{
		poligonos: []dto.PoligonoCuadrillaDTO{{ID: 1, FeatureID: "POL-1"}},
		capa:      []dto.CapaFichaDTO{{ID: 1, FeatureID: "CAPA-1"}},
	}

	engine, ctrl := setupCatastroTest(zonaUC, cuadUC, lugarUC, espUC, ejUC, refUC)
	engine.GET("/zonas", ctrl.Zonas)
	engine.POST("/zonas", ctrl.CrearZona)
	engine.GET("/poligonos", ctrl.Poligonos)
	engine.GET("/cuadrillas", ctrl.Cuadrillas)
	engine.POST("/cuadrillas", ctrl.CrearCuadrilla)
	engine.GET("/lugares", ctrl.Lugares)
	engine.POST("/lugares", ctrl.CrearLugar)
	engine.GET("/especies", ctrl.Especies)
	engine.POST("/especies", ctrl.CrearEspecie)
	engine.GET("/ejemplares", ctrl.Ejemplares)
	engine.POST("/ejemplares", ctrl.CrearEjemplar)
	engine.GET("/ejemplares/:id/codigos", ctrl.Codigos)
	engine.POST("/ejemplares/:id/codigos", ctrl.Recodificar)
	engine.GET("/fauna", ctrl.Fauna)
	engine.GET("/puertas", ctrl.Puertas)
	engine.GET("/playas", ctrl.Playas)
	engine.GET("/veredas", ctrl.Veredas)
	engine.GET("/xerofiticas", ctrl.Xerofiticas)
	engine.GET("/jardines-reserva", ctrl.JardinesReserva)

	// 1. Zonas
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/zonas", nil))
	if w.Code != 200 {
		t.Fatalf("Zonas: expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/zonas", bytes.NewBufferString(`{"codigo":"Z1","nombre":"Zona 1","geojson":"{}"}`)))
	if w.Code != 201 {
		t.Fatalf("CrearZona: expected 201, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/zonas", bytes.NewBufferString(`{invalid-json`)))
	if w.Code != 400 {
		t.Fatalf("CrearZona invalid json: expected 400, got %d", w.Code)
	}

	// 2. Polígonos
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/poligonos", nil))
	if w.Code != 200 {
		t.Fatalf("Poligonos: expected 200, got %d", w.Code)
	}

	// 3. Cuadrillas
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/cuadrillas", nil))
	if w.Code != 200 {
		t.Fatalf("Cuadrillas: expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/cuadrillas", bytes.NewBufferString(`{"id":"C2","nombre_ficticio":"N2","turno":"tarde"}`)))
	if w.Code != 201 {
		t.Fatalf("CrearCuadrilla: expected 201, got %d", w.Code)
	}

	// 4. Lugares
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/lugares", nil))
	if w.Code != 200 {
		t.Fatalf("Lugares: expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/lugares", bytes.NewBufferString(`{"nombre":"P1","lat":-12.07,"lon":-77.08}`)))
	if w.Code != 201 {
		t.Fatalf("CrearLugar: expected 201, got %d", w.Code)
	}

	// 5. Especies
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/especies", nil))
	if w.Code != 200 {
		t.Fatalf("Especies: expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/especies", bytes.NewBufferString(`{"nombre_cientifico":"S2","nombre_comun":"C2"}`)))
	if w.Code != 201 {
		t.Fatalf("CrearEspecie: expected 201, got %d", w.Code)
	}

	// 6. Ejemplares
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ejemplares?limit=20&offset=40", nil))
	if w.Code != 200 {
		t.Fatalf("Ejemplares: expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/ejemplares", bytes.NewBufferString(`{"codigo":"EJ-2","tipo_vegetacion":"Árbol","cantidad":1}`)))
	if w.Code != 201 {
		t.Fatalf("CrearEjemplar: expected 201, got %d", w.Code)
	}

	// 7. Codigos y Recodificar
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ejemplares/1/codigos", nil))
	if w.Code != 200 {
		t.Fatalf("Codigos: expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ejemplares/bad/codigos", nil))
	if w.Code != 400 {
		t.Fatalf("Codigos bad id: expected 400, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/ejemplares/1/codigos", bytes.NewBufferString(`{"codigo_nuevo":"AV-NEW"}`)))
	if w.Code != 201 {
		t.Fatalf("Recodificar: expected 201, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/ejemplares/bad/codigos", bytes.NewBufferString(`{"codigo_nuevo":"AV-NEW"}`)))
	if w.Code != 400 {
		t.Fatalf("Recodificar bad id: expected 400, got %d", w.Code)
	}

	// 8. 6 Capas de referencia
	for _, path := range []string{"/fauna", "/puertas", "/playas", "/veredas", "/xerofiticas", "/jardines-reserva"} {
		w = httptest.NewRecorder()
		engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != 200 {
			t.Fatalf("%s: expected 200, got %d", path, w.Code)
		}
	}
}

func TestCatastroController_ErrorPaths(t *testing.T) {
	errGeneric := errors.New("boom")
	zonaUC := &mockZonaUC{err: errGeneric}
	cuadUC := &mockCuadrillaUC{err: errGeneric}
	lugarUC := &mockLugarUC{err: errGeneric}
	espUC := &mockEspecieUC{err: errGeneric}
	ejUC := &mockEjemplarUC{err: domainErrors.ErrNoEncontrado}
	refUC := &mockRefUC{err: errGeneric}

	engine, ctrl := setupCatastroTest(zonaUC, cuadUC, lugarUC, espUC, ejUC, refUC)
	engine.GET("/zonas", ctrl.Zonas)
	engine.POST("/zonas", ctrl.CrearZona)
	engine.GET("/poligonos", ctrl.Poligonos)
	engine.GET("/cuadrillas", ctrl.Cuadrillas)
	engine.POST("/cuadrillas", ctrl.CrearCuadrilla)
	engine.GET("/lugares", ctrl.Lugares)
	engine.POST("/lugares", ctrl.CrearLugar)
	engine.GET("/especies", ctrl.Especies)
	engine.POST("/especies", ctrl.CrearEspecie)
	engine.GET("/ejemplares", ctrl.Ejemplares)
	engine.POST("/ejemplares", ctrl.CrearEjemplar)
	engine.GET("/ejemplares/:id/codigos", ctrl.Codigos)
	engine.POST("/ejemplares/:id/codigos", ctrl.Recodificar)
	engine.GET("/fauna", ctrl.Fauna)

	// List 500s
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/zonas", nil))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/poligonos", nil))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/cuadrillas", nil))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/lugares", nil))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/especies", nil))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ejemplares", nil))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ejemplares/1/codigos", nil))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/fauna", nil))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	// Recodificar 404
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/ejemplares/1/codigos", bytes.NewBufferString(`{"codigo_nuevo":"AV-NEW"}`)))
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "ejemplar sin código anterior" {
		t.Fatalf("expected 'ejemplar sin código anterior', got %q", resp["error"])
	}
}

func TestCatastroController_ActualizarYBajaZona(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := controller.NewCatastroController(&mockZonaUC{}, &mockCuadrillaUC{}, &mockLugarUC{}, &mockEspecieUC{}, &mockEjemplarUC{}, &mockRefUC{}, zerolog.Nop())
	r := gin.New()
	r.PATCH("/api/v1/catastro/zonas-supervision/:codigo", ctrl.ActualizarZona)
	r.PATCH("/areas-verdes/v1/catastro/zonas-supervision/:codigo", ctrl.ActualizarZona)
	r.POST("/api/v1/catastro/zonas-supervision/:codigo/baja", ctrl.BajaZona)
	r.POST("/areas-verdes/v1/catastro/zonas-supervision/:codigo/baja", ctrl.BajaZona)

	for _, path := range []string{
		"/api/v1/catastro/zonas-supervision/Z1",
		"/areas-verdes/v1/catastro/zonas-supervision/Z1",
	} {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"codigo":"Z1","nombre":"Zona Norte","area_m2":1200,"geom":{"type":"MultiPolygon","coordinates":[]}}`)
		req := httptest.NewRequest(http.MethodPatch, path, body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s status %d body %s", path, w.Code, w.Body.Bytes())
		}
		var got map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got["codigo"] != "Z1" || got["nombre"] != "Zona Norte" || got["activo"] != true {
			t.Fatalf("%s body %s", path, w.Body.Bytes())
		}
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/api/v1/catastro/zonas-supervision/Z1", bytes.NewBufferString(`{"codigo":"Z2","nombre":"Otra"}`)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("código distinto: status %d", w.Code)
	}

	ctrlMissing := controller.NewCatastroController(&mockZonaUC{err: domainErrors.ErrNoEncontrado}, &mockCuadrillaUC{}, &mockLugarUC{}, &mockEspecieUC{}, &mockEjemplarUC{}, &mockRefUC{}, zerolog.Nop())
	rMissing := gin.New()
	rMissing.POST("/api/v1/catastro/zonas-supervision/:codigo/baja", ctrlMissing.BajaZona)
	w = httptest.NewRecorder()
	rMissing.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/catastro/zonas-supervision/Z9/baja", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("baja inexistente: status %d", w.Code)
	}

	ctrlOK := controller.NewCatastroController(&mockZonaUC{}, &mockCuadrillaUC{}, &mockLugarUC{}, &mockEspecieUC{}, &mockEjemplarUC{}, &mockRefUC{}, zerolog.Nop())
	rOK := gin.New()
	rOK.POST("/areas-verdes/v1/catastro/zonas-supervision/:codigo/baja", ctrlOK.BajaZona)
	w = httptest.NewRecorder()
	rOK.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/areas-verdes/v1/catastro/zonas-supervision/Z1/baja", bytes.NewBufferString(`{}`)))
	if w.Code != http.StatusOK {
		t.Fatalf("baja status %d body %s", w.Code, w.Body.Bytes())
	}
	var baja map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &baja); err != nil {
		t.Fatal(err)
	}
	if baja["activo"] != false || baja["codigo"] != "Z1" {
		t.Fatalf("baja body %s", w.Body.Bytes())
	}
}
