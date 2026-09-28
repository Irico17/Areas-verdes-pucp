package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"testing"

	"gorm.io/gorm"

	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func migrarDBCatalogoTemporal(t *testing.T, name string) (*sql.DB, *gorm.DB) {
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
		t.Skipf("sin postgres de prueba: %v", err)
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

	migDir := findMigrationsDir()
	if err := database.Apply(gdb, migDir); err != nil {
		t.Fatalf("error aplicando migraciones: %v", err)
	}

	return rawDB, gdb
}

func TestCatalogoRepository_CRUD(t *testing.T) {
	rawDB, gdb := migrarDBCatalogoTemporal(t, "vp_c_test_catalogo_repo")
	_ = rawDB
	ctx := context.Background()
	repo := postgres.NewCatalogoRepository(gdb)

	// 1. Create new item
	item, err := repo.Create(ctx, "tipo_actividad", "nuevo_riego", "Riego Tecnificado")
	if err != nil {
		t.Fatalf("error creando item de catalogo: %v", err)
	}
	if item.ID == 0 || item.Clase != "tipo_actividad" || item.Codigo != "nuevo_riego" || item.Nombre != "Riego Tecnificado" || !item.Activo {
		t.Fatalf("item creado inesperado: %+v", item)
	}

	// 2. Activo returns true
	activo, err := repo.Activo(ctx, "tipo_actividad", "nuevo_riego")
	if err != nil {
		t.Fatalf("error verificando activo: %v", err)
	}
	if !activo {
		t.Fatal("se esperaba que el item estuviera activo")
	}

	// 3. Deactivate sets activo = false
	if err := repo.Deactivate(ctx, item.ID); err != nil {
		t.Fatalf("error desactivando item: %v", err)
	}

	activo, err = repo.Activo(ctx, "tipo_actividad", "nuevo_riego")
	if err != nil {
		t.Fatalf("error verificando activo post-desactivar: %v", err)
	}
	if activo {
		t.Fatal("se esperaba que el item no estuviera activo")
	}

	// 4. Create on conflict updates name and reactivates (activo = true)
	reactivado, err := repo.Create(ctx, "tipo_actividad", "nuevo_riego", "Riego Actualizado")
	if err != nil {
		t.Fatalf("error reactivando item: %v", err)
	}
	if reactivado.ID != item.ID {
		t.Fatalf("se esperaba mantener id %d, se obtuvo %d", item.ID, reactivado.ID)
	}
	if reactivado.Nombre != "Riego Actualizado" || !reactivado.Activo {
		t.Fatalf("item reactivado inesperado: %+v", reactivado)
	}

	// 5. Deactivate on non-existent id returns ErrItemNoExiste
	err = repo.Deactivate(ctx, 99999999)
	if !errors.Is(err, apperrors.ErrItemNoExiste) {
		t.Fatalf("se esperaba ErrItemNoExiste, se obtuvo %v", err)
	}

	// 6. List with filter
	list, err := repo.List(ctx, "tipo_actividad", false)
	if err != nil {
		t.Fatalf("error listando: %v", err)
	}
	found := false
	for _, it := range list {
		if it.Codigo == "nuevo_riego" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("item creado no fue encontrado en la lista")
	}

	// 7. Verify no physical deletes occur: count total rows before and after deactivation
	var countBefore int64
	if err := gdb.Raw("SELECT count(*) FROM catalogos").Scan(&countBefore).Error; err != nil {
		t.Fatal(err)
	}
	if err := repo.Deactivate(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	var countAfter int64
	if err := gdb.Raw("SELECT count(*) FROM catalogos").Scan(&countAfter).Error; err != nil {
		t.Fatal(err)
	}
	if countBefore != countAfter {
		t.Fatalf("el conteo total de filas cambió tras desactivar: antes=%d, despues=%d", countBefore, countAfter)
	}
}
