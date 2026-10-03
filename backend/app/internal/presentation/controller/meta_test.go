package controller_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/archivos"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestMetaController_Index(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1", nil)

	ctrl := controller.NewMetaController(nil, zerolog.Nop())
	ctrl.Index(c)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, se obtuvo %d", w.Code)
	}

	var res struct {
		Servicio string             `json:"servicio"`
		Version  string             `json:"version"`
		CRS      string             `json:"crs"`
		Rutas    []controller.Route `json:"rutas"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}

	if res.Servicio != "campus-verde-api" || res.Version != "v1" || res.CRS != "EPSG:4326" {
		t.Fatalf("campos de servicio no coinciden: %+v", res)
	}
	if len(res.Rutas) < 30 {
		t.Fatalf("se esperaban al menos 30 rutas en RutasV1, se obtuvieron %d", len(res.Rutas))
	}
}

func TestMetaController_OpenAPI_Exitoso(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/openapi.yaml", nil)

	cfg := &config.Config{
		Datos: config.DatosConfig{
			OpenAPIPath: filepath.Join(testutil.FindRepoRoot(), "backend", "openapi.yaml"),
		},
	}
	adapter := archivos.NewContratoOpenAPIAdapter(cfg)
	ctrl := controller.NewMetaController(adapter, zerolog.Nop())
	ctrl.OpenAPI(c)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, se obtuvo %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/yaml; charset=utf-8" {
		t.Fatalf("Content-Type inesperado: %s", ct)
	}
	if w.Body.Len() == 0 {
		t.Fatal("el cuerpo de openapi.yaml está vacío")
	}
}

func TestMetaController_OpenAPI_NoDisponible(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/openapi.yaml", nil)

	cfg := &config.Config{
		Datos: config.DatosConfig{
			OpenAPIPath: "/ruta/inexistente/openapi.yaml",
		},
	}
	adapter := archivos.NewContratoOpenAPIAdapter(cfg)
	ctrl := controller.NewMetaController(adapter, zerolog.Nop())
	ctrl.OpenAPI(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba 404, se obtuvo %d", w.Code)
	}

	var res map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res["error"] != "openapi.yaml no disponible" {
		t.Fatalf("error inesperado: %s", res["error"])
	}
}

type mockContratoError struct{}

func (mockContratoError) ObtenerContrato() ([]byte, error) {
	return nil, errors.New("io error simulado")
}

func TestMetaController_OpenAPI_ErrorInterno(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/openapi.yaml", nil)

	ctrl := controller.NewMetaController(mockContratoError{}, zerolog.Nop())
	ctrl.OpenAPI(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("se esperaba 500, se obtuvo %d", w.Code)
	}

	var res map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res["error"] != "no se pudo armar el contrato" {
		t.Fatalf("error inesperado: %s", res["error"])
	}
}
