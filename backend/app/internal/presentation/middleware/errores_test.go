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
	r.GET("/error-db-conn", func(c *gin.Context) {
		_ = c.Error(domainErrors.NewApplicationError(domainErrors.DBDatabaseConnection, http.StatusServiceUnavailable, errors.New("timeout")))
	})
	r.GET("/error-db-query", func(c *gin.Context) {
		_ = c.Error(domainErrors.NewApplicationError(domainErrors.DBDatabaseQuery, 0, errors.New("syntax error")))
	})
	r.GET("/error-db-generic", func(c *gin.Context) {
		_ = c.Error(domainErrors.NewApplicationError(domainErrors.DBDatabaseError, 0, errors.New("corrupt table")))
	})
	r.GET("/error-srv-internal", func(c *gin.Context) {
		_ = c.Error(domainErrors.NewApplicationError(domainErrors.SrvInternalServer, 0, errors.New("panic")))
	})
	r.GET("/error-app-custom", func(c *gin.Context) {
		_ = c.Error(domainErrors.NewApplicationErrorWithMessage("CUSTOM-100", http.StatusForbidden, "acceso denegado"))
	})
	r.GET("/error-generico", func(c *gin.Context) {
		_ = c.Error(errors.New("fallo inesperado"))
	})

	tests := []struct {
		path            string
		expectedStatus  int
		expectedMessage string
	}{
		{"/error-app", http.StatusBadRequest, "datos inválidos"},
		{"/error-db-conn", http.StatusServiceUnavailable, "base de datos no disponible"},
		{"/error-db-query", http.StatusInternalServerError, "error de base de datos"},
		{"/error-db-generic", http.StatusInternalServerError, "error de base de datos"},
		{"/error-srv-internal", http.StatusInternalServerError, "error interno"},
		{"/error-app-custom", http.StatusForbidden, "acceso denegado"},
		{"/error-generico", http.StatusInternalServerError, "error interno"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			r.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("se esperaba status %d, se obtuvo %d", tc.expectedStatus, w.Code)
			}
			var res map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
				t.Fatalf("error deserializando JSON: %v", err)
			}
			if res["error"] != tc.expectedMessage {
				t.Fatalf("mensaje esperado %q, obtenido %q", tc.expectedMessage, res["error"])
			}
		})
	}
}
