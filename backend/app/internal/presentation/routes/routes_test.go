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

	permisosSvc := services.NewPermisosService()
	limitador := ratelimit.NewMemoriaLimitador(100, time.Minute)

	healthGrp := groups.NewHealthGroup(healthCtrl)
	metaGrp := groups.NewMetaGroup(metaCtrl)
	legadoGrp := groups.NewLegadoGroup(healthCtrl)
	swaggerGrp := groups.NewSwaggerGroup()
	sesionGrp := groups.NewSesionGroup(sesionCtrl)
	accesosGrp := groups.NewAccesosGroup(usuarioCtrl)
	catalogoGrp := groups.NewCatalogoGroup(catalogoCtrl, permisosSvc)

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
		if path == "/api/v1/catalogos/1/desactivar" {
			path = "/api/v1/catalogos/:id/desactivar"
		} else if path == "/areas-verdes/v1/catalogos/1/desactivar" {
			path = "/areas-verdes/v1/catalogos/:id/desactivar"
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
