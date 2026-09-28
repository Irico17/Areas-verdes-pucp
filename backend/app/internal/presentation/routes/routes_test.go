package routes_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/routes"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/routes/groups"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func setupTestRouter(swaggerEnabled bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	_ = engine.SetTrustedProxies([]string{})

	cfg := config.New()
	cfg.Swagger.Enabled = swaggerEnabled

	healthCtrl := controller.NewHealthController()
	healthGrp := groups.NewHealthGroup(healthCtrl)
	swaggerGrp := groups.NewSwaggerGroup()

	r := routes.NewRouter(routes.RouterParams{
		Engine:       engine,
		Config:       cfg,
		Logger:       zerolog.Nop(),
		HealthGroup:  healthGrp,
		SwaggerGroup: swaggerGrp,
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
