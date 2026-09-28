package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/services"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	domainEntities "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/ratelimit"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/routes"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/routes/groups"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

type mockSaludUC struct{}

func (mockSaludUC) VerificarSalud(_ context.Context) (*dto.EstadoSaludDTO, error) {
	return &dto.EstadoSaludDTO{Database: "up", PostGIS: "3.5"}, nil
}

type mockSesionRoutesUC struct{}

func (mockSesionRoutesUC) Login(_ context.Context, _, _ string) (string, *dto.UsuarioSesionDTO, error) {
	return "token", &dto.UsuarioSesionDTO{}, nil
}

func (mockSesionRoutesUC) Logout(_ context.Context, _ string) {
}

func (mockSesionRoutesUC) Resolver(_ context.Context, token string) (*dto.UsuarioSesionDTO, error) {
	switch token {
	case "token-admin":
		return &dto.UsuarioSesionDTO{ID: 1, Usuario: "admin", Rol: enums.RolAdmin.String(), RolNombre: "Administrador"}, nil
	case "token-coordinacion":
		return &dto.UsuarioSesionDTO{ID: 2, Usuario: "coordinacion", Rol: enums.RolCoordinacion.String(), RolNombre: "Ingeniería/Coordinación"}, nil
	case "token-jefatura":
		return &dto.UsuarioSesionDTO{ID: 3, Usuario: "jefatura", Rol: enums.RolJefatura.String(), RolNombre: "Jefatura"}, nil
	case "token-norte":
		return &dto.UsuarioSesionDTO{ID: 4, Usuario: "norte", Rol: enums.RolCapataz.String(), RolNombre: "Capataz", CapatazID: "cap-norte"}, nil
	default:
		return nil, errors.New("sin sesion")
	}
}

type mockUsuarioRoutesUC struct{}

func (mockUsuarioRoutesUC) ListarUsuarios(_ context.Context) (*dto.UsuariosResponseDTO, error) {
	return &dto.UsuariosResponseDTO{
		Usuarios: []dto.UsuarioSesionDTO{
			{ID: 1, Usuario: "admin", Rol: enums.RolAdmin.String()},
		},
		Permisos: []dto.PermisoDTO{},
		Aviso:    "aviso",
	}, nil
}

type mockContratoOpenAPI struct{}

func (mockContratoOpenAPI) ObtenerContrato() ([]byte, error) {
	return []byte("openapi: 3.0.0"), nil
}

type mockCatalogoRoutesUC struct{}

func (mockCatalogoRoutesUC) Listar(_ context.Context, _ dto.FiltroCatalogoDTO) (*dto.CatalogoListResponseDTO, error) {
	return &dto.CatalogoListResponseDTO{
		Items: []dto.CatalogoItemDTO{
			{ID: 1, Clase: "estado", Codigo: "pendiente", Nombre: "Pendiente", Activo: true, Orden: 1},
		},
		Clases: enums.ClasesCatalogoValidas(),
	}, nil
}

func (mockCatalogoRoutesUC) Crear(_ context.Context, in dto.CrearCatalogoDTO) (*dto.CatalogoItemDTO, error) {
	return &dto.CatalogoItemDTO{
		ID:     10,
		Clase:  in.Clase,
		Codigo: in.Codigo,
		Nombre: in.Nombre,
		Activo: true,
		Orden:  0,
	}, nil
}

func (mockCatalogoRoutesUC) Desactivar(_ context.Context, id int64) (*dto.DesactivarCatalogoResponseDTO, error) {
	return &dto.DesactivarCatalogoResponseDTO{
		Activo: false,
		ID:     id,
	}, nil
}

type mockGeoRoutesUC struct{}

func (mockGeoRoutesUC) Areas(_ context.Context, _ dto.FiltroGeoDTO) (domainEntities.FeatureCollection, error) {
	return domainEntities.Collection("areas_verdes"), nil
}

func (mockGeoRoutesUC) Zonas(_ context.Context, _ dto.FiltroGeoDTO) (domainEntities.FeatureCollection, error) {
	return domainEntities.Collection("zonas"), nil
}

func (mockGeoRoutesUC) Capa(_ context.Context, capa string, _ dto.FiltroGeoDTO) (domainEntities.FeatureCollection, error) {
	return domainEntities.Collection(capa), nil
}

func (mockGeoRoutesUC) Capas(_ context.Context) (dto.CapasIndexDTO, error) {
	return dto.CapasIndexDTO{CapasConocidas: []string{"jardines_reserva", "xerofitica"}, Cargadas: []dto.CapaCountDTO{}}, nil
}

func (mockGeoRoutesUC) Resumen(_ context.Context) (dto.ResumenDTO, error) {
	return dto.ResumenDTO{CRS: "EPSG:4326", Areas: 521, Zonas: 534}, nil
}

func (mockGeoRoutesUC) Edificios(_ context.Context) ([]byte, error) {
	return []byte(`{"type":"FeatureCollection","name":"edificios","features":[]}`), nil
}

type mockAreaVerdeRoutesUC struct{}

func (mockAreaVerdeRoutesUC) Fichas(_ context.Context, _ string) ([]dto.FichaDTO, error) {
	return []dto.FichaDTO{
		{FeatureID: "AV-0001", Nombre: "Área 1", ConGeom: true},
	}, nil
}

func (mockAreaVerdeRoutesUC) ActualizarFicha(_ context.Context, id string, req dto.ActualizarFichaDTO, _ *int64) (*dto.FichaDTO, error) {
	return &dto.FichaDTO{FeatureID: id, Nombre: req.Nombre, Uso: req.Uso, ConGeom: true}, nil
}

func (mockAreaVerdeRoutesUC) CrearSinGeom(_ context.Context, req dto.CrearAreaSinGeomDTO, _ *int64) (*dto.FichaDTO, error) {
	return &dto.FichaDTO{FeatureID: req.FeatureID, Nombre: req.Nombre, Uso: req.Uso, ConGeom: false}, nil
}

type mockZonaRoutesUC struct{}

func (mockZonaRoutesUC) Listar(_ context.Context) ([]dto.ZonaSupervisionDTO, error) {
	return []dto.ZonaSupervisionDTO{{ID: 1, Codigo: "Z1", Nombre: "Zona 1", ConGeom: true, Activo: true}}, nil
}

func (mockZonaRoutesUC) Crear(_ context.Context, req dto.CrearZonaSupervisionDTO) (dto.ZonaSupervisionDTO, error) {
	return dto.ZonaSupervisionDTO{ID: 1, Codigo: req.Codigo, Nombre: req.Nombre, ConGeom: true, Activo: true}, nil
}

type mockCuadrillaRoutesUC struct{}

func (mockCuadrillaRoutesUC) Listar(_ context.Context) ([]dto.CuadrillaDTO, error) {
	return []dto.CuadrillaDTO{{ID: "C1", NombreFicticio: "Equipo 1", Turno: "manana", Activo: true}}, nil
}

func (mockCuadrillaRoutesUC) Crear(_ context.Context, req dto.CrearCuadrillaDTO) (dto.CuadrillaDTO, error) {
	return dto.CuadrillaDTO{ID: req.ID, NombreFicticio: req.Nombre, Turno: req.Turno, Activo: true}, nil
}

type mockLugarRoutesUC struct{}

func (mockLugarRoutesUC) Listar(_ context.Context) ([]dto.LugarDTO, error) {
	return []dto.LugarDTO{{ID: 1, Nombre: "Lugar 1", Lat: -12.07, Lon: -77.08, Activo: true}}, nil
}

func (mockLugarRoutesUC) Crear(_ context.Context, req dto.CrearLugarDTO) (dto.LugarDTO, error) {
	return dto.LugarDTO{ID: 1, Nombre: req.Nombre, Lat: req.Lat, Lon: req.Lon, Activo: true}, nil
}

type mockEspecieRoutesUC struct{}

func (mockEspecieRoutesUC) Listar(_ context.Context) ([]dto.EspecieDTO, error) {
	return []dto.EspecieDTO{{ID: 1, NombreCientifico: "Species 1", NombreComun: "Comun 1", Activo: true}}, nil
}

func (mockEspecieRoutesUC) Crear(_ context.Context, req dto.CrearEspecieDTO) (dto.EspecieDTO, error) {
	return dto.EspecieDTO{ID: 1, NombreCientifico: req.Cientifico, NombreComun: req.Comun, Activo: true}, nil
}

type mockEjemplarRoutesUC struct{}

func (mockEjemplarRoutesUC) Listar(_ context.Context, limit, offset int) (dto.EjemplaresPaginadosDTO, error) {
	return dto.EjemplaresPaginadosDTO{
		Ejemplares: []dto.EjemplarDTO{{ID: 1, Codigo: "EJ-1", Activo: true}},
		Total:      1,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (mockEjemplarRoutesUC) Crear(_ context.Context, req dto.EjemplarDTO) (dto.EjemplarDTO, error) {
	req.ID = 10
	req.Activo = true
	return req, nil
}

func (mockEjemplarRoutesUC) Recodificar(_ context.Context, id int64, req dto.RecodificarDTO) (dto.CodigoHistoricoDTO, error) {
	return dto.CodigoHistoricoDTO{ID: 1, EjemplarID: id, CodigoAnterior: "OLD", CodigoNuevo: req.Codigo}, nil
}

func (mockEjemplarRoutesUC) ListarCodigos(_ context.Context, id int64) ([]dto.CodigoHistoricoDTO, error) {
	return []dto.CodigoHistoricoDTO{{ID: 1, EjemplarID: id, CodigoAnterior: "OLD", CodigoNuevo: "EJ-1"}}, nil
}

type mockCatastroRefRoutesUC struct{}

func (mockCatastroRefRoutesUC) ListarPoligonos(_ context.Context) ([]dto.PoligonoCuadrillaDTO, error) {
	return []dto.PoligonoCuadrillaDTO{{ID: 1, FeatureID: "POL-1", Codigo: "P1", Nombre: "Poligono 1", Activo: true}}, nil
}

func (mockCatastroRefRoutesUC) ListarCapa(_ context.Context, _ string) ([]dto.CapaFichaDTO, error) {
	return []dto.CapaFichaDTO{{ID: 1, FeatureID: "F-1", Activo: true}}, nil
}

func setupTestRouter(swaggerEnabled bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	_ = engine.SetTrustedProxies([]string{})

	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:    "8080",
			GinMode: "release",
		},
		Swagger: config.SwaggerConfig{
			Enabled: swaggerEnabled,
		},
		Seguridad: config.SeguridadConfig{
			CORSOrigins:    []string{"*"},
			CookieSecure:   false,
			CookieSameSite: "lax",
		},
	}

	healthCtrl := controller.NewHealthController(mockSaludUC{})
	metaCtrl := controller.NewMetaController(mockContratoOpenAPI{})
	sesionCtrl := controller.NewSesionController(mockSesionRoutesUC{})
	usuarioCtrl := controller.NewUsuarioController(mockUsuarioRoutesUC{})
	catalogoCtrl := controller.NewCatalogoController(mockCatalogoRoutesUC{})
	geoCtrl := controller.NewGeoController(mockGeoRoutesUC{})
	areaVerdeCtrl := controller.NewAreaVerdeController(mockAreaVerdeRoutesUC{})
	catastroCtrl := controller.NewCatastroController(
		mockZonaRoutesUC{},
		mockCuadrillaRoutesUC{},
		mockLugarRoutesUC{},
		mockEspecieRoutesUC{},
		mockEjemplarRoutesUC{},
		mockCatastroRefRoutesUC{},
	)

	permisosSvc := services.NewPermisosService()
	limitador := ratelimit.NewMemoriaLimitador(100, time.Minute)

	healthGrp := groups.NewHealthGroup(healthCtrl)
	metaGrp := groups.NewMetaGroup(metaCtrl)
	legadoGrp := groups.NewLegadoGroup(healthCtrl)
	swaggerGrp := groups.NewSwaggerGroup()
	sesionGrp := groups.NewSesionGroup(sesionCtrl)
	accesosGrp := groups.NewAccesosGroup(usuarioCtrl)
	catalogoGrp := groups.NewCatalogoGroup(catalogoCtrl, permisosSvc)
	geoGrp := groups.NewGeoGroup(geoCtrl, permisosSvc)
	catastroGrp := groups.NewCatastroGroup(areaVerdeCtrl, catastroCtrl, permisosSvc)

	r := routes.NewRouter(routes.RouterParams{
		Engine:        engine,
		Config:        cfg,
		Logger:        zerolog.Nop(),
		Limitador:     limitador,
		SesionUC:      mockSesionRoutesUC{},
		HealthGroup:   healthGrp,
		MetaGroup:     metaGrp,
		LegadoGroup:   legadoGrp,
		SwaggerGroup:  swaggerGrp,
		SesionGroup:   sesionGrp,
		AccesosGroup:  accesosGrp,
		CatalogoGroup: catalogoGrp,
		GeoGroup:      geoGrp,
		CatastroGroup: catastroGrp,
	})
	r.Setup()
	return engine
}

func TestNoRouteRetorna404ConFormatoEsperado(t *testing.T) {
	engine := setupTestRouter(true)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ruta-que-no-existe", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba 404, se obtuvo %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("error leyendo respuesta JSON: %v", err)
	}
	if body["error"] != "ruta no encontrada" {
		t.Fatalf("mensaje esperado 'ruta no encontrada', obtenido %q", body["error"])
	}
}

func TestSwaggerCondicionadoPorConfiguracion(t *testing.T) {
	// 1. Con SWAGGER_ENABLED = false, swagger no debe estar montado -> 404
	engineSinSwagger := setupTestRouter(false)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/areas-verdes/v1/swagger/index.html", nil)
	engineSinSwagger.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba 404 con swagger deshabilitado, se obtuvo %d", w.Code)
	}
	var bodySinSwagger map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &bodySinSwagger)
	if bodySinSwagger["error"] != "ruta no encontrada" {
		t.Fatalf("se esperaba error 'ruta no encontrada', obtenido %q", bodySinSwagger["error"])
	}

	// 2. Con SWAGGER_ENABLED = true, swagger debe responder (200 o redirect)
	engineConSwagger := setupTestRouter(true)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/areas-verdes/v1/swagger/index.html", nil)
	engineConSwagger.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatalf("con swagger habilitado no debe dar 404")
	}
}

func TestMontajeDoblePrefijos(t *testing.T) {
	engine := setupTestRouter(true)

	for _, ruta := range []string{
		"/areas-verdes/v1/health",
		"/areas-verdes/v1",
		"/api/v1",
		"/health",
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, ruta, nil)
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("se esperaba 200 en %s, obtenido %d", ruta, w.Code)
		}
	}

	// /api/v1/health debe dar 404 para paridad con la API anterior (decisión 17)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("se esperaba 404 en /api/v1/health, obtenido %d", w.Code)
	}
}

func TestRutasActualesRespondenIgual(t *testing.T) {
	engine := setupTestRouter(true)

	rutasCongeladas := []struct {
		metodo string
		ruta   string
		codigo int
	}{
		{http.MethodGet, "/health", 200},
		{http.MethodGet, "/api/v1", 200},
		{http.MethodGet, "/areas-verdes/v1", 200},
		{http.MethodGet, "/api/v1/openapi.yaml", 200},
		{http.MethodGet, "/areas-verdes/v1/openapi.yaml", 200},
		{http.MethodGet, "/areas-verdes/v1/health", 200},
		{http.MethodGet, "/api/v1/health", 404},
		{http.MethodGet, "/api/v1/sesion", 401},
		{http.MethodGet, "/areas-verdes/v1/sesion", 401},
		{http.MethodDelete, "/api/v1/sesion", 200},
		{http.MethodDelete, "/areas-verdes/v1/sesion", 200},
		// Accesos (sin sesión da 403 por paridad con API anterior)
		{http.MethodGet, "/api/v1/accesos/usuarios", 403},
		{http.MethodGet, "/areas-verdes/v1/accesos/usuarios", 403},
		{http.MethodGet, "/api/v1/catalogos", 401},
		{http.MethodGet, "/areas-verdes/v1/catalogos", 401},
		{http.MethodPost, "/api/v1/catalogos", 401},
		{http.MethodPost, "/areas-verdes/v1/catalogos", 401},
		{http.MethodPost, "/api/v1/catalogos/1/desactivar", 401},
		{http.MethodPost, "/areas-verdes/v1/catalogos/1/desactivar", 401},
		// Geo de lectura (Lote 8)
		{http.MethodGet, "/api/v1/geo/resumen", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/resumen", 401},
		{http.MethodGet, "/api/v1/geo/areas", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/areas", 401},
		{http.MethodGet, "/api/v1/geo/zonas", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/zonas", 401},
		{http.MethodGet, "/api/v1/geo/capas", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/capas", 401},
		{http.MethodGet, "/api/v1/geo/capas/jardines_reserva", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/capas/jardines_reserva", 401},
		{http.MethodGet, "/api/v1/geo/edificios", 401},
		{http.MethodGet, "/areas-verdes/v1/geo/edificios", 401},
		// Fichas de área (Lote 8)
		{http.MethodGet, "/api/v1/catastro/areas", 401},
		{http.MethodGet, "/areas-verdes/v1/catastro/areas", 401},
		{http.MethodPost, "/api/v1/catastro/areas", 401},
		{http.MethodPost, "/areas-verdes/v1/catastro/areas", 401},
		{http.MethodPatch, "/api/v1/catastro/areas/AV-0001", 401},
		{http.MethodPatch, "/areas-verdes/v1/catastro/areas/AV-0001", 401},
		{http.MethodGet, "/api/v1/no-existe", 404},
		{http.MethodGet, "/areas-verdes/v1/no-existe", 404},
	}

	vistas := map[string]bool{}
	for _, rt := range engine.Routes() {
		vistas[rt.Method+" "+rt.Path] = true
	}

	for _, want := range rutasCongeladas {
		if want.codigo == 404 {
			continue
		}
		path := want.ruta
		switch path {
		case "/api/v1/catalogos/1/desactivar":
			path = "/api/v1/catalogos/:id/desactivar"
		case "/areas-verdes/v1/catalogos/1/desactivar":
			path = "/areas-verdes/v1/catalogos/:id/desactivar"
		case "/api/v1/geo/capas/jardines_reserva":
			path = "/api/v1/geo/capas/:capa"
		case "/areas-verdes/v1/geo/capas/jardines_reserva":
			path = "/areas-verdes/v1/geo/capas/:capa"
		case "/api/v1/catastro/areas/AV-0001":
			path = "/api/v1/catastro/areas/:id"
		case "/areas-verdes/v1/catastro/areas/AV-0001":
			path = "/areas-verdes/v1/catastro/areas/:id"
		}
		key := want.metodo + " " + path
		if !vistas[key] {
			t.Errorf("falta la ruta registrada %s", key)
		}
	}

	for _, want := range rutasCongeladas {
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, httptest.NewRequest(want.metodo, want.ruta, nil))
		if w.Code != want.codigo {
			t.Errorf("%s %s -> %d, se esperaba %d (%s)", want.metodo, want.ruta, w.Code, want.codigo, w.Body.String())
		}
	}
}

func TestRutasCatalogos_SinAutenticacionDa401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/catalogos"},
		{http.MethodPost, "/api/v1/catalogos"},
		{http.MethodPost, "/api/v1/catalogos/1/desactivar"},
		{http.MethodGet, "/areas-verdes/v1/catalogos"},
		{http.MethodPost, "/areas-verdes/v1/catalogos"},
		{http.MethodPost, "/areas-verdes/v1/catalogos/1/desactivar"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasCatalogos_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	type caso struct {
		metodo         string
		ruta           string
		token          string
		body           string
		statusEsperado int
		errorEsperado  string
	}

	casos := []caso{
		// Capataz (norte): consultar=true, catalogos=false
		{http.MethodGet, "/api/v1/catalogos", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/catalogos", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catalogos", "token-norte", `{"clase":"estado","codigo":"test","nombre":"Test"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/catalogos", "token-norte", `{"clase":"estado","codigo":"test","nombre":"Test"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catalogos/1/desactivar", "token-norte", "", 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/catalogos/1/desactivar", "token-norte", "", 403, "su rol no tiene ese permiso"},

		// Coordinación: consultar=true, catalogos=false
		{http.MethodGet, "/api/v1/catalogos", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/catalogos", "token-coordinacion", `{"clase":"estado","codigo":"test","nombre":"Test"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catalogos/1/desactivar", "token-coordinacion", "", 403, "su rol no tiene ese permiso"},

		// Jefatura: consultar=true, catalogos=false
		{http.MethodGet, "/api/v1/catalogos", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/catalogos", "token-jefatura", `{"clase":"estado","codigo":"test","nombre":"Test"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catalogos/1/desactivar", "token-jefatura", "", 403, "su rol no tiene ese permiso"},

		// Admin: consultar=true, catalogos=true
		{http.MethodGet, "/api/v1/catalogos", "token-admin", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/catalogos", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/catalogos", "token-admin", `{"clase":"estado","codigo":"nuevo","nombre":"Nuevo"}`, 201, ""},
		{http.MethodPost, "/areas-verdes/v1/catalogos", "token-admin", `{"clase":"estado","codigo":"nuevo","nombre":"Nuevo"}`, 201, ""},
		{http.MethodPost, "/api/v1/catalogos/1/desactivar", "token-admin", "", 200, ""},
		{http.MethodPost, "/areas-verdes/v1/catalogos/1/desactivar", "token-admin", "", 200, ""},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		var req *http.Request
		if c.body != "" {
			req = httptest.NewRequest(c.metodo, c.ruta, bytes.NewBufferString(c.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(c.metodo, c.ruta, nil)
		}
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
		if c.errorEsperado != "" {
			var body map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["error"] != c.errorEsperado {
				t.Errorf("%s %s (token=%s): error esperado %q, obtenido %q", c.metodo, c.ruta, c.token, c.errorEsperado, body["error"])
			}
		}
	}
}

func TestRutasGeoYCatastro_SinAutenticacionDa401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/geo/resumen"},
		{http.MethodGet, "/areas-verdes/v1/geo/resumen"},
		{http.MethodGet, "/api/v1/geo/areas"},
		{http.MethodGet, "/areas-verdes/v1/geo/areas"},
		{http.MethodGet, "/api/v1/geo/zonas"},
		{http.MethodGet, "/areas-verdes/v1/geo/zonas"},
		{http.MethodGet, "/api/v1/geo/capas"},
		{http.MethodGet, "/areas-verdes/v1/geo/capas"},
		{http.MethodGet, "/api/v1/geo/capas/jardines_reserva"},
		{http.MethodGet, "/areas-verdes/v1/geo/capas/jardines_reserva"},
		{http.MethodGet, "/api/v1/geo/edificios"},
		{http.MethodGet, "/areas-verdes/v1/geo/edificios"},
		{http.MethodGet, "/api/v1/catastro/areas"},
		{http.MethodGet, "/areas-verdes/v1/catastro/areas"},
		{http.MethodPost, "/api/v1/catastro/areas"},
		{http.MethodPost, "/areas-verdes/v1/catastro/areas"},
		{http.MethodPatch, "/api/v1/catastro/areas/AV-0001"},
		{http.MethodPatch, "/areas-verdes/v1/catastro/areas/AV-0001"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasGeoYCatastro_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	type caso struct {
		metodo         string
		ruta           string
		token          string
		body           string
		statusEsperado int
		errorEsperado  string
	}

	casos := []caso{
		// Capataz (norte): consultar=true, registrar=true
		{http.MethodGet, "/api/v1/geo/resumen", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/geo/resumen", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/geo/areas", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/areas", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/areas", "token-norte", `{"nombre":"Nueva","uso":"jardín"}`, 201, ""},
		{http.MethodPatch, "/api/v1/catastro/areas/AV-0001", "token-norte", `{"nombre":"Modif","uso":"jardín"}`, 200, ""},

		// Coordinación: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/geo/zonas", "token-coordinacion", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/geo/zonas", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/areas", "token-coordinacion", `{"nombre":"Nueva","uso":"jardín"}`, 201, ""},
		{http.MethodPatch, "/api/v1/catastro/areas/AV-0001", "token-coordinacion", `{"nombre":"Modif","uso":"jardín"}`, 200, ""},

		// Jefatura: consultar=true, registrar=false
		{http.MethodGet, "/api/v1/geo/capas", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/geo/capas", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/areas", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/areas", "token-jefatura", `{"nombre":"Nueva","uso":"jardín"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/catastro/areas", "token-jefatura", `{"nombre":"Nueva","uso":"jardín"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/api/v1/catastro/areas/AV-0001", "token-jefatura", `{"nombre":"Modif","uso":"jardín"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPatch, "/areas-verdes/v1/catastro/areas/AV-0001", "token-jefatura", `{"nombre":"Modif","uso":"jardín"}`, 403, "su rol no tiene ese permiso"},

		// Admin: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/geo/edificios", "token-admin", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/geo/edificios", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/areas", "token-admin", `{"nombre":"Nueva","uso":"jardín"}`, 201, ""},
		{http.MethodPatch, "/api/v1/catastro/areas/AV-0001", "token-admin", `{"nombre":"Modif","uso":"jardín"}`, 200, ""},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		var req *http.Request
		if c.body != "" {
			req = httptest.NewRequest(c.metodo, c.ruta, bytes.NewBufferString(c.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(c.metodo, c.ruta, nil)
		}
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
		if c.errorEsperado != "" {
			var body map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["error"] != c.errorEsperado {
				t.Errorf("%s %s (token=%s): error esperado %q, obtenido %q", c.metodo, c.ruta, c.token, c.errorEsperado, body["error"])
			}
		}
	}
}

func TestRutasCatastroMaestro_SinAutenticacionDa401(t *testing.T) {
	engine := setupTestRouter(true)

	rutas := []struct {
		metodo string
		ruta   string
	}{
		{http.MethodGet, "/api/v1/catastro/zonas-supervision"},
		{http.MethodGet, "/areas-verdes/v1/catastro/zonas-supervision"},
		{http.MethodPost, "/api/v1/catastro/zonas-supervision"},
		{http.MethodPost, "/areas-verdes/v1/catastro/zonas-supervision"},

		{http.MethodGet, "/api/v1/catastro/poligonos"},
		{http.MethodGet, "/areas-verdes/v1/catastro/poligonos"},

		{http.MethodGet, "/api/v1/catastro/cuadrillas"},
		{http.MethodGet, "/areas-verdes/v1/catastro/cuadrillas"},
		{http.MethodPost, "/api/v1/catastro/cuadrillas"},
		{http.MethodPost, "/areas-verdes/v1/catastro/cuadrillas"},

		{http.MethodGet, "/api/v1/catastro/lugares"},
		{http.MethodGet, "/areas-verdes/v1/catastro/lugares"},
		{http.MethodPost, "/api/v1/catastro/lugares"},
		{http.MethodPost, "/areas-verdes/v1/catastro/lugares"},

		{http.MethodGet, "/api/v1/catastro/especies"},
		{http.MethodGet, "/areas-verdes/v1/catastro/especies"},
		{http.MethodPost, "/api/v1/catastro/especies"},
		{http.MethodPost, "/areas-verdes/v1/catastro/especies"},

		{http.MethodGet, "/api/v1/catastro/ejemplares"},
		{http.MethodGet, "/areas-verdes/v1/catastro/ejemplares"},
		{http.MethodPost, "/api/v1/catastro/ejemplares"},
		{http.MethodPost, "/areas-verdes/v1/catastro/ejemplares"},

		{http.MethodGet, "/api/v1/catastro/ejemplares/1/codigos"},
		{http.MethodGet, "/areas-verdes/v1/catastro/ejemplares/1/codigos"},
		{http.MethodPost, "/api/v1/catastro/ejemplares/1/codigos"},
		{http.MethodPost, "/areas-verdes/v1/catastro/ejemplares/1/codigos"},

		{http.MethodGet, "/api/v1/catastro/fauna"},
		{http.MethodGet, "/areas-verdes/v1/catastro/fauna"},
		{http.MethodGet, "/api/v1/catastro/puertas"},
		{http.MethodGet, "/areas-verdes/v1/catastro/puertas"},
		{http.MethodGet, "/api/v1/catastro/playas"},
		{http.MethodGet, "/areas-verdes/v1/catastro/playas"},
		{http.MethodGet, "/api/v1/catastro/veredas"},
		{http.MethodGet, "/areas-verdes/v1/catastro/veredas"},
		{http.MethodGet, "/api/v1/catastro/xerofiticas"},
		{http.MethodGet, "/areas-verdes/v1/catastro/xerofiticas"},
		{http.MethodGet, "/api/v1/catastro/jardines-reserva"},
		{http.MethodGet, "/areas-verdes/v1/catastro/jardines-reserva"},
	}

	for _, r := range rutas {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(r.metodo, r.ruta, bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin auth: esperado 401, obtenido %d", r.metodo, r.ruta, w.Code)
		}
		if w.Body.String() != `{"error":"inicie sesión"}` {
			t.Errorf("%s %s sin auth: cuerpo inesperado %s", r.metodo, r.ruta, w.Body.String())
		}
	}
}

func TestRutasCatastroMaestro_PermisosPorRol(t *testing.T) {
	engine := setupTestRouter(true)

	type caso struct {
		metodo         string
		ruta           string
		token          string
		body           string
		statusEsperado int
		errorEsperado  string
	}

	casos := []caso{
		// 1. Capataz (norte): consultar=true, registrar=true
		{http.MethodGet, "/api/v1/catastro/zonas-supervision", "token-norte", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/catastro/zonas-supervision", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/zonas-supervision", "token-norte", `{"codigo":"Z1","nombre":"Zona 1"}`, 201, ""},
		{http.MethodGet, "/api/v1/catastro/poligonos", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/cuadrillas", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/cuadrillas", "token-norte", `{"id":"C1","nombre_ficticio":"N1","turno":"manana"}`, 201, ""},
		{http.MethodGet, "/api/v1/catastro/lugares", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/lugares", "token-norte", `{"nombre":"L1","lat":-12.07,"lon":-77.08}`, 201, ""},
		{http.MethodGet, "/api/v1/catastro/especies", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/especies", "token-norte", `{"nombre_cientifico":"S1","nombre_comun":"C1"}`, 201, ""},
		{http.MethodGet, "/api/v1/catastro/ejemplares", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/ejemplares", "token-norte", `{"codigo":"EJ-1","tipo_vegetacion":"Árbol","cantidad":1}`, 201, ""},
		{http.MethodGet, "/api/v1/catastro/ejemplares/1/codigos", "token-norte", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/ejemplares/1/codigos", "token-norte", `{"codigo_nuevo":"AV-NEW"}`, 201, ""},
		{http.MethodGet, "/api/v1/catastro/fauna", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/puertas", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/playas", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/veredas", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/xerofiticas", "token-norte", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/jardines-reserva", "token-norte", "", 200, ""},

		// 2. Coordinación: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/catastro/zonas-supervision", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/zonas-supervision", "token-coordinacion", `{"codigo":"Z1","nombre":"Zona 1"}`, 201, ""},
		{http.MethodGet, "/api/v1/catastro/ejemplares", "token-coordinacion", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/ejemplares", "token-coordinacion", `{"codigo":"EJ-1","tipo_vegetacion":"Árbol","cantidad":1}`, 201, ""},
		{http.MethodPost, "/api/v1/catastro/ejemplares/1/codigos", "token-coordinacion", `{"codigo_nuevo":"AV-NEW"}`, 201, ""},

		// 3. Admin: consultar=true, registrar=true
		{http.MethodGet, "/api/v1/catastro/zonas-supervision", "token-admin", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/zonas-supervision", "token-admin", `{"codigo":"Z1","nombre":"Zona 1"}`, 201, ""},
		{http.MethodPost, "/api/v1/catastro/cuadrillas", "token-admin", `{"id":"C1","nombre_ficticio":"N1","turno":"manana"}`, 201, ""},
		{http.MethodPost, "/api/v1/catastro/lugares", "token-admin", `{"nombre":"L1","lat":-12.07,"lon":-77.08}`, 201, ""},
		{http.MethodPost, "/api/v1/catastro/especies", "token-admin", `{"nombre_cientifico":"S1","nombre_comun":"C1"}`, 201, ""},
		{http.MethodPost, "/api/v1/catastro/ejemplares", "token-admin", `{"codigo":"EJ-1","tipo_vegetacion":"Árbol","cantidad":1}`, 201, ""},
		{http.MethodPost, "/api/v1/catastro/ejemplares/1/codigos", "token-admin", `{"codigo_nuevo":"AV-NEW"}`, 201, ""},

		// 4. Jefatura: consultar=true, registrar=false
		{http.MethodGet, "/api/v1/catastro/zonas-supervision", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/areas-verdes/v1/catastro/zonas-supervision", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/poligonos", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/cuadrillas", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/lugares", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/especies", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/ejemplares", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/ejemplares/1/codigos", "token-jefatura", "", 200, ""},
		{http.MethodGet, "/api/v1/catastro/fauna", "token-jefatura", "", 200, ""},
		{http.MethodPost, "/api/v1/catastro/zonas-supervision", "token-jefatura", `{"codigo":"Z1","nombre":"Zona 1"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/catastro/zonas-supervision", "token-jefatura", `{"codigo":"Z1","nombre":"Zona 1"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catastro/cuadrillas", "token-jefatura", `{"id":"C1","nombre_ficticio":"N1","turno":"manana"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catastro/lugares", "token-jefatura", `{"nombre":"L1","lat":-12.07,"lon":-77.08}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catastro/especies", "token-jefatura", `{"nombre_cientifico":"S1","nombre_comun":"C1"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catastro/ejemplares", "token-jefatura", `{"codigo":"EJ-1","tipo_vegetacion":"Árbol","cantidad":1}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/api/v1/catastro/ejemplares/1/codigos", "token-jefatura", `{"codigo_nuevo":"AV-NEW"}`, 403, "su rol no tiene ese permiso"},
		{http.MethodPost, "/areas-verdes/v1/catastro/ejemplares/1/codigos", "token-jefatura", `{"codigo_nuevo":"AV-NEW"}`, 403, "su rol no tiene ese permiso"},
	}

	for _, c := range casos {
		w := httptest.NewRecorder()
		var req *http.Request
		if c.body != "" {
			req = httptest.NewRequest(c.metodo, c.ruta, bytes.NewBufferString(c.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(c.metodo, c.ruta, nil)
		}
		req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: c.token})
		engine.ServeHTTP(w, req)

		if w.Code != c.statusEsperado {
			t.Errorf("%s %s (token=%s): esperado status %d, obtenido %d (%s)", c.metodo, c.ruta, c.token, c.statusEsperado, w.Code, w.Body.String())
		}
		if c.errorEsperado != "" {
			var body map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["error"] != c.errorEsperado {
				t.Errorf("%s %s (token=%s): error esperado %q, obtenido %q", c.metodo, c.ruta, c.token, c.errorEsperado, body["error"])
			}
		}
	}
}
