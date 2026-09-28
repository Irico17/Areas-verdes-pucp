package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/ratelimit"
)

func TestLimiteLoginCortaExcesos(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	_ = r.SetTrustedProxies([]string{})
	lim := ratelimit.NewMemoriaLimitador(2, time.Minute)
	r.Use(LimiteLogin(lim))

	r.POST("/areas-verdes/v1/sesion", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/areas-verdes/v1/sesion", nil)
		r.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			t.Fatalf("intento %d cortado antes de tiempo", i)
		}
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/areas-verdes/v1/sesion", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("se esperaba 429 Too Many Requests, se obtuvo %d (%s)", w.Code, w.Body.String())
	}
}

func TestLimiteLoginNoConfiaEnXForwardedFor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Disable trusted proxies
	_ = r.SetTrustedProxies([]string{})
	lim := ratelimit.NewMemoriaLimitador(2, time.Minute)
	r.Use(LimiteLogin(lim))

	r.POST("/areas-verdes/v1/sesion", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Make 2 requests with different X-Forwarded-For headers from the same RemoteAddr
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/areas-verdes/v1/sesion", nil)
		req.Header.Set("X-Forwarded-For", "1.2.3."+string(rune('1'+i)))
		req.RemoteAddr = "192.0.2.1:1234"
		r.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			t.Fatalf("intento %d cortado antes de tiempo", i)
		}
	}

	// 3rd request with another X-Forwarded-For should be blocked because RemoteAddr is the same
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/areas-verdes/v1/sesion", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.99")
	req.RemoteAddr = "192.0.2.1:1234"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("X-Forwarded-For no debió evadir el límite; status: %d (%s)", w.Code, w.Body.String())
	}
}
