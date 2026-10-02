package controller_test

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/etl"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

type mockImportacionUC struct {
	entidadesFn     func(ctx context.Context) []string
	previsualizarFn func(ctx context.Context, entidad, nombre string, body []byte, usuarioID int64) (*dto.VistaPreviaResponseDTO, error)
	confirmarFn     func(ctx context.Context, loteID, usuarioID int64) (*dto.ConfirmarImportacionResponseDTO, error)
}

func (m *mockImportacionUC) Entidades(ctx context.Context) []string {
	if m.entidadesFn != nil {
		return m.entidadesFn(ctx)
	}
	return []string{"areas_verdes", "lugares"}
}

func (m *mockImportacionUC) Previsualizar(ctx context.Context, entidad, nombre string, body []byte, usuarioID int64) (*dto.VistaPreviaResponseDTO, error) {
	if m.previsualizarFn != nil {
		return m.previsualizarFn(ctx, entidad, nombre, body, usuarioID)
	}
	return &dto.VistaPreviaResponseDTO{
		ID:      1,
		LoteID:  1,
		Entidad: entidad,
		Validas: 1,
	}, nil
}

func (m *mockImportacionUC) Confirmar(ctx context.Context, loteID, usuarioID int64) (*dto.ConfirmarImportacionResponseDTO, error) {
	if m.confirmarFn != nil {
		return m.confirmarFn(ctx, loteID, usuarioID)
	}
	return &dto.ConfirmarImportacionResponseDTO{
		LoteID:  loteID,
		Validas: 1,
		Escrito: true,
	}, nil
}

func setupImportacionTestRouter(uc *mockImportacionUC, usuario *dto.UsuarioSesionDTO) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if usuario != nil {
		r.Use(func(c *gin.Context) {
			c.Set("usuario", *usuario)
			c.Next()
		})
	}
	ctrl := controller.NewImportacionController(uc, zerolog.Nop())
	r.GET("/api/v1/importaciones/entidades", ctrl.Entidades)
	r.POST("/api/v1/importaciones", ctrl.Previsualizar)
	r.POST("/api/v1/importaciones/:id/confirmar", ctrl.Confirmar)
	return r
}

func TestImportacionController_Entidades(t *testing.T) {
	r := setupImportacionTestRouter(&mockImportacionUC{}, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/importaciones/entidades", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtenido %d", w.Code)
	}
}

func TestImportacionController_Previsualizar_SinAuth(t *testing.T) {
	r := setupImportacionTestRouter(&mockImportacionUC{}, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/importaciones?entidad=lugares", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtenido %d", w.Code)
	}
}

func TestImportacionController_Previsualizar_FaltaArchivo(t *testing.T) {
	u := &dto.UsuarioSesionDTO{ID: 1, Usuario: "coordinacion", Rol: "coordinacion"}
	r := setupImportacionTestRouter(&mockImportacionUC{}, u)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/importaciones?entidad=lugares", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, obtenido %d (%s)", w.Code, w.Body.String())
	}
}

func TestImportacionController_Previsualizar_Exito(t *testing.T) {
	u := &dto.UsuarioSesionDTO{ID: 1, Usuario: "coordinacion", Rol: "coordinacion"}
	r := setupImportacionTestRouter(&mockImportacionUC{}, u)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("archivo", "lugares.csv")
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte("lugar,lat,lon\nParque,-12.0,-77.0\n"))
	mw.Close()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/importaciones?entidad=lugares", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtenido %d (%s)", w.Code, w.Body.String())
	}
}

func TestImportacionController_Confirmar_Errores(t *testing.T) {
	u := &dto.UsuarioSesionDTO{ID: 1, Usuario: "coordinacion", Rol: "coordinacion"}

	t.Run("lote inválido", func(t *testing.T) {
		r := setupImportacionTestRouter(&mockImportacionUC{}, u)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/importaciones/0/confirmar", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperado 400, obtenido %d", w.Code)
		}
	})

	t.Run("lote no confirmable", func(t *testing.T) {
		uc := &mockImportacionUC{
			confirmarFn: func(ctx context.Context, loteID, usuarioID int64) (*dto.ConfirmarImportacionResponseDTO, error) {
				return nil, etl.ErrLote
			},
		}
		r := setupImportacionTestRouter(uc, u)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/importaciones/5/confirmar", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusConflict {
			t.Fatalf("esperado 409, obtenido %d", w.Code)
		}
	})

	t.Run("sin válidas", func(t *testing.T) {
		uc := &mockImportacionUC{
			confirmarFn: func(ctx context.Context, loteID, usuarioID int64) (*dto.ConfirmarImportacionResponseDTO, error) {
				return nil, etl.ErrSinValidas
			},
		}
		r := setupImportacionTestRouter(uc, u)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/importaciones/5/confirmar", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("esperado 422, obtenido %d", w.Code)
		}
	})

	t.Run("éxito 201", func(t *testing.T) {
		r := setupImportacionTestRouter(&mockImportacionUC{}, u)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/importaciones/5/confirmar", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("esperado 201, obtenido %d", w.Code)
		}
	})

	t.Run("error de base de datos 500", func(t *testing.T) {
		uc := &mockImportacionUC{
			confirmarFn: func(ctx context.Context, loteID, usuarioID int64) (*dto.ConfirmarImportacionResponseDTO, error) {
				return nil, errors.New("pq: deadlock detected")
			},
		}
		r := setupImportacionTestRouter(uc, u)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/importaciones/5/confirmar", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("esperado 500, obtenido %d", w.Code)
		}
	})
}
