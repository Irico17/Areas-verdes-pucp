package controller_test

import (
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

type mockReservasMockUC struct {
	agendaFunc func(ctx context.Context) (dto.ReservasMockResponseDTO, []byte, error)
}

func (m *mockReservasMockUC) ObtenerAgenda(ctx context.Context) (dto.ReservasMockResponseDTO, []byte, error) {
	if m.agendaFunc != nil {
		return m.agendaFunc(ctx)
	}
	return dto.ReservasMockResponseDTO{}, nil, nil
}

func TestReservasMockController_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Success 200
	expectedJSON := []byte(`{"fake":true,"aviso":"Agenda ficticia","total":1,"reservas":[{"id":"RES-1","jardin":"Rosales"}]}`)
	uc := &mockReservasMockUC{
		agendaFunc: func(_ context.Context) (dto.ReservasMockResponseDTO, []byte, error) {
			return dto.ReservasMockResponseDTO{
				Fake:  true,
				Aviso: "Agenda ficticia",
				Total: 1,
				Reservas: []dto.ReservaItemDTO{
					{ID: "RES-1", Jardin: "Rosales"},
				},
			}, expectedJSON, nil
		},
	}
	ctrl := controller.NewReservasMockController(uc, zerolog.Nop())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/geo/reservas-mock", nil)
	ctrl.Get(c)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, obtenido %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type esperado 'application/json; charset=utf-8', obtenido %q", ct)
	}
	if w.Body.String() != string(expectedJSON) {
		t.Fatalf("cuerpo inesperado: %s", w.Body.String())
	}

	// 2. Error referencia externa 500
	ucExt := &mockReservasMockUC{
		agendaFunc: func(_ context.Context) (dto.ReservasMockResponseDTO, []byte, error) {
			return dto.ReservasMockResponseDTO{}, nil, domainErrors.ErrAgendaFicticiaReferenciaExterna
		},
	}
	ctrlExt := controller.NewReservasMockController(ucExt, zerolog.Nop())
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/geo/reservas-mock", nil)
	ctrlExt.Get(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("se esperaba 500, obtenido %d", w.Code)
	}
	var errBody map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &errBody)
	if errBody["error"] != "la agenda ficticia todavía arrastra una referencia externa" {
		t.Fatalf("error inesperado: %v", errBody)
	}

	// 3. Error genérico 500
	ucErr := &mockReservasMockUC{
		agendaFunc: func(_ context.Context) (dto.ReservasMockResponseDTO, []byte, error) {
			return dto.ReservasMockResponseDTO{}, nil, errors.New("read error")
		},
	}
	ctrlErr := controller.NewReservasMockController(ucErr, zerolog.Nop())
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/geo/reservas-mock", nil)
	ctrlErr.Get(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("se esperaba 500, obtenido %d", w.Code)
	}
	_ = json.Unmarshal(w.Body.Bytes(), &errBody)
	if errBody["error"] != "no se pudo leer la agenda ficticia" {
		t.Fatalf("error inesperado: %v", errBody)
	}
}
