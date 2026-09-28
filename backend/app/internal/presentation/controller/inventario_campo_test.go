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
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

type mockInventarioCampoUC struct {
	listarTachosRes    dto.ListarTachosResponseDTO
	guardarTachoRes    *dto.TachoDTO
	actualizarTachoRes *dto.TachoDTO
	csvTachosRes       string

	listarBebederosRes    dto.ListarBebederosResponseDTO
	guardarBebederoRes    *dto.BebederoDTO
	actualizarBebederoRes *dto.BebederoDTO

	listarPuntosRes    dto.ListarPuntosResponseDTO
	guardarPuntoRes    *dto.PuntoDTO
	actualizarPuntoRes *dto.PuntoDTO
	formatoPuntosRes   dto.FormatoPuntosResponseDTO

	listarReservasRes    dto.ListarReservasResponseDTO
	guardarReservaRes    *dto.ReservaDTO
	actualizarReservaRes *dto.ReservaDTO

	listarCapaRes     dto.ListarFichasCapaResponseDTO
	guardarCapaRes    *dto.FichaCapaDTO
	actualizarCapaRes *dto.FichaCapaDTO
	csvCapaRes        string

	err error
}

func (m *mockInventarioCampoUC) ListarTachos(_ context.Context) (dto.ListarTachosResponseDTO, error) {
	return m.listarTachosRes, m.err
}

func (m *mockInventarioCampoUC) GuardarTacho(_ context.Context, _ dto.TachoDTO) (*dto.TachoDTO, error) {
	return m.guardarTachoRes, m.err
}

func (m *mockInventarioCampoUC) ActualizarTacho(_ context.Context, _ int64, _ dto.TachoDTO, _ []byte) (*dto.TachoDTO, error) {
	return m.actualizarTachoRes, m.err
}

func (m *mockInventarioCampoUC) CSVTachos(_ context.Context) (string, error) {
	return m.csvTachosRes, m.err
}

func (m *mockInventarioCampoUC) BajaTacho(_ context.Context, id int64) error {
	if id < 1 {
		return domainErrors.ErrRegistroNoEncontrado
	}
	return m.err
}

func (m *mockInventarioCampoUC) ListarBebederos(_ context.Context) (dto.ListarBebederosResponseDTO, error) {
	return m.listarBebederosRes, m.err
}

func (m *mockInventarioCampoUC) GuardarBebedero(_ context.Context, _ dto.BebederoDTO) (*dto.BebederoDTO, error) {
	return m.guardarBebederoRes, m.err
}

func (m *mockInventarioCampoUC) ActualizarBebedero(_ context.Context, _ int64, _ dto.BebederoDTO, _ []byte) (*dto.BebederoDTO, error) {
	return m.actualizarBebederoRes, m.err
}

func (m *mockInventarioCampoUC) BajaBebedero(_ context.Context, id int64) error {
	if id < 1 {
		return domainErrors.ErrRegistroNoEncontrado
	}
	return m.err
}

func (m *mockInventarioCampoUC) ListarPuntos(_ context.Context, _ string) (dto.ListarPuntosResponseDTO, error) {
	return m.listarPuntosRes, m.err
}

func (m *mockInventarioCampoUC) GuardarPunto(_ context.Context, _ dto.PuntoDTO, _ []byte) (*dto.PuntoDTO, error) {
	return m.guardarPuntoRes, m.err
}

func (m *mockInventarioCampoUC) ActualizarPunto(_ context.Context, _ int64, _ dto.PuntoDTO, _ []byte) (*dto.PuntoDTO, error) {
	return m.actualizarPuntoRes, m.err
}

func (m *mockInventarioCampoUC) BajaPunto(_ context.Context, id int64) error {
	if id < 1 {
		return domainErrors.ErrRegistroNoEncontrado
	}
	return m.err
}

func (m *mockInventarioCampoUC) FormatoPuntos(_ context.Context, _ []byte) (dto.FormatoPuntosResponseDTO, error) {
	return m.formatoPuntosRes, m.err
}

func (m *mockInventarioCampoUC) ListarReservas(_ context.Context, _, _ string) (dto.ListarReservasResponseDTO, error) {
	return m.listarReservasRes, m.err
}

func (m *mockInventarioCampoUC) GuardarReserva(_ context.Context, _ dto.ReservaDTO) (*dto.ReservaDTO, error) {
	return m.guardarReservaRes, m.err
}

func (m *mockInventarioCampoUC) ActualizarReserva(_ context.Context, _ int64, _ dto.ReservaDTO, _ []byte) (*dto.ReservaDTO, error) {
	return m.actualizarReservaRes, m.err
}

func (m *mockInventarioCampoUC) BajaReserva(_ context.Context, id int64) error {
	if id < 1 {
		return domainErrors.ErrRegistroNoEncontrado
	}
	return m.err
}

func (m *mockInventarioCampoUC) ListarCapa(_ context.Context, _ string) (dto.ListarFichasCapaResponseDTO, error) {
	return m.listarCapaRes, m.err
}

func (m *mockInventarioCampoUC) GuardarCapa(_ context.Context, _ string, _ dto.FichaCapaDTO) (*dto.FichaCapaDTO, error) {
	return m.guardarCapaRes, m.err
}

func (m *mockInventarioCampoUC) ActualizarCapa(_ context.Context, _ string, _ int64, _ dto.FichaCapaDTO, _ []byte) (*dto.FichaCapaDTO, error) {
	return m.actualizarCapaRes, m.err
}

func (m *mockInventarioCampoUC) CSVCapa(_ context.Context, _ string) (string, error) {
	return m.csvCapaRes, m.err
}

func (m *mockInventarioCampoUC) BajaCapa(_ context.Context, _ string, id int64) error {
	if id < 1 {
		return domainErrors.ErrRegistroNoEncontrado
	}
	return m.err
}

func TestInventarioCampoController_Tachos(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zerolog.Nop()

	// 1. Listar 200
	uc := &mockInventarioCampoUC{
		listarTachosRes: dto.ListarTachosResponseDTO{Tachos: []dto.TachoDTO{{ID: 1, Codigo: "PT-01"}}},
	}
	ctrl := controller.NewInventarioCampoController(uc, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/inventario/tachos", nil)
	ctrl.ListarTachos(c)
	if w.Code != http.StatusOK {
		t.Fatalf("ListarTachos esperado 200, obtenido %d", w.Code)
	}

	// 2. Listar 500
	ucErr := &mockInventarioCampoUC{err: errors.New("db error")}
	ctrlErr := controller.NewInventarioCampoController(ucErr, logger)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/inventario/tachos", nil)
	ctrlErr.ListarTachos(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("ListarTachos error esperado 500, obtenido %d", w.Code)
	}

	// 3. Guardar 201
	ucSave := &mockInventarioCampoUC{
		guardarTachoRes: &dto.TachoDTO{ID: 1, Codigo: "PT-01"},
	}
	ctrlSave := controller.NewInventarioCampoController(ucSave, logger)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/inventario/tachos", bytes.NewBufferString(`{"codigo":"PT-01"}`))
	ctrlSave.GuardarTacho(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("GuardarTacho esperado 201, obtenido %d", w.Code)
	}

	// 4. Guardar 400 (JSON invalido)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/inventario/tachos", bytes.NewBufferString(`invalid json`))
	ctrlSave.GuardarTacho(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("GuardarTacho esperado 400 por JSON invalido, obtenido %d", w.Code)
	}

	// 5. CSV Tachos 200
	ucCSV := &mockInventarioCampoUC{csvTachosRes: "codigo,lugar\nPT-01,Central\n"}
	ctrlCSV := controller.NewInventarioCampoController(ucCSV, logger)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/inventario/tachos.csv", nil)
	ctrlCSV.CSVTachos(c)
	if w.Code != http.StatusOK {
		t.Fatalf("CSVTachos esperado 200, obtenido %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/csv; charset=utf-8" {
		t.Fatalf("CSVTachos Content-Type esperado 'text/csv; charset=utf-8', obtenido %q", ct)
	}

	// 6. Baja 200 y 404
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/inventario/tachos/1", nil)
	ctrlSave.BajaTacho(c)
	if w.Code != http.StatusOK {
		t.Fatalf("BajaTacho esperado 200, obtenido %d", w.Code)
	}

	ucBajaErr := &mockInventarioCampoUC{err: domainErrors.ErrRegistroNoEncontrado}
	ctrlBajaErr := controller.NewInventarioCampoController(ucBajaErr, logger)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/inventario/tachos/999", nil)
	ctrlBajaErr.BajaTacho(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("BajaTacho esperado 404, obtenido %d", w.Code)
	}
}

func TestInventarioCampoController_PuntosContacto(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zerolog.Nop()

	uc := &mockInventarioCampoUC{err: domainErrors.ErrPuntoContacto}
	ctrl := controller.NewInventarioCampoController(uc, logger)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/inventario/puntos", bytes.NewBufferString(`{"titulo":"Comedor","phone":"999"}`))
	ctrl.GuardarPunto(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("GuardarPunto con contacto esperado 400, obtenido %d", w.Code)
	}
	var res map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res["error"] != "no se guardan teléfono, placeId ni website" {
		t.Fatalf("mensaje inesperado: %v", res["error"])
	}
	cols, ok := res["columnas_omitidas"].([]any)
	if !ok || len(cols) != 3 {
		t.Fatalf("columnas_omitidas inesperadas: %v", res["columnas_omitidas"])
	}
}

func TestInventarioCampoController_Reservas(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zerolog.Nop()

	uc := &mockInventarioCampoUC{
		guardarReservaRes: &dto.ReservaDTO{ID: 1, Origen: "ficticio"},
	}
	ctrl := controller.NewInventarioCampoController(uc, logger)

	// Origen no ficticio -> 400
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/inventario/reservas", bytes.NewBufferString(`{"origen":"sheet"}`))
	ctrl.GuardarReserva(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("GuardarReserva con origen != ficticio esperado 400, obtenido %d", w.Code)
	}
}

func TestInventarioCampoController_Capas(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zerolog.Nop()

	uc := &mockInventarioCampoUC{
		csvCapaRes: "feature_id,nombre\nFAU-01,Ardilla\n",
	}
	ctrl := controller.NewInventarioCampoController(uc, logger)

	// CSV Capa 200
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "capa", Value: "fauna"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/inventario/export/fauna", nil)
	ctrl.CSVCapa(c)
	if w.Code != http.StatusOK {
		t.Fatalf("CSVCapa esperado 200, obtenido %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/csv; charset=utf-8" {
		t.Fatalf("CSVCapa Content-Type esperado 'text/csv; charset=utf-8', obtenido %q", ct)
	}

	// CSV Capa desconocida -> 404
	ucDesconocida := &mockInventarioCampoUC{err: domainErrors.ErrCapaDesconocida}
	ctrlDesconocida := controller.NewInventarioCampoController(ucDesconocida, logger)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "capa", Value: "desconocida"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/inventario/export/desconocida", nil)
	ctrlDesconocida.CSVCapa(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("CSVCapa desconocida esperado 404, obtenido %d", w.Code)
	}
}
