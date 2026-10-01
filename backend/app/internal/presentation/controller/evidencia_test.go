package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

type mockEvidenciaUC struct {
	listarFunc func(ctx context.Context, actividadID string) (*dto.ListarEvidenciasResponseDTO, error)
	subirFunc  func(ctx context.Context, in dto.SubirEvidenciaDTO) (*dto.SubirEvidenciaResponseDTO, error)
	abrirFunc  func(ctx context.Context, id string) (io.ReadCloser, string, error)
}

func (m *mockEvidenciaUC) Listar(ctx context.Context, actividadID string) (*dto.ListarEvidenciasResponseDTO, error) {
	if m.listarFunc != nil {
		return m.listarFunc(ctx, actividadID)
	}
	return &dto.ListarEvidenciasResponseDTO{Evidencias: []dto.EvidenciaDTO{}}, nil
}

func (m *mockEvidenciaUC) Subir(ctx context.Context, in dto.SubirEvidenciaDTO) (*dto.SubirEvidenciaResponseDTO, error) {
	if m.subirFunc != nil {
		return m.subirFunc(ctx, in)
	}
	return &dto.SubirEvidenciaResponseDTO{ID: in.ID, Idempotente: false}, nil
}

func (m *mockEvidenciaUC) Abrir(ctx context.Context, id string) (io.ReadCloser, string, error) {
	if m.abrirFunc != nil {
		return m.abrirFunc(ctx, id)
	}
	return io.NopCloser(bytes.NewReader([]byte("test"))), "image/jpeg", nil
}

func TestEvidenciaController_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("exitoso", func(t *testing.T) {
		mockUC := &mockEvidenciaUC{
			listarFunc: func(_ context.Context, actID string) (*dto.ListarEvidenciasResponseDTO, error) {
				return &dto.ListarEvidenciasResponseDTO{
					Evidencias: []dto.EvidenciaDTO{
						{ID: "e1", Nombre: "foto1.jpg", Mime: "image/jpeg", Bytes: 100},
					},
				}, nil
			},
		}
		ctrl := controller.NewEvidenciaController(mockUC, zerolog.Nop())

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/evidencias?actividad_id=act-1", nil)

		ctrl.List(c)

		if w.Code != http.StatusOK {
			t.Fatalf("esperado status 200, obtenido %d", w.Code)
		}
		var res dto.ListarEvidenciasResponseDTO
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatal(err)
		}
		if len(res.Evidencias) != 1 || res.Evidencias[0].ID != "e1" {
			t.Fatalf("respuesta inesperada: %+v", res)
		}
	})

	t.Run("error_bd", func(t *testing.T) {
		mockUC := &mockEvidenciaUC{
			listarFunc: func(_ context.Context, _ string) (*dto.ListarEvidenciasResponseDTO, error) {
				return nil, errors.New("db error")
			},
		}
		ctrl := controller.NewEvidenciaController(mockUC, zerolog.Nop())

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/evidencias", nil)

		ctrl.List(c)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("esperado status 500, obtenido %d", w.Code)
		}
		if !bytes.Contains(w.Body.Bytes(), []byte("no se pudieron leer las evidencias")) {
			t.Fatalf("cuerpo inesperado: %s", w.Body.String())
		}
	})
}

func TestEvidenciaController_Upload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sin_sesion", func(t *testing.T) {
		ctrl := controller.NewEvidenciaController(&mockEvidenciaUC{}, zerolog.Nop())
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/evidencias", nil)

		ctrl.Upload(c)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("esperado status 401, obtenido %d", w.Code)
		}
	})

	t.Run("falta_archivo", func(t *testing.T) {
		ctrl := controller.NewEvidenciaController(&mockEvidenciaUC{}, zerolog.Nop())
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/evidencias", bytes.NewBufferString("{}"))
		c.Request.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
		c.Set("usuario", dto.UsuarioSesionDTO{ID: 1, Rol: "capataz"})

		ctrl.Upload(c)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("esperado status 400, obtenido %d", w.Code)
		}
		if !bytes.Contains(w.Body.Bytes(), []byte("falta el archivo")) {
			t.Fatalf("cuerpo inesperado: %s", w.Body.String())
		}
	})

	t.Run("alta_exitosa", func(t *testing.T) {
		mockUC := &mockEvidenciaUC{
			subirFunc: func(_ context.Context, in dto.SubirEvidenciaDTO) (*dto.SubirEvidenciaResponseDTO, error) {
				return &dto.SubirEvidenciaResponseDTO{ID: in.ID, Idempotente: false}, nil
			},
		}
		ctrl := controller.NewEvidenciaController(mockUC, zerolog.Nop())

		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		_ = mw.WriteField("id", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
		_ = mw.WriteField("actividad_id", "11111111-1111-4111-8111-111111111111")
		part, _ := mw.CreateFormFile("archivo", "test.jpg")
		_, _ = part.Write([]byte{0xFF, 0xD8, 0xFF, 0xD9})
		_ = mw.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/evidencias", &buf)
		c.Request.Header.Set("Content-Type", mw.FormDataContentType())
		c.Set("usuario", dto.UsuarioSesionDTO{ID: 1, Rol: "capataz", CapatazID: "cap-norte"})

		ctrl.Upload(c)

		if w.Code != http.StatusCreated {
			t.Fatalf("esperado status 201, obtenido %d (%s)", w.Code, w.Body.String())
		}
		var res dto.SubirEvidenciaResponseDTO
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		if res.Idempotente {
			t.Fatalf("esperaba idempotente=false")
		}
	})

	t.Run("idempotente", func(t *testing.T) {
		mockUC := &mockEvidenciaUC{
			subirFunc: func(_ context.Context, in dto.SubirEvidenciaDTO) (*dto.SubirEvidenciaResponseDTO, error) {
				return &dto.SubirEvidenciaResponseDTO{ID: in.ID, Idempotente: true}, nil
			},
		}
		ctrl := controller.NewEvidenciaController(mockUC, zerolog.Nop())

		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		_ = mw.WriteField("id", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
		_ = mw.WriteField("actividad_id", "11111111-1111-4111-8111-111111111111")
		part, _ := mw.CreateFormFile("archivo", "test.jpg")
		_, _ = part.Write([]byte{0xFF, 0xD8, 0xFF, 0xD9})
		_ = mw.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/evidencias", &buf)
		c.Request.Header.Set("Content-Type", mw.FormDataContentType())
		c.Set("usuario", dto.UsuarioSesionDTO{ID: 1, Rol: "capataz"})

		ctrl.Upload(c)

		if w.Code != http.StatusOK {
			t.Fatalf("esperado status 200, obtenido %d", w.Code)
		}
		var res dto.SubirEvidenciaResponseDTO
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		if !res.Idempotente {
			t.Fatalf("esperaba idempotente=true")
		}
	})

	t.Run("conflicto_409", func(t *testing.T) {
		mockUC := &mockEvidenciaUC{
			subirFunc: func(_ context.Context, _ dto.SubirEvidenciaDTO) (*dto.SubirEvidenciaResponseDTO, error) {
				return nil, domainErrors.ErrLaborConflicto
			},
		}
		ctrl := controller.NewEvidenciaController(mockUC, zerolog.Nop())

		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		_ = mw.WriteField("id", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
		_ = mw.WriteField("actividad_id", "11111111-1111-4111-8111-111111111111")
		part, _ := mw.CreateFormFile("archivo", "test.jpg")
		_, _ = part.Write([]byte{0xFF, 0xD8, 0xFF, 0xD9})
		_ = mw.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/evidencias", &buf)
		c.Request.Header.Set("Content-Type", mw.FormDataContentType())
		c.Set("usuario", dto.UsuarioSesionDTO{ID: 1, Rol: "capataz"})

		ctrl.Upload(c)

		if w.Code != http.StatusConflict {
			t.Fatalf("esperado status 409, obtenido %d", w.Code)
		}
		if !bytes.Contains(w.Body.Bytes(), []byte("ese id ya existe con otro contenido")) {
			t.Fatalf("cuerpo inesperado: %s", w.Body.String())
		}
	})

	t.Run("prohibido_403", func(t *testing.T) {
		mockUC := &mockEvidenciaUC{
			subirFunc: func(_ context.Context, _ dto.SubirEvidenciaDTO) (*dto.SubirEvidenciaResponseDTO, error) {
				return nil, domainErrors.ErrOperacionProhibido
			},
		}
		ctrl := controller.NewEvidenciaController(mockUC, zerolog.Nop())

		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		_ = mw.WriteField("id", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
		_ = mw.WriteField("actividad_id", "11111111-1111-4111-8111-111111111111")
		part, _ := mw.CreateFormFile("archivo", "test.jpg")
		_, _ = part.Write([]byte{0xFF, 0xD8, 0xFF, 0xD9})
		_ = mw.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/evidencias", &buf)
		c.Request.Header.Set("Content-Type", mw.FormDataContentType())
		c.Set("usuario", dto.UsuarioSesionDTO{ID: 1, Rol: "capataz"})

		ctrl.Upload(c)

		if w.Code != http.StatusForbidden {
			t.Fatalf("esperado status 403, obtenido %d", w.Code)
		}
		if !bytes.Contains(w.Body.Bytes(), []byte("este rol no puede hacer esa acción")) {
			t.Fatalf("cuerpo inesperado: %s", w.Body.String())
		}
	})
}

func TestEvidenciaController_File(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("exitoso_cabeceras", func(t *testing.T) {
		fotoBytes := []byte{0xFF, 0xD8, 0xFF, 0xD9}
		mockUC := &mockEvidenciaUC{
			abrirFunc: func(_ context.Context, id string) (io.ReadCloser, string, error) {
				return io.NopCloser(bytes.NewReader(fotoBytes)), "image/jpeg", nil
			},
		}
		ctrl := controller.NewEvidenciaController(mockUC, zerolog.Nop())

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/evidencias/ev-1/archivo", nil)
		c.Params = gin.Params{{Key: "id", Value: "ev-1"}}

		ctrl.File(c)

		if w.Code != http.StatusOK {
			t.Fatalf("esperado status 200, obtenido %d", w.Code)
		}
		if got := w.Header().Get("Content-Type"); got != "image/jpeg" {
			t.Fatalf("Content-Type = %q", got)
		}
		if got := w.Header().Get("Cache-Control"); got != "private, max-age=86400" {
			t.Fatalf("Cache-Control = %q", got)
		}
		if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Fatalf("X-Content-Type-Options = %q", got)
		}
		if got := w.Header().Get("Content-Disposition"); got != "inline" {
			t.Fatalf("Content-Disposition = %q", got)
		}
		if !bytes.Equal(w.Body.Bytes(), fotoBytes) {
			t.Fatalf("cuerpo no coincide")
		}
	})

	t.Run("archivo_no_disponible_404", func(t *testing.T) {
		mockUC := &mockEvidenciaUC{
			abrirFunc: func(_ context.Context, _ string) (io.ReadCloser, string, error) {
				return nil, "", domainErrors.ErrArchivoNoDisponible
			},
		}
		ctrl := controller.NewEvidenciaController(mockUC, zerolog.Nop())

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/evidencias/ev-1/archivo", nil)
		c.Params = gin.Params{{Key: "id", Value: "ev-1"}}

		ctrl.File(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("esperado status 404, obtenido %d", w.Code)
		}
		if !bytes.Contains(w.Body.Bytes(), []byte("archivo no disponible")) {
			t.Fatalf("cuerpo inesperado: %s", w.Body.String())
		}
	})

	t.Run("labor_no_encontrada_404", func(t *testing.T) {
		mockUC := &mockEvidenciaUC{
			abrirFunc: func(_ context.Context, _ string) (io.ReadCloser, string, error) {
				return nil, "", domainErrors.ErrLaborNoEncontrada
			},
		}
		ctrl := controller.NewEvidenciaController(mockUC, zerolog.Nop())

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/evidencias/ev-1/archivo", nil)
		c.Params = gin.Params{{Key: "id", Value: "ev-1"}}

		ctrl.File(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("esperado status 404, obtenido %d", w.Code)
		}
		if !bytes.Contains(w.Body.Bytes(), []byte("labor no encontrada")) {
			t.Fatalf("cuerpo inesperado: %s", w.Body.String())
		}
	})
}
