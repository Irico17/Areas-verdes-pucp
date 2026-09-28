package postgres_test

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"testing"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func migrarDBGeoTemporal(t *testing.T, name string) (*sql.DB, *gorm.DB) {
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

	migDir := testutil.FindMigrationsDir()
	if err := database.Apply(gdb, migDir); err != nil {
		t.Fatalf("error aplicando migraciones: %v", err)
	}

	return rawDB, gdb
}

func TestGeoRepository_Database(t *testing.T) {
	_, gdb := migrarDBGeoTemporal(t, "vp_c_test_geo_repo")
	ctx := context.Background()
	repo := postgres.NewGeoRepository(gdb)

	// Resumen on freshly migrated DB
	res, err := repo.Resumen(ctx)
	if err != nil {
		t.Fatalf("error en Resumen: %v", err)
	}
	if res.CRS != "EPSG:4326" {
		t.Fatalf("CRS esperado EPSG:4326, obtenido: %s", res.CRS)
	}

	// Capas on freshly migrated DB
	capas, err := repo.Capas(ctx)
	if err != nil {
		t.Fatalf("error en Capas: %v", err)
	}
	if capas.Cargadas == nil {
		t.Fatal("cargadas no debe ser nil")
	}

	// Areas on empty DB
	areas, err := repo.Areas(ctx, dto.FiltroGeoDTO{})
	if err != nil {
		t.Fatalf("error en Areas: %v", err)
	}
	if areas.Type != "FeatureCollection" || areas.Name != "areas_verdes" {
		t.Fatalf("FeatureCollection de areas inválida: %+v", areas)
	}

	// Zonas on empty DB
	zonas, err := repo.Zonas(ctx, dto.FiltroGeoDTO{})
	if err != nil {
		t.Fatalf("error en Zonas: %v", err)
	}
	if zonas.Type != "FeatureCollection" || zonas.Name != "zonas" {
		t.Fatalf("FeatureCollection de zonas inválida: %+v", zonas)
	}
}
