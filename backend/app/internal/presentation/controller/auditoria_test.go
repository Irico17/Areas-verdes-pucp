package controller_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

type mockAuditoriaControllerUseCase struct {
	importarFn  func(ctx context.Context, usuarioID int64, req dto.ImportarLoteDTO) (*dto.ImportarLoteResponseDTO, error)
	revertirFn  func(ctx context.Context, loteID, usuarioID int64, confirmar bool) (*dto.ReporteReversionDTO, error)
	editarFn    func(ctx context.Context, usuarioID int64, req dto.EditarAuditoriaDTO) (*dto.EditarAuditoriaResponseDTO, error)
	timelineFn  func(ctx context.Context, f dto.FiltroAuditoriaDTO) ([]dto.EventoAuditoriaDTO, error)
	historialFn func(ctx context.Context, f dto.FiltroAuditoriaDTO) ([]dto.EventoAuditoriaDTO, error)
}

func (m *mockAuditoriaControllerUseCase) Importar(ctx context.Context, usuarioID int64, req dto.ImportarLoteDTO) (*dto.ImportarLoteResponseDTO, error) {
	if m.importarFn != nil {
		return m.importarFn(ctx, usuarioID, req)
	}
	return &dto.ImportarLoteResponseDTO{LoteID: 10, Filas: len(req.Filas)}, nil
}

func (m *mockAuditoriaControllerUseCase) Revertir(ctx context.Context, loteID, usuarioID int64, confirmar bool) (*dto.ReporteReversionDTO, error) {
	if m.revertirFn != nil {
		return m.revertirFn(ctx, loteID, usuarioID, confirmar)
	}
	return &dto.ReporteReversionDTO{LoteID: loteID, Revertidas: []string{"1"}, Excluidas: []dto.ExcluidaDTO{}}, nil
}

func (m *mockAuditoriaControllerUseCase) Editar(ctx context.Context, usuarioID int64, req dto.EditarAuditoriaDTO) (*dto.EditarAuditoriaResponseDTO, error) {
	if m.editarFn != nil {
		return m.editarFn(ctx, usuarioID, req)
	}
	return &dto.EditarAuditoriaResponseDTO{Editada: true, EntidadID: req.EntidadID}, nil
}

func (m *mockAuditoriaControllerUseCase) Timeline(ctx context.Context, f dto.FiltroAuditoriaDTO) ([]dto.EventoAuditoriaDTO, error) {
	if m.timelineFn != nil {
		return m.timelineFn(ctx, f)
	}
	return []dto.EventoAuditoriaDTO{}, nil
}

func (m *mockAuditoriaControllerUseCase) Historial(ctx context.Context, f dto.FiltroAuditoriaDTO) ([]dto.EventoAuditoriaDTO, error) {
	if m.historialFn != nil {
		return m.historialFn(ctx, f)
	}
	return []dto.EventoAuditoriaDTO{}, nil
}

func TestAuditoriaController_Importar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := &mockAuditoriaControllerUseCase{}
	ctrl := controller.NewAuditoriaController(mockUC, zerolog.Nop())

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("usuario", dto.UsuarioSesionDTO{ID: 1, Usuario: "admin"})
		c.Next()
	})
	r.POST("/api/v1/lotes", ctrl.Importar)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lotes", strings.NewReader(`{"entidad":"catalogos","filas":[{"entidad_id":"1","accion":"alta","antes":null,"despues":{"nombre":"test"}}]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("esperado 201, obtenido %d: %s", w.Code, w.Body.String())
	}
}

func TestAuditoriaController_RevertirConflicto(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := &mockAuditoriaControllerUseCase{
		revertirFn: func(ctx context.Context, loteID, usuarioID int64, confirmar bool) (*dto.ReporteReversionDTO, error) {
			return &dto.ReporteReversionDTO{
				LoteID:     loteID,
				Revertidas: []string{},
				Excluidas: []dto.ExcluidaDTO{
					{EntidadID: "1", Motivo: "hay una edición posterior; no se pisa"},
				},
			}, apperrors.ErrConfirmacion
		},
	}
	ctrl := controller.NewAuditoriaController(mockUC, zerolog.Nop())

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("usuario", dto.UsuarioSesionDTO{ID: 1, Usuario: "admin"})
		c.Next()
	})
	r.POST("/api/v1/lotes/:id/revertir", ctrl.Revertir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lotes/5/revertir", strings.NewReader(`{"confirmar":false}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("esperado 409, obtenido %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["error"] != "hay filas editadas después del lote" {
		t.Fatalf("error inesperado: %v", resp["error"])
	}
}

func TestAuditoriaController_RevertirNoEncontrado(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := &mockAuditoriaControllerUseCase{
		revertirFn: func(ctx context.Context, loteID, usuarioID int64, confirmar bool) (*dto.ReporteReversionDTO, error) {
			return nil, apperrors.ErrLoteNoEncontrado
		},
	}
	ctrl := controller.NewAuditoriaController(mockUC, zerolog.Nop())

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("usuario", dto.UsuarioSesionDTO{ID: 1, Usuario: "admin"})
		c.Next()
	})
	r.POST("/api/v1/lotes/:id/revertir", ctrl.Revertir)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/lotes/999/revertir", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, obtenido %d: %s", w.Code, w.Body.String())
	}
}

func TestAuditoriaController_Editar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := &mockAuditoriaControllerUseCase{}
	ctrl := controller.NewAuditoriaController(mockUC, zerolog.Nop())

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("usuario", dto.UsuarioSesionDTO{ID: 1, Usuario: "admin"})
		c.Next()
	})
	r.POST("/api/v1/auditoria/ediciones", ctrl.Editar)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auditoria/ediciones", strings.NewReader(`{"entidad":"catalogos","entidad_id":"1","despues":{"nombre":"Editado"}}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtenido %d: %s", w.Code, w.Body.String())
	}
}

func TestAuditoriaController_TimelineError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := &mockAuditoriaControllerUseCase{
		timelineFn: func(ctx context.Context, f dto.FiltroAuditoriaDTO) ([]dto.EventoAuditoriaDTO, error) {
			return nil, apperrors.InputError{Reason: "entidad y entidad_id son obligatorios"}
		},
	}
	ctrl := controller.NewAuditoriaController(mockUC, zerolog.Nop())

	r := gin.New()
	r.GET("/api/v1/auditoria/timeline", ctrl.Timeline)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria/timeline", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, obtenido %d: %s", w.Code, w.Body.String())
	}
}

func TestAuditoriaController_ErrorInesperado500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := &mockAuditoriaControllerUseCase{
		historialFn: func(ctx context.Context, f dto.FiltroAuditoriaDTO) ([]dto.EventoAuditoriaDTO, error) {
			return nil, errors.New("falla interna")
		},
	}
	ctrl := controller.NewAuditoriaController(mockUC, zerolog.Nop())

	r := gin.New()
	r.GET("/api/v1/auditoria/cambios", ctrl.Historial)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auditoria/cambios", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("esperado 500, obtenido %d: %s", w.Code, w.Body.String())
	}
}
