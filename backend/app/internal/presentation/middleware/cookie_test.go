package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCookieHttpOnlyYSecureConfigurable(t *testing.T) {
	rec := httptest.NewRecorder()
	EscribirCookie(rec, OpcionesCookie{Secure: true, SameSite: http.SameSiteStrictMode}, "tok", 60)
	res := rec.Result()
	defer res.Body.Close()
	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != CookieSesion || !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteStrictMode {
		t.Fatalf("cookie %+v", c)
	}
}

func TestConCookieMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ConCookie(OpcionesCookie{Secure: true, SameSite: http.SameSiteStrictMode}))
	var opts OpcionesCookie
	r.GET("/test", func(c *gin.Context) {
		opts = ObtenerOpcionesCookie(c)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if !opts.Secure || opts.SameSite != http.SameSiteStrictMode {
		t.Fatalf("opciones cookie en contexto incorrectas: %+v", opts)
	}
}
