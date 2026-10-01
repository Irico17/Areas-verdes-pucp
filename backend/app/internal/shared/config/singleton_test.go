package config

import (
	"reflect"
	"testing"
)

func TestGetConfigSingleton(t *testing.T) {
	resetConfigForTesting()
	defer resetConfigForTesting()

	cfg1 := GetConfig()
	cfg2 := GetConfig()
	if cfg1 != cfg2 {
		t.Fatal("GetConfig() debe retornar la misma instancia singleton")
	}
}

func TestGetConfigIdenticoANew(t *testing.T) {
	t.Setenv("SERVER_PORT", "8095")
	t.Setenv("SERVER_GIN_MODE", "debug")
	t.Setenv("DATABASE_HOST", "127.0.0.1")
	t.Setenv("DATABASE_PORT", "5432")
	t.Setenv("DATABASE_USER", "campus")
	t.Setenv("DATABASE_PASSWORD", "campus")
	t.Setenv("DATABASE_NAME", "campus_verde")
	t.Setenv("DATABASE_SCHEMA", "public")
	t.Setenv("DATABASE_SSL_MODE", "disable")
	t.Setenv("CAMPUS_DEV_PASSWORD", "pando-local")

	resetConfigForTesting()
	defer resetConfigForTesting()

	fromGet := GetConfig()
	fromNew := New()

	if !reflect.DeepEqual(fromGet, fromNew) {
		t.Fatalf("GetConfig() y New() deben producir configuración idéntica:\nGet: %+v\nNew: %+v", fromGet, fromNew)
	}
}

func TestDatabaseSSLMode(t *testing.T) {
	tests := []struct {
		envValue string
		expected string
	}{
		{"", "disable"},
		{"false", "disable"},
		{"0", "disable"},
		{"no", "disable"},
		{"true", "require"},
		{"1", "require"},
		{"yes", "require"},
		{"disable", "disable"},
		{"require", "require"},
		{"verify-ca", "verify-ca"},
		{"verify-full", "verify-full"},
	}

	for _, tc := range tests {
		t.Run("val_"+tc.envValue, func(t *testing.T) {
			t.Setenv("DATABASE_SSL_MODE", tc.envValue)
			got := databaseSSLMode()
			if got != tc.expected {
				t.Fatalf("con DATABASE_SSL_MODE=%q esperado %q, obtenido %q", tc.envValue, tc.expected, got)
			}
		})
	}
}
