package controller_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

type mockIAUseCase struct {
	sugerirFn func(ctx context.Context, titulo string) dto.SugerenciaIADTO
}

func (m *mockIAUseCase) Sugerir(ctx context.Context, titulo string) dto.SugerenciaIADTO {
	if m.sugerirFn != nil {
		return m.sugerirFn(ctx, titulo)
	}
	return dto.SugerenciaIADTO{
		Codigo:         "riego",
		Etiqueta:       "Riego",
		Explicacion:    "Regla local sobre el título. No es un modelo externo: confirme antes de guardar.",
		Confianza:      "baja",
		RequiereHumano: true,
	}
}

func TestIAController_Sugerir(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &mockIAUseCase{}
	ctrl := controller.NewIAController(uc, zerolog.Nop())

	r := gin.New()
	r.POST("/ia/sugerir-tipo", ctrl.Sugerir)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ia/sugerir-tipo", strings.NewReader(`{"titulo":"Revisar aspersores del eje"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado status 200, obtenido %d", w.Code)
	}
	if !stringsContains(w.Body.String(), `"codigo":"riego"`) {
		t.Fatalf("cuerpo inesperado: %s", w.Body.String())
	}
}

func TestIAController_JSONInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &mockIAUseCase{}
	ctrl := controller.NewIAController(uc, zerolog.Nop())

	r := gin.New()
	r.POST("/ia/sugerir-tipo", ctrl.Sugerir)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/ia/sugerir-tipo", strings.NewReader(`{invalido`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperado status 400, obtenido %d", w.Code)
	}
	if !stringsContains(w.Body.String(), `"error":"JSON inválido"`) {
		t.Fatalf("cuerpo inesperado: %s", w.Body.String())
	}
}
