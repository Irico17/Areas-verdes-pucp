package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

func TestErroresMiddlewareTraduceErrores(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Errores(zerolog.Nop()))

	r.GET("/error-app", func(c *gin.Context) {
		_ = c.Error(domainErrors.NewAppError(http.StatusBadRequest, "datos inválidos"))
	})
	r.GET("/error-generico", func(c *gin.Context) {
		_ = c.Error(errors.New("fallo inesperado"))
	})

	// Test AppError
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/error-app", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("se esperaba status 400, se obtuvo %d", w.Code)
	}
	var res map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("error deserializando JSON: %v", err)
	}
	if res["error"] != "datos inválidos" {
		t.Fatalf("mensaje esperado 'datos inválidos', obtenido %q", res["error"])
	}

	// Test generic error
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/error-generico", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("se esperaba status 500, se obtuvo %d", w.Code)
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("error deserializando JSON: %v", err)
	}
	if res["error"] != "error interno" {
		t.Fatalf("mensaje esperado 'error interno', obtenido %q", res["error"])
	}
}
