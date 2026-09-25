package migrate

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"campusverde/api/internal/db"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestMigraciones001a008EnVacioYEnActual(t *testing.T) {
	base := os.Getenv("MIGRATE_TEST_URL")
	if base == "" {
		base = "postgres://campus:campus@127.0.0.1:5432/postgres?sslmode=disable"
	}
	admin, err := sql.Open("pgx", base)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		t.Skipf("sin postgres de prueba: %v", err)
	}

	dir := filepath.Join("..", "..", "migrations")
	vacia := "campus_verde_ola0_vacia"
	actual := "campus_verde_ola0_actual"
	recrear(t, admin, vacia)
	recrear(t, admin, actual)
	defer func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname IN ($1, $2) AND pid <> pg_backend_pid()`, vacia, actual)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + vacia)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + actual)
	}()

	if err := aplicar(vacia, dir); err != nil {
		t.Fatalf("base vacía: %v", err)
	}
	assertEsquema(t, vacia)

	previas := filepath.Join(t.TempDir(), "previas")
	if err := os.MkdirAll(previas, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() || e.Name() >= "007" {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(previas, e.Name()), body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := aplicar(actual, previas); err != nil {
		t.Fatalf("001-006 sobre base vacía: %v", err)
	}
	if err := aplicar(actual, dir); err != nil {
		t.Fatalf("007-008 sobre la base actual: %v", err)
	}
	assertEsquema(t, actual)

	gdb, err := db.Open(urlDe(actual))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		d, _ := gdb.DB()
		d.Close()
	}()
	var n int
	if err := gdb.Raw(`SELECT count(*) FROM actividades WHERE titulo = 'Poda contratada del borde sur'`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("la semilla de 006 no sobrevivió: %d", n)
	}
	var aplicadas int
	if err := gdb.Raw(`SELECT count(*) FROM schema_migrations`).Scan(&aplicadas).Error; err != nil {
		t.Fatal(err)
	}
	if aplicadas != 8 {
		t.Fatalf("migraciones aplicadas = %d", aplicadas)
	}
}

func recrear(t *testing.T, admin *sql.DB, name string) {
	t.Helper()
	if _, err := admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("DROP DATABASE IF EXISTS " + name); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
}

func aplicar(name, dir string) error {
	gdb, err := db.Open(urlDe(name))
	if err != nil {
		return err
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return Apply(gdb, dir)
}

func urlDe(name string) string {
	return fmt.Sprintf("postgres://campus:campus@127.0.0.1:5432/%s?sslmode=disable", name)
}

func assertEsquema(t *testing.T, name string) {
	t.Helper()
	gdb, err := db.Open(urlDe(name))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	var cambios int
	if err := gdb.Raw(`SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'cambios'`).Scan(&cambios).Error; err != nil {
		t.Fatal(err)
	}
	if cambios != 1 {
		t.Fatalf("%s: tabla cambios = %d", name, cambios)
	}
	var fk int
	if err := gdb.Raw(`SELECT count(*) FROM pg_constraint WHERE conname = 'permisos_rol_fkey'`).Scan(&fk).Error; err != nil {
		t.Fatal(err)
	}
	if fk != 1 {
		t.Fatalf("%s: falta permisos_rol_fkey", name)
	}
	var roles int
	if err := gdb.Raw(`SELECT count(*) FROM roles`).Scan(&roles).Error; err != nil {
		t.Fatal(err)
	}
	if roles != 4 {
		t.Fatalf("%s: roles semilla = %d", name, roles)
	}
	if err := gdb.Exec(`INSERT INTO permisos (rol, accion) VALUES ('rol_inexistente', 'consultar')`).Error; err == nil {
		t.Fatalf("%s: la FK aceptó un rol fuera del catálogo", name)
	}
}
