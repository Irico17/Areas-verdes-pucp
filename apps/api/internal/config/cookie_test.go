package config

import (
	"testing"
)

func TestCookieSecureSoloSiSePide(t *testing.T) {
	t.Setenv("CAMPUS_ENV", "production")
	t.Setenv("CAMPUS_COOKIE_SECURE", "")
	if cookieSecure() {
		t.Fatal("producción por HTTP no debe marcar la cookie Secure")
	}
	t.Setenv("CAMPUS_COOKIE_SECURE", "false")
	if cookieSecure() {
		t.Fatal("false explícito tiene que apagar Secure")
	}
	t.Setenv("CAMPUS_COOKIE_SECURE", "true")
	if !cookieSecure() {
		t.Fatal("true explícito tiene que encender Secure")
	}
}
