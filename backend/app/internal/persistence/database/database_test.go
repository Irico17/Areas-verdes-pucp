package database_test

import (
	"os"
	"testing"
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func TestNewConnection(t *testing.T) {
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
	gdb, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("skipping database connection test: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		t.Skipf("database not reachable: %v", err)
	}
	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 5 {
		t.Fatalf("expected MaxOpenConnections 5, got %d", stats.MaxOpenConnections)
	}
}
