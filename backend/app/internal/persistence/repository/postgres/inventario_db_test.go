package postgres_test

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"testing"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func migrarDBInventarioTemporal(t *testing.T, name string) (*sql.DB, *gorm.DB) {
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

func TestInventarioRepository_Database(t *testing.T) {
	rawDB, gdb := migrarDBInventarioTemporal(t, "vp_c_test_inventario")
	defer rawDB.Close()

	ctx := context.Background()
	repo := postgres.NewInventarioRepository(gdb)

	// 1. Index inicialmente vacío
	idx, err := repo.Index(ctx)
	if err != nil {
		t.Fatalf("error en Index inicial: %v", err)
	}
	if len(idx.Capas) != 11 {
		t.Errorf("se esperaban 11 capas conocidas, obtenidas %d", len(idx.Capas))
	}
	if len(idx.Cargadas) != 0 {
		t.Errorf("se esperaban 0 cargadas inicialmente, obtenidas %d", len(idx.Cargadas))
	}

	// 2. Insertamos registros en inventario
	query := `
		INSERT INTO inventario (capa, feature_id, nombre, subtipo, detalle, lugar, foto, geom)
		VALUES
		('bebederos', 'BB-TEST-1', 'Bebedero 1', 'fuente', 'Detalle 1', 'Lugar 1', 'foto1.jpg', ST_SetSRID(ST_MakePoint(-77.08, -12.07), 4326)),
		('bebederos', 'BB-TEST-2', 'Bebedero 2', 'llenador', 'Detalle 2', 'Lugar 2', 'foto2.jpg', ST_SetSRID(ST_MakePoint(-77.081, -12.071), 4326)),
		('fauna', 'FAU-TEST-1', 'Ardilla', 'roedor', 'En árbol', 'Bosque', '', ST_SetSRID(ST_MakePoint(-77.082, -12.072), 4326))`
	if err := gdb.Exec(query).Error; err != nil {
		t.Fatalf("error insertando semillas en inventario: %v", err)
	}

	// 3. Verificar Index con capas cargadas
	idx, err = repo.Index(ctx)
	if err != nil {
		t.Fatalf("error en Index tras insertar: %v", err)
	}
	if len(idx.Cargadas) != 2 {
		t.Fatalf("se esperaban 2 capas cargadas, obtenidas %d", len(idx.Cargadas))
	}
	for _, c := range idx.Cargadas {
		if c.Capa == "bebederos" && c.Features != 2 {
			t.Errorf("se esperaban 2 features en bebederos, obtenidos %d", c.Features)
		}
		if c.Capa == "fauna" && c.Features != 1 {
			t.Errorf("se esperaba 1 feature en fauna, obtenido %d", c.Features)
		}
	}

	// 4. Consultar Capa conocida ("bebederos")
	fc, err := repo.Capa(ctx, "bebederos")
	if err != nil {
		t.Fatalf("error consultando capa bebederos: %v", err)
	}
	if fc.Type != "FeatureCollection" || fc.Name != "bebederos" {
		t.Errorf("metadatos de FeatureCollection inesperados: %+v", fc)
	}
	if len(fc.Features) != 2 {
		t.Fatalf("se esperaban 2 features en bebederos, obtenidos %d", len(fc.Features))
	}

	// 5. Consultar Capa desconocida
	_, errDesc := repo.Capa(ctx, "desconocida")
	if errDesc == nil {
		t.Fatal("se esperaba error consultando capa desconocida")
	}
}
