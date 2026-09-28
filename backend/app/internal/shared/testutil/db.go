package testutil

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// MigrarDBTemporal creates a throwaway database with migrations applied, registers cleanup, and returns connections.
// The database name starts with "vp_c_test_" and ends with a random hex suffix to prevent race conditions.
func MigrarDBTemporal(t *testing.T, prefix string) (*sql.DB, *gorm.DB) {
	t.Helper()
	base := os.Getenv("MIGRATE_TEST_URL")
	if base == "" {
		t.Skip("MIGRATE_TEST_URL no configurada; omitiendo test de base de datos")
	}

	admin, err := sql.Open("pgx", base)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		t.Fatalf("sin postgres de prueba: %v", err)
	}

	cleanPrefix := strings.TrimPrefix(prefix, "vp_c_test_")
	cleanPrefix = strings.TrimPrefix(cleanPrefix, "vp_c_")
	cleanPrefix = strings.Trim(cleanPrefix, "_")
	if len(cleanPrefix) > 20 {
		cleanPrefix = cleanPrefix[:20]
	}

	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	randHex := hex.EncodeToString(b)

	var name string
	if cleanPrefix != "" {
		name = fmt.Sprintf("vp_c_test_%s_%s", cleanPrefix, randHex)
	} else {
		name = fmt.Sprintf("vp_c_test_%s", randHex)
	}

	if _, err := admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("DROP DATABASE IF EXISTS " + name); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		admin2, err := sql.Open("pgx", base)
		if err != nil {
			return
		}
		defer admin2.Close()
		_, _ = admin2.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name)
		_, _ = admin2.Exec("DROP DATABASE IF EXISTS " + name)
	})

	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name

	gdb, err := database.NewConnection(&config.Config{
		Database: config.DatabaseConfig{
			URL: u.String(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	rawDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}

	migDir := FindMigrationsDir()
	if err := database.Apply(gdb, migDir); err != nil {
		t.Fatalf("error aplicando migraciones: %v", err)
	}

	return rawDB, gdb
}
