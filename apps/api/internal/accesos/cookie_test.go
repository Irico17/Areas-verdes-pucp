package accesos

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
	if c.Name != Cookie || !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteStrictMode {
		t.Fatalf("cookie %+v", c)
	}
}

func TestLimiteDeLogin(t *testing.T) {
	l := NuevoLimite(2, time.Minute)
	if !l.Permitir("a") || !l.Permitir("a") {
		t.Fatal("los dos primeros pasan")
	}
	if l.Permitir("a") {
		t.Fatal("el tercero debe cortar")
	}
	if !l.Permitir("b") {
		t.Fatal("otra clave sigue libre")
	}
}
