package etl

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"campusverde/api/internal/auditoria"
	"campusverde/api/internal/db"
	"campusverde/api/internal/migrate"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestConfirmarEscribeYRevertirDeshace(t *testing.T) {
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
	name := "campus_verde_import_3a"
	if _, err := admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name); err != nil {
		t.Fatal(err)
	}
	_, _ = admin.Exec("DROP DATABASE IF EXISTS " + name)
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + name)
	})
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
	var usuario int64
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, password_hash)
		VALUES ('import.coord', 'Coordinación', 'coordinacion', 'no-es-clave')
		RETURNING id`).Row().Scan(&usuario); err != nil {
		t.Fatal(err)
	}

	invalido := []byte("lugar,latitud,longitud\nAfuera,-1.20,-7.70\n")
	vista, err := Previsualizar("lugares", "lugares.csv", invalido)
	if err != nil {
		t.Fatal(err)
	}
	if vista.Validas != 0 {
		t.Fatalf("el CSV inválido no debía traer filas: %+v", vista)
	}
	var idMalo int64
	if err := gdb.Raw(`
		INSERT INTO lotes_importacion (entidad, estado, usuario_id, filas, contenido, nombre_archivo)
		VALUES ('lugares', 'vista_previa', $1, 0, $2, 'lugares.csv')
		RETURNING id`, usuario, invalido).Row().Scan(&idMalo); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Confirmar(gdb, idMalo, usuario); err == nil {
		t.Fatal("confirmar un CSV inválido no debía escribir")
	}
	var n int
	if err := gdb.Raw(`SELECT count(*) FROM lugares`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("lugares = %d", n)
	}

	valido := []byte("lugar,latitud,longitud\nBiblioteca,-12.0704,-77.0808\n")
	var id int64
	if err := gdb.Raw(`
		INSERT INTO lotes_importacion (entidad, estado, usuario_id, filas, contenido, nombre_archivo)
		VALUES ('lugares', 'vista_previa', $1, 1, $2, 'lugares.csv')
		RETURNING id`, usuario, valido).Row().Scan(&id); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Confirmar(gdb, id, usuario); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Raw(`SELECT count(*) FROM lugares WHERE activo`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("tras confirmar, lugares activos = %d", n)
	}
	rep, err := auditoria.NewStore(gdb).Revertir(context.Background(), id, usuario, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Revertidas) != 1 {
		t.Fatalf("revertidas %+v", rep)
	}
	if err := gdb.Raw(`SELECT count(*) FROM lugares WHERE activo`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("tras revertir, lugares activos = %d", n)
	}
}
