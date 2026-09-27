package etl

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"campusverde/api/internal/db"
	"campusverde/api/internal/migrate"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestLoadSobreBaseMigrada(t *testing.T) {
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
	name := "campus_verde_etl_1a"
	if _, err := admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("DROP DATABASE IF EXISTS " + name); err != nil {
		t.Fatal(err)
	}
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
	dir := filepath.Join("..", "..", "migrations")
	if err := migrate.Apply(gdb, dir); err != nil {
		t.Fatal(err)
	}

	geom := json.RawMessage(`{"type":"MultiPolygon","coordinates":[[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.069],[-77.08,-12.07]]]]}`)
	nombre := "Polígono de carga"
	area := Record{FeatureID: "AV-9001", SourceIndex: 1, Nombre: &nombre, Geometry: geom}
	zona := Record{FeatureID: "Z-9002", SourceIndex: 1, Nombre: &nombre, Geometry: geom}
	if err := Load(gdb, []Record{area}, []Record{zona}, map[string][]Record{}); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := gdb.Raw(`SELECT count(*) FROM poligonos_cuadrilla WHERE feature_id = 'Z-9002'`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("poligonos_cuadrilla = %d", n)
	}
	if err := gdb.Raw(`SELECT count(*) FROM zonas WHERE feature_id = 'Z-9002'`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("vista zonas = %d", n)
	}
}

// TestLoadRenombraCodigoDuplicado reproduce el catastro real, que repite código
// (p. ej. "D 20" en AV-0255 y AV-0259): Load no debe fallar por
// areas_verdes_codigo_uidx, sino renombrar el duplicado y dejarlo en cambios,
// igual que las migraciones 014/034. Una segunda corrida no debe crear más
// renombres ni duplicar el registro en cambios.
func TestLoadRenombraCodigoDuplicado(t *testing.T) {
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
	name := "campus_verde_etl_1b"
	if _, err := admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("DROP DATABASE IF EXISTS " + name); err != nil {
		t.Fatal(err)
	}
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
	dir := filepath.Join("..", "..", "migrations")
	if err := migrate.Apply(gdb, dir); err != nil {
		t.Fatal(err)
	}

	geom := json.RawMessage(`{"type":"MultiPolygon","coordinates":[[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.069],[-77.08,-12.07]]]]}`)
	codigo := "D 20"
	otro := "G 13"
	nombreA := "Área menor id"
	nombreB := "Área duplicada"
	areas := []Record{
		{FeatureID: "AV-0255", SourceIndex: 254, Codigo: &codigo, Nombre: &nombreA, Geometry: geom},
		{FeatureID: "AV-0256", SourceIndex: 255, Codigo: &otro, Nombre: &nombreA, Geometry: geom},
		{FeatureID: "AV-0259", SourceIndex: 258, Codigo: &codigo, Nombre: &nombreB, Geometry: geom},
	}

	cargar := func() {
		t.Helper()
		if err := Load(gdb, areas, nil, map[string][]Record{}); err != nil {
			t.Fatal(err)
		}
	}
	cargar()

	var codigoMenor, codigoMayor string
	if err := gdb.Raw(`SELECT codigo FROM areas_verdes WHERE feature_id = 'AV-0255'`).Scan(&codigoMenor).Error; err != nil {
		t.Fatal(err)
	}
	if codigoMenor != codigo {
		t.Fatalf("AV-0255 (menor id) codigo = %q, se esperaba sin cambios %q", codigoMenor, codigo)
	}
	var idMayor int64
	if err := gdb.Raw(`SELECT id, codigo FROM areas_verdes WHERE feature_id = 'AV-0259'`).Row().Scan(&idMayor, &codigoMayor); err != nil {
		t.Fatal(err)
	}
	esperado := fmt.Sprintf("%s ·%d", codigo, idMayor)
	if codigoMayor != esperado {
		t.Fatalf("AV-0259 (mayor id) codigo = %q, se esperaba %q", codigoMayor, esperado)
	}

	var totalAreas, totalCambios int
	if err := gdb.Raw(`SELECT count(*) FROM areas_verdes`).Scan(&totalAreas).Error; err != nil {
		t.Fatal(err)
	}
	if totalAreas != len(areas) {
		t.Fatalf("areas_verdes = %d, se esperaban %d (ninguna fila borrada)", totalAreas, len(areas))
	}
	if err := gdb.Raw(`SELECT count(*) FROM cambios WHERE entidad = 'areas_verdes'`).Scan(&totalCambios).Error; err != nil {
		t.Fatal(err)
	}
	if totalCambios != 1 {
		t.Fatalf("cambios registrados = %d, se esperaba 1", totalCambios)
	}

	// Segunda corrida: mismo resultado, sin renombres nuevos ni cambios duplicados.
	cargar()
	if err := gdb.Raw(`SELECT count(*) FROM areas_verdes`).Scan(&totalAreas).Error; err != nil {
		t.Fatal(err)
	}
	if totalAreas != len(areas) {
		t.Fatalf("segunda corrida: areas_verdes = %d, se esperaban %d", totalAreas, len(areas))
	}
	if err := gdb.Raw(`SELECT count(*) FROM cambios WHERE entidad = 'areas_verdes'`).Scan(&totalCambios).Error; err != nil {
		t.Fatal(err)
	}
	if totalCambios != 1 {
		t.Fatalf("segunda corrida: cambios registrados = %d, se esperaba seguir en 1 (idempotente)", totalCambios)
	}
	if err := gdb.Raw(`SELECT codigo FROM areas_verdes WHERE feature_id = 'AV-0259'`).Scan(&codigoMayor).Error; err != nil {
		t.Fatal(err)
	}
	if codigoMayor != esperado {
		t.Fatalf("segunda corrida: AV-0259 codigo = %q, se esperaba seguir en %q", codigoMayor, esperado)
	}
}
