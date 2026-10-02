package config

import (
	"testing"
	"time"
)

func TestAppEnvVacioConservaHistorico(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("SERVER_GIN_MODE", "release")
	t.Setenv("SWAGGER_ENABLED", "")
	t.Setenv("CAMPUS_CORS_ORIGINS", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "")
	t.Setenv("DATABASE_CONN_MAX_LIFETIME", "")

	cfg := New()
	if cfg.AppEnv != "" {
		t.Fatalf("APP_ENV vacío debe quedar vacío, obtuvo %q", cfg.AppEnv)
	}
	if cfg.Swagger.Enabled {
		t.Fatal("sin APP_ENV y en release, Swagger sigue apagado")
	}
	if cfg.Database.MaxOpenConns != 10 || cfg.Database.MaxIdleConns != 4 || cfg.Database.ConnMaxLifetime != 30*time.Minute {
		t.Fatalf("pool histórico 10/4/30m, obtuvo %d/%d/%s", cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.ConnMaxLifetime)
	}
	if len(cfg.Seguridad.CORSOrigins) != 0 {
		t.Fatalf("sin APP_ENV el CORS sigue vacío: %v", cfg.Seguridad.CORSOrigins)
	}
	if cfg.Server.LogLevel != "info" || cfg.Server.LogFormat != "json" {
		t.Fatalf("release sin APP_ENV usa info/json, obtuvo %s/%s", cfg.Server.LogLevel, cfg.Server.LogFormat)
	}
}

func TestAppEnvDevelopQaProduccion(t *testing.T) {
	t.Setenv("SERVER_GIN_MODE", "")
	t.Setenv("SWAGGER_ENABLED", "")
	t.Setenv("CAMPUS_CORS_ORIGINS", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "")
	t.Setenv("DATABASE_CONN_MAX_LIFETIME", "")

	t.Setenv("APP_ENV", "develop")
	dev := New()
	if dev.AppEnv != "develop" || dev.Server.GinMode != "debug" {
		t.Fatalf("develop: %+v", dev.Server)
	}
	if !dev.Swagger.Enabled {
		t.Fatal("develop enciende Swagger")
	}
	if dev.Database.MaxOpenConns != 5 || dev.Database.MaxIdleConns != 2 || dev.Database.ConnMaxLifetime != 15*time.Minute {
		t.Fatalf("pool develop 5/2/15m, obtuvo %d/%d/%s", dev.Database.MaxOpenConns, dev.Database.MaxIdleConns, dev.Database.ConnMaxLifetime)
	}
	if dev.Server.LogLevel != "debug" || dev.Server.LogFormat != "console" {
		t.Fatalf("develop usa debug/console, obtuvo %s/%s", dev.Server.LogLevel, dev.Server.LogFormat)
	}
	if !contains(dev.Seguridad.CORSOrigins, "http://127.0.0.1:8088") {
		t.Fatalf("CORS develop: %v", dev.Seguridad.CORSOrigins)
	}

	t.Setenv("APP_ENV", "qa")
	qa := New()
	if qa.AppEnv != "qa" || !qa.Swagger.Enabled || qa.Server.GinMode != "release" {
		t.Fatalf("qa: env=%s swagger=%v gin=%s", qa.AppEnv, qa.Swagger.Enabled, qa.Server.GinMode)
	}
	if qa.Database.MaxOpenConns != 8 || qa.Database.MaxIdleConns != 3 || qa.Database.ConnMaxLifetime != 20*time.Minute {
		t.Fatalf("pool qa 8/3/20m, obtuvo %d/%d/%s", qa.Database.MaxOpenConns, qa.Database.MaxIdleConns, qa.Database.ConnMaxLifetime)
	}
	if !contains(qa.Seguridad.CORSOrigins, "http://127.0.0.1:8188") {
		t.Fatalf("CORS qa: %v", qa.Seguridad.CORSOrigins)
	}
	if qa.Server.LogLevel != "info" || qa.Server.LogFormat != "json" {
		t.Fatalf("qa usa info/json, obtuvo %s/%s", qa.Server.LogLevel, qa.Server.LogFormat)
	}

	t.Setenv("APP_ENV", "production")
	prod := New()
	if prod.AppEnv != "produccion" || prod.Swagger.Enabled {
		t.Fatalf("produccion apaga Swagger: env=%s swagger=%v", prod.AppEnv, prod.Swagger.Enabled)
	}
	if prod.Database.MaxOpenConns != 10 || prod.Database.MaxIdleConns != 4 || prod.Database.ConnMaxLifetime != 30*time.Minute {
		t.Fatalf("pool produccion 10/4/30m, obtuvo %d/%d/%s", prod.Database.MaxOpenConns, prod.Database.MaxIdleConns, prod.Database.ConnMaxLifetime)
	}
	if len(prod.Seguridad.CORSOrigins) != 0 {
		t.Fatalf("produccion no inventa orígenes CORS: %v", prod.Seguridad.CORSOrigins)
	}
}

func TestOverridesExplicitosGananAAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "produccion")
	t.Setenv("SWAGGER_ENABLED", "true")
	t.Setenv("CAMPUS_CORS_ORIGINS", "https://campus.example")
	t.Setenv("SERVER_GIN_MODE", "debug")
	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("LOG_FORMAT", "console")
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "12")
	t.Setenv("DATABASE_MAX_IDLE_CONNS", "6")
	t.Setenv("DATABASE_CONN_MAX_LIFETIME", "45m")

	cfg := New()
	if !cfg.Swagger.Enabled {
		t.Fatal("SWAGGER_ENABLED=true gana a produccion")
	}
	if len(cfg.Seguridad.CORSOrigins) != 1 || cfg.Seguridad.CORSOrigins[0] != "https://campus.example" {
		t.Fatalf("CORS explícito: %v", cfg.Seguridad.CORSOrigins)
	}
	if cfg.Server.GinMode != "debug" || cfg.Server.LogLevel != "warn" || cfg.Server.LogFormat != "console" {
		t.Fatalf("overrides de server: %+v", cfg.Server)
	}
	if cfg.Database.MaxOpenConns != 12 || cfg.Database.MaxIdleConns != 6 || cfg.Database.ConnMaxLifetime != 45*time.Minute {
		t.Fatalf("overrides de pool: %d/%d/%s", cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.ConnMaxLifetime)
	}

	t.Setenv("APP_ENV", "develop")
	t.Setenv("SWAGGER_ENABLED", "false")
	cfg = New()
	if cfg.Swagger.Enabled {
		t.Fatal("SWAGGER_ENABLED=false apaga Swagger también en develop")
	}
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
