package catastro

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"campusverde/api/internal/db"
	"campusverde/api/internal/migrate"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestRecodificarConCodigoPrevioVacio(t *testing.T) {
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
	name := "campus_verde_recodifica"
	if _, err := admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name); err != nil {
		t.Fatal(err)
	}
	_, _ = admin.Exec("DROP DATABASE IF EXISTS " + name)
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + name)
	}()

	gdb, err := db.Open(fmt.Sprintf("postgres://campus:campus@127.0.0.1:5432/%s?sslmode=disable", name))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	if err := migrate.Apply(gdb, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	store := NewStore(gdb)
	ej, err := store.CrearEjemplar(context.Background(), Ejemplar{Cantidad: 1, NombreComun: "Tipuana"})
	if err != nil {
		t.Fatal(err)
	}
	hist, err := store.Recodificar(context.Background(), ej.ID, "AV-100")
	if err != nil {
		t.Fatal(err)
	}
	if hist.CodigoAnterior != "" || hist.CodigoNuevo != "AV-100" || hist.ID == 0 {
		t.Fatalf("historial = %+v", hist)
	}
	var codigo string
	if err := gdb.Raw(`SELECT codigo FROM ejemplares WHERE id = ?`, ej.ID).Scan(&codigo).Error; err != nil {
		t.Fatal(err)
	}
	if codigo != "AV-100" {
		t.Fatalf("codigo = %s", codigo)
	}
}
