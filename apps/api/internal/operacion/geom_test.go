package operacion

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"campusverde/api/internal/db"
	"campusverde/api/internal/migrate"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type filaNula struct{}

func (filaNula) Scan(dest ...any) error {
	*dest[0].(*string) = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2"
	*dest[1].(*string) = "riego"
	*dest[2].(*string) = "pendiente"
	*dest[3].(*string) = "Con lugar"
	*dest[4].(*string) = ""
	*dest[9].(*bool) = false
	*dest[10].(*time.Time) = time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	*dest[11].(*time.Time) = time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	*dest[12].(*string) = "propia"
	*dest[13].(*sql.NullString) = sql.NullString{}
	return nil
}

func TestScanFeatureAceptaGeomNula(t *testing.T) {
	f, err := scanFeature(filaNula{})
	if err != nil {
		t.Fatal(err)
	}
	if string(f.Geometry) != "null" {
		t.Fatalf("geometry = %s", f.Geometry)
	}
	body, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(body) {
		t.Fatalf("json inválido: %s", body)
	}
}

func TestListLaborSinGeom(t *testing.T) {
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
	name := "campus_verde_labor_geom"
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
	if err := gdb.Exec(`
		INSERT INTO lugares (nombre, nombre_norm, lat, lon) VALUES ('Eje', 'eje', -12.07, -77.08);
		INSERT INTO actividades (id, tipo, estado, titulo, detalle, geom, lugar_id)
		VALUES (
		  'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', 'riego', 'pendiente', 'Con lugar', '', NULL,
		  (SELECT id FROM lugares WHERE nombre_norm = 'eje')
		)`).Error; err != nil {
		t.Fatal(err)
	}
	fc, err := NewStore(gdb).List(context.Background(), Query{Rol: RolCoordinacion})
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, f := range fc.Features {
		if f.ID != "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2" {
			continue
		}
		found = true
		if !json.Valid(f.Geometry) || string(f.Geometry) == "" {
			t.Fatalf("geometría = %s", f.Geometry)
		}
	}
	if !found {
		t.Fatal("la labor sin geom no salió en el listado")
	}
}
