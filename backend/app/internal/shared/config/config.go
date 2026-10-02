// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"net/url"
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
	// AppEnv is develop, qa, produccion, or empty when APP_ENV is unset.
	// Empty keeps the historical defaults (pool 10/4/30 min, Swagger by Gin mode).
	AppEnv      string
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
	LogLevel       string
	LogFormat      string
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
	Dir     string
	Semilla string
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

	envName := appEnv()
	ginMode := ginModeFor(envName)
	logLevel, logFormat := logProfile(envName, ginMode)
	openConns, idleConns, lifetime := poolFor(envName)

	rawDir := valueOrDefault("DATA_RAW_DIR", filepath.Join(root, "data", "raw"))

	return &Config{
		AppEnv: envName,
		Server: ServerConfig{
			Port:           serverPort(),
			GinMode:        ginMode,
			LogLevel:       logLevel,
			LogFormat:      logFormat,
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
			MaxOpenConns:    intEnv("DATABASE_MAX_OPEN_CONNS", openConns),
			MaxIdleConns:    intEnv("DATABASE_MAX_IDLE_CONNS", idleConns),
			ConnMaxLifetime: connMaxLifetime(lifetime),
		},
		Seguridad: SeguridadConfig{
			CORSOrigins:    origenesCORS(corsRaw(envName)),
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
			Dir:     valueOrDefault("MIGRATIONS_DIR", filepath.Join(root, "db", "migrations")),
			Semilla: valueOrDefault("SEED_FILE", filepath.Join(root, "deploy", "seed", "ficticio.sql")),
		},
		Accesos: AccesosConfig{
			DevPassword: os.Getenv("CAMPUS_DEV_PASSWORD"),
		},
		Swagger: SwaggerConfig{
			Enabled: swaggerEnabled(envName, ginMode),
			Host:    os.Getenv("SWAGGER_HOST"),
		},
	}
}

// Validar checks configuration invariants.
// In produccion (APP_ENV=produccion), laboratory passwords (pando-local, campus-lab)
// or passwords shorter than 16 characters are rejected for CAMPUS_DEV_PASSWORD
// and for PostgreSQL credentials.
func (c *Config) Validar() error {
	if c.AppEnv != "produccion" {
		return nil
	}

	devPwd := strings.TrimSpace(c.Accesos.DevPassword)
	if devPwd == "pando-local" || devPwd == "campus-lab" {
		return fmt.Errorf("en producción, CAMPUS_DEV_PASSWORD no puede ser una clave de laboratorio")
	}
	if len(devPwd) < 16 {
		return fmt.Errorf("en producción, CAMPUS_DEV_PASSWORD debe tener al menos 16 caracteres (longitud actual: %d)", len(devPwd))
	}

	if c.Database.URL != "" {
		if parsed, err := url.Parse(c.Database.URL); err == nil && parsed.User != nil {
			if pwd, ok := parsed.User.Password(); ok {
				pgPwd := strings.TrimSpace(pwd)
				if pgPwd == "pando-local" || pgPwd == "campus-lab" {
					return fmt.Errorf("en producción, la clave de Postgres no puede ser una clave de laboratorio")
				}
				if len(pgPwd) < 16 {
					return fmt.Errorf("en producción, la clave de Postgres debe tener al menos 16 caracteres (longitud actual: %d)", len(pgPwd))
				}
			}
		}
	} else if os.Getenv("DATABASE_PASSWORD") != "" {
		pgPwd := strings.TrimSpace(c.Database.Password)
		if pgPwd == "pando-local" || pgPwd == "campus-lab" {
			return fmt.Errorf("en producción, la clave de Postgres no puede ser una clave de laboratorio")
		}
		if len(pgPwd) < 16 {
			return fmt.Errorf("en producción, la clave de Postgres debe tener al menos 16 caracteres (longitud actual: %d)", len(pgPwd))
		}
	}

	return nil
}

// appEnv normalizes APP_ENV. Unset or unknown leaves the historical behavior.
func appEnv() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV"))) {
	case "develop", "development", "dev":
		return "develop"
	case "qa", "staging":
		return "qa"
	case "produccion", "producción", "production", "prod":
		return "produccion"
	default:
		return ""
	}
}

func ginModeFor(envName string) string {
	if value := strings.TrimSpace(os.Getenv("SERVER_GIN_MODE")); value != "" {
		return value
	}
	if envName == "develop" {
		return "debug"
	}
	return "release"
}

func logProfile(envName, ginMode string) (level, format string) {
	level = strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	format = strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT")))
	if level == "" || format == "" {
		defLevel, defFormat := "info", "json"
		switch envName {
		case "develop":
			defLevel, defFormat = "debug", "console"
		case "qa", "produccion":
			defLevel, defFormat = "info", "json"
		default:
			if ginMode == "debug" {
				defLevel, defFormat = "debug", "console"
			}
		}
		if level == "" {
			level = defLevel
		}
		if format == "" {
			format = defFormat
		}
	}
	switch level {
	case "debug", "info", "warn", "warning", "error":
	default:
		level = "info"
	}
	if level == "warning" {
		level = "warn"
	}
	if format != "console" {
		format = "json"
	}
	return level, format
}

func poolFor(envName string) (open, idle int, lifetime time.Duration) {
	switch envName {
	case "develop":
		return 5, 2, 15 * time.Minute
	case "qa":
		return 8, 3, 20 * time.Minute
	default:
		// produccion y APP_ENV vacío: el pool histórico (no el 10/100 del equipo).
		return 10, 4, 30 * time.Minute
	}
}

func intEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n := 0
	for _, r := range raw {
		if r < '0' || r > '9' {
			return fallback
		}
		n = n*10 + int(r-'0')
	}
	if n < 1 {
		return fallback
	}
	return n
}

func connMaxLifetime(fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv("DATABASE_CONN_MAX_LIFETIME"))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

func corsRaw(envName string) string {
	if raw := strings.TrimSpace(os.Getenv("CAMPUS_CORS_ORIGINS")); raw != "" {
		return raw
	}
	switch envName {
	case "develop":
		return "http://127.0.0.1:8088,http://localhost:8088,http://127.0.0.1:4317,http://localhost:4317"
	case "qa":
		return "http://127.0.0.1:8188,http://localhost:8188"
	default:
		return ""
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

func swaggerEnabled(envName, ginMode string) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv("SWAGGER_ENABLED")))
	if raw != "" {
		switch raw {
		case "1", "true", "yes", "si", "sí":
			return true
		case "0", "false", "no":
			return false
		}
	}
	switch envName {
	case "develop", "qa":
		return true
	case "produccion":
		return false
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
