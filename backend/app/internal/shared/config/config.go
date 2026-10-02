// Package config loads application configuration from environment variables.
package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

var (
	instance *Config
	once     sync.Once
)

// GetConfig returns the singleton application configuration.
func GetConfig() *Config {
	once.Do(func() {
		instance = New()
	})
	return instance
}

// resetConfigForTesting resets the singleton instance for unit testing.
func resetConfigForTesting() {
	instance = nil
	once = sync.Once{}
}

// Config contains all application configuration.
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Seguridad   SeguridadConfig
	Evidencias  EvidenciasConfig
	Datos       DatosConfig
	Migraciones MigracionesConfig
	Accesos     AccesosConfig
	Swagger     SwaggerConfig
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Port           string
	GinMode        string
	TrustedProxies []string
}

// DatabaseConfig contains PostgreSQL connection settings.
type DatabaseConfig struct {
	URL             string
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	Schema          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// SeguridadConfig contains security and access settings.
type SeguridadConfig struct {
	CORSOrigins    []string
	CookieSecure   bool
	CookieSameSite string
	LoginMax       int
	LoginVentana   time.Duration
}

// EvidenciasConfig contains file storage settings.
type EvidenciasConfig struct {
	Dir    string
	Bucket string
}

// DatosConfig contains static and reference data file paths.
type DatosConfig struct {
	RawDir        string
	V1Dir         string
	EdificiosPath string
	ReservasPath  string
	FotosDir      string
	OpenAPIPath   string
}

// MigracionesConfig contains database migrations path.
type MigracionesConfig struct {
	Dir string
}

// AccesosConfig contains access seeds and credentials.
type AccesosConfig struct {
	DevPassword string
}

// SwaggerConfig contains Swagger documentation settings.
type SwaggerConfig struct {
	Enabled bool
	Host    string
}

// New builds configuration from the process environment with local defaults.
func New() *Config {
	_ = godotenv.Load()
	root := findRepoRoot()
	if root != "." {
		_ = godotenv.Load(filepath.Join(root, ".env"))
	}

	ginMode := valueOrDefault("SERVER_GIN_MODE", "release")

	rawDir := valueOrDefault("DATA_RAW_DIR", filepath.Join(root, "data", "raw"))

	return &Config{
		Server: ServerConfig{
			Port:           serverPort(),
			GinMode:        ginMode,
			TrustedProxies: trustedProxies(),
		},
		Database: DatabaseConfig{
			URL:             os.Getenv("DATABASE_URL"),
			Host:            valueOrDefault("DATABASE_HOST", "localhost"),
			Port:            valueOrDefault("DATABASE_PORT", "5432"),
			User:            valueOrDefault("DATABASE_USER", "areasverdes"),
			Password:        valueOrDefault("DATABASE_PASSWORD", "areasverdes"),
			Name:            valueOrDefault("DATABASE_NAME", "areasverdes"),
			Schema:          valueOrDefault("DATABASE_SCHEMA", "public"),
			SSLMode:         databaseSSLMode(),
			MaxOpenConns:    10,
			MaxIdleConns:    4,
			ConnMaxLifetime: 30 * time.Minute,
		},
		Seguridad: SeguridadConfig{
			CORSOrigins:    origenesCORS(os.Getenv("CAMPUS_CORS_ORIGINS")),
			CookieSecure:   cookieSecure(),
			CookieSameSite: valueOrDefault("CAMPUS_COOKIE_SAMESITE", "Lax"),
			LoginMax:       loginMax(),
			LoginVentana:   time.Minute,
		},
		Evidencias: EvidenciasConfig{
			Dir:    valueOrDefault("EVIDENCIAS_DIR", filepath.Join(root, "data", "evidencias")),
			Bucket: os.Getenv("EVIDENCIAS_BUCKET"),
		},
		Datos: DatosConfig{
			RawDir:        rawDir,
			V1Dir:         valueOrDefault("DATA_V1_DIR", filepath.Join(root, "data", "v1")),
			EdificiosPath: valueOrDefault("EDIFICIOS_PATH", filepath.Join(root, "data", "osm", "edificios_pando.geojson")),
			ReservasPath:  valueOrDefault("RESERVAS_MOCK_PATH", filepath.Join(root, "data", "mocks", "reservas_agenda.mock.json")),
			FotosDir:      valueOrDefault("DRIVE_FOTOS_DIR", filepath.Join(rawDir, "drive_fotos")),
			OpenAPIPath:   valueOrDefault("OPENAPI_PATH", filepath.Join(root, "apps", "api", "openapi.yaml")),
		},
		Migraciones: MigracionesConfig{
			Dir: valueOrDefault("MIGRATIONS_DIR", filepath.Join(root, "apps", "api", "migrations")),
		},
		Accesos: AccesosConfig{
			DevPassword: os.Getenv("CAMPUS_DEV_PASSWORD"),
		},
		Swagger: SwaggerConfig{
			Enabled: swaggerEnabled(ginMode),
			Host:    os.Getenv("SWAGGER_HOST"),
		},
	}
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func databaseSSLMode() string {
	raw := strings.TrimSpace(os.Getenv("DATABASE_SSL_MODE"))
	if raw == "" {
		return "disable"
	}
	switch strings.ToLower(raw) {
	case "false", "0", "no":
		return "disable"
	case "true", "1", "yes":
		return "require"
	default:
		return raw
	}
}

func serverPort() string {
	if apiAddr := strings.TrimSpace(os.Getenv("API_ADDR")); apiAddr != "" {
		if idx := strings.LastIndex(apiAddr, ":"); idx != -1 {
			return apiAddr[idx+1:]
		}
		return apiAddr
	}
	raw := strings.TrimSpace(valueOrDefault("SERVER_PORT", "8080"))
	return strings.TrimPrefix(raw, ":")
}

func origenesCORS(raw string) []string {
	out := []string{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" || part == "*" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func cookieSecure() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("CAMPUS_COOKIE_SECURE"))) {
	case "1", "true", "yes", "si", "sí":
		return true
	default:
		return false
	}
}

func loginMax() int {
	raw := strings.TrimSpace(os.Getenv("CAMPUS_LOGIN_MAX"))
	if raw == "" {
		return 8
	}
	n := 0
	for _, r := range raw {
		if r < '0' || r > '9' {
			return 8
		}
		n = n*10 + int(r-'0')
	}
	if n < 1 {
		return 8
	}
	return n
}

func trustedProxies() []string {
	raw := os.Getenv("SERVER_TRUSTED_PROXIES")
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func swaggerEnabled(ginMode string) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("SWAGGER_ENABLED")))
	if raw != "" {
		switch raw {
		case "1", "true", "yes", "si", "sí":
			return true
		case "0", "false", "no":
			return false
		}
	}
	return ginMode != "release"
}

func findRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := wd
	for {
		if fileExists(filepath.Join(dir, "data", "raw")) && (fileExists(filepath.Join(dir, "apps", "api")) || fileExists(filepath.Join(dir, "backend"))) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return wd
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
