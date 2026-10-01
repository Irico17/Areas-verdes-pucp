package database_test

import (
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	appErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func TestNewDatabaseConnection(t *testing.T) {
	dbURL := os.Getenv("MIGRATE_TEST_URL")
	if dbURL == "" {
		t.Skip("MIGRATE_TEST_URL no definido: se omite la prueba con base de datos")
	}
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			URL:             dbURL,
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 10 * time.Minute,
		},
	}
	gdb, err := database.NewDatabaseConnection(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("database not reachable: %v", err)
	}
	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 5 {
		t.Fatalf("expected MaxOpenConnections 5, got %d", stats.MaxOpenConnections)
	}

	// Test alias NewConnection
	aliasGDB, err := database.NewConnection(cfg)
	if err != nil {
		t.Fatalf("NewConnection alias failed: %v", err)
	}
	aliasSQL, err := aliasGDB.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB from alias: %v", err)
	}
	defer aliasSQL.Close()
	if err := aliasSQL.Ping(); err != nil {
		t.Fatalf("alias connection ping failed: %v", err)
	}
}

func TestNewDatabaseConnectionFailsWithApplicationError(t *testing.T) {
	// Connect to an invalid unreachable port with a fast failure
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			URL: "postgres://campus:campus@127.0.0.1:54399/no_existe?connect_timeout=1&sslmode=disable",
		},
	}
	_, err := database.NewDatabaseConnection(cfg)
	if err == nil {
		t.Fatal("esperaba error al conectar a un puerto inexistente")
	}

	var appErr *appErrors.ApplicationError
	if !errors.As(err, &appErr) {
		t.Fatalf("se esperaba un error de tipo *ApplicationError, obtenido: %T (%v)", err, err)
	}
	if appErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("se esperaba StatusCode 503, obtenido %d", appErr.StatusCode)
	}
	if appErr.Code != appErrors.DBDatabaseConnection {
		t.Fatalf("se esperaba Code %q, obtenido %q", appErrors.DBDatabaseConnection, appErr.Code)
	}
}
