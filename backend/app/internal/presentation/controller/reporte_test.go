package controller_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

type mockReporteUseCase struct {
	obtenerFn  func(ctx context.Context, f dto.FiltroReporteDTO) (*dto.ReporteResponseDTO, error)
	exportarFn func(ctx context.Context, f dto.FiltroReporteDTO, formato string, w io.Writer) error
}

func (m *mockReporteUseCase) ObtenerReporte(ctx context.Context, f dto.FiltroReporteDTO) (*dto.ReporteResponseDTO, error) {
	if m.obtenerFn != nil {
		return m.obtenerFn(ctx, f)
	}
	return &dto.ReporteResponseDTO{
		Aviso:      "aviso test",
		PorEstado:  []dto.ConteoReporteDTO{},
		Filas:      []dto.FilaReporteDTO{},
		Pendientes: []dto.HuecoIndicadorDTO{},
	}, nil
}

func (m *mockReporteUseCase) Exportar(ctx context.Context, f dto.FiltroReporteDTO, formato string, w io.Writer) error {
	if m.exportarFn != nil {
		return m.exportarFn(ctx, f, formato, w)
	}
	if formato == "csv" {
		_, err := io.WriteString(w, "csv content")
		return err
	}
	if formato == "xls" {
		_, err := io.WriteString(w, "xls content")
		return err
	}
	return errors.New("formato invalido")
}

func TestReporteController_JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &mockReporteUseCase{}
	ctrl := controller.NewReporteController(uc, zerolog.Nop())

	r := gin.New()
	r.GET("/reportes/labores", ctrl.ReporteLabores)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reportes/labores", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obtenido %d", w.Code)
	}
	if !stringsContains(w.Body.String(), `"aviso":"aviso test"`) {
		t.Fatalf("cuerpo inesperado: %s", w.Body.String())
	}
}

func TestReporteController_CSV(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &mockReporteUseCase{}
	ctrl := controller.NewReporteController(uc, zerolog.Nop())

	r := gin.New()
	r.GET("/reportes/labores", ctrl.ReporteLabores)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reportes/labores?formato=csv", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obtenido %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/csv; charset=utf-8" {
		t.Errorf("Content-Type = %q, esperado text/csv; charset=utf-8", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); cd != `attachment; filename="labores.csv"` {
		t.Errorf("Content-Disposition = %q", cd)
	}
	if w.Body.String() != "csv content" {
		t.Errorf("cuerpo inesperado: %s", w.Body.String())
	}
}

func TestReporteController_XLS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &mockReporteUseCase{}
	ctrl := controller.NewReporteController(uc, zerolog.Nop())

	r := gin.New()
	r.GET("/reportes/labores", ctrl.ReporteLabores)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reportes/labores?formato=xls", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obtenido %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/vnd.ms-excel" {
		t.Errorf("Content-Type = %q, esperado application/vnd.ms-excel", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); cd != `attachment; filename="labores.xls"` {
		t.Errorf("Content-Disposition = %q", cd)
	}
	if w.Body.String() != "xls content" {
		t.Errorf("cuerpo inesperado: %s", w.Body.String())
	}
}

func TestReporteController_ErrorFiltro(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &mockReporteUseCase{
		obtenerFn: func(ctx context.Context, f dto.FiltroReporteDTO) (*dto.ReporteResponseDTO, error) {
			return nil, domainErrors.InputError{Reason: "desde usa AAAA-MM-DD"}
		},
	}
	ctrl := controller.NewReporteController(uc, zerolog.Nop())

	r := gin.New()
	r.GET("/reportes/labores", ctrl.ReporteLabores)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reportes/labores?desde=invalido", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperado status 400, obtenido %d", w.Code)
	}
	if !stringsContains(w.Body.String(), `desde usa AAAA-MM-DD`) {
		t.Fatalf("cuerpo inesperado: %s", w.Body.String())
	}
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && stringContainsSub(s, substr)))
}

func stringContainsSub(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
