package config

import (
	"testing"
)

func TestServerPortConAPIAddrYPort(t *testing.T) {
	t.Setenv("SERVER_PORT", "8080")
	t.Setenv("API_ADDR", "")
	if got := serverPort(); got != "8080" {
		t.Fatalf("esperado 8080, obtenido %s", got)
	}

	t.Setenv("SERVER_PORT", ":8082")
	t.Setenv("API_ADDR", "")
	if got := serverPort(); got != "8082" {
		t.Fatalf("esperado 8082 (sin dos puntos), obtenido %s", got)
	}

	t.Setenv("SERVER_PORT", "8080")
	t.Setenv("API_ADDR", ":8091")
	if got := serverPort(); got != "8091" {
		t.Fatalf("API_ADDR :8091 debe sobrescribir SERVER_PORT: obtenido %s", got)
	}

	t.Setenv("API_ADDR", "127.0.0.1:8093")
	if got := serverPort(); got != "8093" {
		t.Fatalf("API_ADDR host:port debe retornar 8093: obtenido %s", got)
	}
}
