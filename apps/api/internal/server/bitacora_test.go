package server

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBitacoraNoEscribeCuerpoNiCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(prev) })

	r := gin.New()
	r.Use(Bitacora())
	r.POST("/api/v1/evidencias", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/evidencias?token=secreto-query", strings.NewReader("cuerpo-archivo-binario"))
	req.Header.Set("Cookie", "cv_sesion=secreto-cookie")
	req.Header.Set("Authorization", "Bearer secreto-bearer")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status %d", w.Code)
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("falta X-Request-ID")
	}
	linea := buf.String()
	if !strings.Contains(linea, "request_id=") || !strings.Contains(linea, "path=/api/v1/evidencias") {
		t.Fatalf("log %q", linea)
	}
	for _, prohibido := range []string{"cuerpo-archivo-binario", "secreto-cookie", "secreto-bearer", "secreto-query"} {
		if strings.Contains(linea, prohibido) {
			t.Fatalf("el log incluye %q: %s", prohibido, linea)
		}
	}
}
