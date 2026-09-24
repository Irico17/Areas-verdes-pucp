package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Config reúne la configuración de proceso. Los secretos vienen de entorno, nunca del repo.
type Config struct {
	DatabaseURL   string
	APIAddr       string
	RepoRoot      string
	RawDir        string
	V1Dir         string
	MigrationsDir string
	OpenAPIPath   string
}

// Load lee .env del raíz del repo (sin pisar variables ya exportadas) y aplica defaults locales.
func Load() Config {
	root := findRepoRoot()
	loadEnvFile(filepath.Join(root, ".env"))

	return Config{
		DatabaseURL:   env("DATABASE_URL", "postgres://campus:campus@127.0.0.1:5432/campus_verde?sslmode=disable"),
		APIAddr:       env("API_ADDR", ":8091"),
		RepoRoot:      root,
		RawDir:        env("DATA_RAW_DIR", filepath.Join(root, "data", "raw")),
		V1Dir:         env("DATA_V1_DIR", filepath.Join(root, "data", "v1")),
		MigrationsDir: env("MIGRATIONS_DIR", filepath.Join(root, "apps", "api", "migrations")),
		OpenAPIPath:   env("OPENAPI_PATH", filepath.Join(root, "apps", "api", "openapi.yaml")),
	}
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func findRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := wd
	for {
		if fileExists(filepath.Join(dir, "data", "raw")) && fileExists(filepath.Join(dir, "apps", "api")) {
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

// loadEnvFile aplica KEY=VALUE si la variable aún no está en el entorno.
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, val)
	}
}
