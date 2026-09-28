package accesos_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"campusverde/api/internal/accesos"
	"campusverde/api/internal/db"
	"campusverde/api/internal/migrate"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func migrarDBTemporal(t *testing.T, name string) *sql.DB {
	t.Helper()
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

	gdb, err := db.Open(fmt.Sprintf("postgres://campus:campus@127.0.0.1:5432/%s?sslmode=disable", name))
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate.Apply(gdb, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	return sqlDB
}

func TestEnsureConservaFilasExtraYEsIdempotente(t *testing.T) {
	name := "campus_verde_test_accesos_ensure"
	sqlDB := migrarDBTemporal(t, name)
	defer sqlDB.Close()

	gdb, err := db.Open(fmt.Sprintf("postgres://campus:campus@127.0.0.1:5432/%s?sslmode=disable", name))
	if err != nil {
		t.Fatal(err)
	}

	// 1. Insertamos un permiso extra
	if err := gdb.Exec(`INSERT INTO permisos (rol, accion) VALUES ('admin', 'accion_extra') ON CONFLICT DO NOTHING`).Error; err != nil {
		t.Fatal(err)
	}

	// 2. Ejecutamos Ensure
	if err := accesos.Ensure(gdb, "pando-local"); err != nil {
		t.Fatalf("primer Ensure falló: %v", err)
	}

	// Verificamos que el permiso extra sobrevive
	var n int
	if err := gdb.Raw(`SELECT count(*) FROM permisos WHERE rol = 'admin' AND accion = 'accion_extra'`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("el permiso extra no sobrevivió a Ensure: n=%d", n)
	}

	// Verificamos que jefatura tiene 'evidencias'
	var nJef int
	if err := gdb.Raw(`SELECT count(*) FROM permisos WHERE rol = 'jefatura' AND accion = 'evidencias'`).Scan(&nJef).Error; err != nil {
		t.Fatal(err)
	}
	if nJef != 1 {
		t.Fatalf("jefatura no tiene permiso de evidencias: n=%d", nJef)
	}

	var count1 int
	if err := gdb.Raw(`SELECT count(*) FROM permisos`).Scan(&count1).Error; err != nil {
		t.Fatal(err)
	}

	// 3. Ejecutamos Ensure por segunda vez (idempotencia)
	if err := accesos.Ensure(gdb, "pando-local"); err != nil {
		t.Fatalf("segundo Ensure falló: %v", err)
	}

	var count2 int
	if err := gdb.Raw(`SELECT count(*) FROM permisos`).Scan(&count2).Error; err != nil {
		t.Fatal(err)
	}
	if count1 != count2 {
		t.Fatalf("Ensure no fue idempotente: count1=%d, count2=%d", count1, count2)
	}

	// 4. Verificamos que Login y FromToken devuelven rol_nombre
	store := accesos.NewStore(gdb)
	ctx := context.Background()

	token, u, err := store.Login(ctx, "jefatura", "pando-local")
	if err != nil {
		t.Fatalf("Login falló: %v", err)
	}
	if u.RolNombre != "Jefatura / Jefe de sección" {
		t.Fatalf("Login rol_nombre esperado 'Jefatura / Jefe de sección', obtuve %q", u.RolNombre)
	}

	uTok, err := store.FromToken(ctx, token)
	if err != nil {
		t.Fatalf("FromToken falló: %v", err)
	}
	if uTok.RolNombre != "Jefatura / Jefe de sección" {
		t.Fatalf("FromToken rol_nombre esperado 'Jefatura / Jefe de sección', obtuve %q", uTok.RolNombre)
	}

	// Verificamos Usuarios()
	usuarios, err := store.Usuarios(ctx)
	if err != nil {
		t.Fatalf("Usuarios falló: %v", err)
	}
	encontrado := false
	for _, usr := range usuarios {
		if usr.Usuario == "coordinacion" {
			encontrado = true
			if usr.RolNombre != "Ingeniería/Coordinación" {
				t.Fatalf("Usuarios rol_nombre esperado 'Ingeniería/Coordinación', obtuve %q", usr.RolNombre)
			}
		}
	}
	if !encontrado {
		t.Fatal("usuario coordinacion no encontrado en Usuarios()")
	}
}
