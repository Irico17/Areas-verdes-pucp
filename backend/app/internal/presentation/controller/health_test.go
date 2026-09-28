package controller_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

type mockSaludUseCase struct {
	resultado *dto.EstadoSaludDTO
	err       error
}

func (m *mockSaludUseCase) VerificarSalud(_ context.Context) (*dto.EstadoSaludDTO, error) {
	return m.resultado, m.err
}

func TestHealthController_Health_Healthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/areas-verdes/v1/health", nil)

	mockUC := &mockSaludUseCase{
		resultado: &dto.EstadoSaludDTO{
			Database: "up",
			PostGIS:  "3.5",
		},
	}
	ctrl := controller.NewHealthController(mockUC)
	ctrl.Health(c)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, se obtuvo %d", w.Code)
	}

	var res controller.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Status != "healthy" || res.Database != "up" || res.PostGIS != "3.5" || res.Timestamp == "" {
		t.Fatalf("respuesta inesperada: %+v", res)
	}
}

func TestHealthController_Health_Unhealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/areas-verdes/v1/health", nil)

	mockUC := &mockSaludUseCase{
		resultado: &dto.EstadoSaludDTO{
			Database: "down",
		},
		err: errors.New("db unreachable"),
	}
	ctrl := controller.NewHealthController(mockUC)
	ctrl.Health(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("se esperaba 503, se obtuvo %d", w.Code)
	}

	var res controller.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Status != "unhealthy" || res.Database != "down" {
		t.Fatalf("respuesta inesperada: %+v", res)
	}
}

func TestHealthController_LegacyHealth_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	mockUC := &mockSaludUseCase{
		resultado: &dto.EstadoSaludDTO{
			Database: "up",
			PostGIS:  "3.5 USE_GEOS=1",
		},
	}
	ctrl := controller.NewHealthController(mockUC)
	ctrl.LegacyHealth(c)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, se obtuvo %d", w.Code)
	}

	var res map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res["status"] != "ok" || res["service"] != "campus-verde-api" || res["database"] != "up" || res["postgis"] != "3.5 USE_GEOS=1" {
		t.Fatalf("respuesta inesperada: %+v", res)
	}
}

func TestHealthController_LegacyHealth_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	mockUC := &mockSaludUseCase{
		resultado: &dto.EstadoSaludDTO{
			Database: "down",
		},
		err: errors.New("timeout"),
	}
	ctrl := controller.NewHealthController(mockUC)
	ctrl.LegacyHealth(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("se esperaba 503, se obtuvo %d", w.Code)
	}

	var res map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res["status"] != "error" || res["service"] != "campus-verde-api" || res["database"] != "down" {
		t.Fatalf("respuesta inesperada: %+v", res)
	}
}
