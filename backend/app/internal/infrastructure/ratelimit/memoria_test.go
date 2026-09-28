package ratelimit

import (
	"testing"
	"time"
)

func TestLimiteDeLogin(t *testing.T) {
	l := NewMemoriaLimitador(2, time.Minute)
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
