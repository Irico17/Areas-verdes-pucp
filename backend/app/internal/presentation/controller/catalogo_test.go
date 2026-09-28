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

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

type mockCatalogoControllerUseCase struct {
	listarFn     func(ctx context.Context, filtro dto.FiltroCatalogoDTO) (*dto.CatalogoListResponseDTO, error)
	crearFn      func(ctx context.Context, in dto.CrearCatalogoDTO) (*dto.CatalogoItemDTO, error)
	desactivarFn func(ctx context.Context, id int64) (*dto.DesactivarCatalogoResponseDTO, error)
}

func (m *mockCatalogoControllerUseCase) Listar(ctx context.Context, filtro dto.FiltroCatalogoDTO) (*dto.CatalogoListResponseDTO, error) {
	if m.listarFn != nil {
		return m.listarFn(ctx, filtro)
	}
	return &dto.CatalogoListResponseDTO{Items: []dto.CatalogoItemDTO{}, Clases: []string{}}, nil
}

func (m *mockCatalogoControllerUseCase) Crear(ctx context.Context, in dto.CrearCatalogoDTO) (*dto.CatalogoItemDTO, error) {
	if m.crearFn != nil {
		return m.crearFn(ctx, in)
	}
	return &dto.CatalogoItemDTO{ID: 1, Clase: in.Clase, Codigo: in.Codigo, Nombre: in.Nombre, Activo: true, Orden: 0}, nil
}

func (m *mockCatalogoControllerUseCase) Desactivar(ctx context.Context, id int64) (*dto.DesactivarCatalogoResponseDTO, error) {
	if m.desactivarFn != nil {
		return m.desactivarFn(ctx, id)
	}
	return &dto.DesactivarCatalogoResponseDTO{Activo: false, ID: id}, nil
}

func TestCatalogoController_Listar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sin base de datos da 503", func(t *testing.T) {
		ctrl := controller.NewCatalogoController(nil)
		r := gin.New()
		r.GET("/catalogos", ctrl.Listar)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/catalogos", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("se esperaba 503, se obtuvo %d", w.Code)
		}
		if w.Body.String() != `{"error":"base de datos no disponible"}` {
			t.Fatalf("cuerpo 503 inesperado: %s", w.Body.String())
		}
	})

	t.Run("error de lectura da 500", func(t *testing.T) {
		mockUC := &mockCatalogoControllerUseCase{
			listarFn: func(ctx context.Context, filtro dto.FiltroCatalogoDTO) (*dto.CatalogoListResponseDTO, error) {
				return nil, errors.New("db error")
			},
		}
		ctrl := controller.NewCatalogoController(mockUC)
		r := gin.New()
		r.GET("/catalogos", ctrl.Listar)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/catalogos", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("se esperaba 500, se obtuvo %d", w.Code)
		}
		if w.Body.String() != `{"error":"no se pudo leer el catálogo"}` {
			t.Fatalf("cuerpo 500 inesperado: %s", w.Body.String())
		}
	})

	t.Run("exito pasa filtros y retorna 200", func(t *testing.T) {
		var capturedFiltro dto.FiltroCatalogoDTO
		mockUC := &mockCatalogoControllerUseCase{
			listarFn: func(ctx context.Context, filtro dto.FiltroCatalogoDTO) (*dto.CatalogoListResponseDTO, error) {
				capturedFiltro = filtro
				return &dto.CatalogoListResponseDTO{
					Items: []dto.CatalogoItemDTO{
						{ID: 1, Clase: "estado", Codigo: "pendiente", Nombre: "Pendiente", Activo: true, Orden: 1},
					},
					Clases: []string{"estado"},
				}, nil
			},
		}
		ctrl := controller.NewCatalogoController(mockUC)
		r := gin.New()
		r.GET("/catalogos", ctrl.Listar)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/catalogos?clase=estado&activos=1", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("se esperaba 200, se obtuvo %d", w.Code)
		}
		if capturedFiltro.Clase != "estado" || !capturedFiltro.SoloActivos {
			t.Fatalf("filtros capturados inesperados: %+v", capturedFiltro)
		}

		var resp dto.CatalogoListResponseDTO
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if len(resp.Items) != 1 || resp.Items[0].Codigo != "pendiente" {
			t.Fatalf("items retornados inesperados: %+v", resp.Items)
		}
	})
}

func TestCatalogoController_Crear(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sin base de datos da 503", func(t *testing.T) {
		ctrl := controller.NewCatalogoController(nil)
		r := gin.New()
		r.POST("/catalogos", ctrl.Crear)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/catalogos", bytes.NewBufferString(`{"clase":"estado"}`))
		r.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("se esperaba 503, se obtuvo %d", w.Code)
		}
	})

	t.Run("JSON invalido da 400", func(t *testing.T) {
		ctrl := controller.NewCatalogoController(&mockCatalogoControllerUseCase{})
		r := gin.New()
		r.POST("/catalogos", ctrl.Crear)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/catalogos", bytes.NewBufferString(`{invalido`))
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("se esperaba 400, se obtuvo %d", w.Code)
		}
		if w.Body.String() != `{"error":"JSON inválido"}` {
			t.Fatalf("cuerpo 400 inesperado: %s", w.Body.String())
		}
	})

	t.Run("validaciones de entrada retornan 400 con mensaje exacto", func(t *testing.T) {
		validationErrors := []error{
			apperrors.ErrClaseNoReconocida,
			apperrors.ErrCodigoInvalido,
			apperrors.ErrNombreObligatorio,
		}

		for _, vErr := range validationErrors {
			mockUC := &mockCatalogoControllerUseCase{
				crearFn: func(ctx context.Context, in dto.CrearCatalogoDTO) (*dto.CatalogoItemDTO, error) {
					return nil, vErr
				},
			}
			ctrl := controller.NewCatalogoController(mockUC)
			r := gin.New()
			r.POST("/catalogos", ctrl.Crear)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/catalogos", bytes.NewBufferString(`{"clase":"x","codigo":"y","nombre":"z"}`))
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("para %v se esperaba 400, se obtuvo %d", vErr, w.Code)
			}
			expectedBody, _ := json.Marshal(gin.H{"error": vErr.Error()})
			if w.Body.String() != string(expectedBody) {
				t.Fatalf("esperado %s, obtenido %s", expectedBody, w.Body.String())
			}
		}
	})

	t.Run("exito retorna 201 y el item creado", func(t *testing.T) {
		mockUC := &mockCatalogoControllerUseCase{
			crearFn: func(ctx context.Context, in dto.CrearCatalogoDTO) (*dto.CatalogoItemDTO, error) {
				return &dto.CatalogoItemDTO{
					ID:     7,
					Clase:  in.Clase,
					Codigo: in.Codigo,
					Nombre: in.Nombre,
					Activo: true,
					Orden:  0,
				}, nil
			},
		}
		ctrl := controller.NewCatalogoController(mockUC)
		r := gin.New()
		r.POST("/catalogos", ctrl.Crear)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/catalogos", bytes.NewBufferString(`{"clase":"estado","codigo":"nuevo","nombre":"Nuevo"}`))
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("se esperaba 201, se obtuvo %d", w.Code)
		}

		var item dto.CatalogoItemDTO
		if err := json.Unmarshal(w.Body.Bytes(), &item); err != nil {
			t.Fatal(err)
		}
		if item.ID != 7 || item.Codigo != "nuevo" {
			t.Fatalf("item creado inesperado: %+v", item)
		}
	})
}

func TestCatalogoController_Desactivar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sin base de datos da 503", func(t *testing.T) {
		ctrl := controller.NewCatalogoController(nil)
		r := gin.New()
		r.POST("/catalogos/:id/desactivar", ctrl.Desactivar)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/catalogos/1/desactivar", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("se esperaba 503, se obtuvo %d", w.Code)
		}
	})

	t.Run("id invalido da 400", func(t *testing.T) {
		ctrl := controller.NewCatalogoController(&mockCatalogoControllerUseCase{})
		r := gin.New()
		r.POST("/catalogos/:id/desactivar", ctrl.Desactivar)

		invalidIDs := []string{"0", "-5", "abc", "12a"}
		for _, id := range invalidIDs {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/catalogos/"+id+"/desactivar", nil)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("para id %q se esperaba 400, se obtuvo %d", id, w.Code)
			}
			if w.Body.String() != `{"error":"id inválido"}` {
				t.Fatalf("cuerpo 400 inesperado: %s", w.Body.String())
			}
		}
	})

	t.Run("item no existe da 404", func(t *testing.T) {
		mockUC := &mockCatalogoControllerUseCase{
			desactivarFn: func(ctx context.Context, id int64) (*dto.DesactivarCatalogoResponseDTO, error) {
				return nil, apperrors.ErrItemNoExiste
			},
		}
		ctrl := controller.NewCatalogoController(mockUC)
		r := gin.New()
		r.POST("/catalogos/:id/desactivar", ctrl.Desactivar)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/catalogos/999/desactivar", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("se esperaba 404, se obtuvo %d", w.Code)
		}
		if w.Body.String() != `{"error":"no existe ese ítem"}` {
			t.Fatalf("cuerpo 404 inesperado: %s", w.Body.String())
		}
	})

	t.Run("exito retorna 200 con activo=false y id", func(t *testing.T) {
		mockUC := &mockCatalogoControllerUseCase{
			desactivarFn: func(ctx context.Context, id int64) (*dto.DesactivarCatalogoResponseDTO, error) {
				return &dto.DesactivarCatalogoResponseDTO{Activo: false, ID: id}, nil
			},
		}
		ctrl := controller.NewCatalogoController(mockUC)
		r := gin.New()
		r.POST("/catalogos/:id/desactivar", ctrl.Desactivar)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/catalogos/5/desactivar", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("se esperaba 200, se obtuvo %d", w.Code)
		}
		var res dto.DesactivarCatalogoResponseDTO
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatal(err)
		}
		if res.Activo != false || res.ID != 5 {
			t.Fatalf("respuesta inesperada: %+v", res)
		}
	})
}
