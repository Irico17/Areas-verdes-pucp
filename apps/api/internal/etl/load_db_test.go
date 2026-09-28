package etl

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	if err := gdb.Exec(`TRUNCATE actividades, actividad_eventos, ordenes_servicio, riego_registros, cambios CASCADE`).Error; err != nil {
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

	if err := gdb.Exec(`TRUNCATE actividades, actividad_eventos, ordenes_servicio, riego_registros, cambios CASCADE`).Error; err != nil {
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

	// Segunda corrida: debe negarse porque areas_verdes ya contiene filas.
	if err := Load(gdb, areas, nil, map[string][]Record{}); err == nil {
		t.Fatal("segunda corrida: Load debía fallar porque areas_verdes ya contiene filas")
	}
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
		t.Fatalf("segunda corrida: cambios registrados = %d, se esperaba seguir en 1", totalCambios)
	}
	if err := gdb.Raw(`SELECT codigo FROM areas_verdes WHERE feature_id = 'AV-0259'`).Scan(&codigoMayor).Error; err != nil {
		t.Fatal(err)
	}
	if codigoMayor != esperado {
		t.Fatalf("segunda corrida: AV-0259 codigo = %q, se esperaba seguir en %q", codigoMayor, esperado)
	}
}

// TestLoadAplicaSectoresEnCatalogoCompleto carga las 521 áreas y 534 zonas
// reales (sin PII: NormalizeZonas descarta "jefes") y comprueba que Load deja
// el sector completo con los conteos de data/v1/zonas_sector.json, y que un
// segundo Load se niega por existir datos en poligonos_cuadrilla/areas_verdes.
func TestLoadAplicaSectoresEnCatalogoCompleto(t *testing.T) {
	gdb := migrarDBTemporal(t, "campus_verde_etl_sector_c")

	if err := gdb.Exec(`TRUNCATE actividades, actividad_eventos, ordenes_servicio, riego_registros, cambios CASCADE`).Error; err != nil {
		t.Fatal(err)
	}

	dir := rawDir(t)

	areasBody, err := os.ReadFile(filepath.Join(dir, "areas_verdes.geojson"))
	if err != nil {
		t.Fatal(err)
	}
	areas, err := NormalizeAreas(areasBody)
	if err != nil {
		t.Fatal(err)
	}
	zonasBody, err := os.ReadFile(filepath.Join(dir, "jefe_de_grupo.json"))
	if err != nil {
		t.Fatal(err)
	}
	zonas, err := NormalizeZonas(zonasBody)
	if err != nil {
		t.Fatal(err)
	}

	esperado := map[string]int{"cua-valeria": 259, "cua-mateo": 168, "cua-renato": 104, "campo-deportivo": 2, "bosque-humedo": 1}
	verificarConteos := func(momento string) {
		t.Helper()
		var sinSector int
		if err := gdb.Raw(`SELECT count(*) FROM poligonos_cuadrilla WHERE sector IS NULL`).Scan(&sinSector).Error; err != nil {
			t.Fatal(err)
		}
		if sinSector != 0 {
			t.Fatalf("%s: %d polígono(s) sin sector", momento, sinSector)
		}
		for sector, n := range esperado {
			var got int
			if err := gdb.Raw(`SELECT count(*) FROM poligonos_cuadrilla WHERE sector = $1`, sector).Scan(&got).Error; err != nil {
				t.Fatal(err)
			}
			if got != n {
				t.Fatalf("%s: sector=%s conteo=%d, se esperaba %d", momento, sector, got, n)
			}
		}
	}

	if err := Load(gdb, areas, zonas, map[string][]Record{}); err != nil {
		t.Fatal(err)
	}
	verificarConteos("primer Load")

	if err := Load(gdb, areas, zonas, map[string][]Record{}); err == nil {
		t.Fatal("segundo Load debía ser rechazado por existir datos en catastro")
	}
	verificarConteos("después de segundo Load rechazado")
}

// TestLoadSeNiegaSiHayEjemplares comprueba que Load no vuelva a hacer
// TRUNCATE ... CASCADE sobre una base con inventario real: ejemplares no tiene
// filtro por FK en un TRUNCATE CASCADE, así que una recarga borraría todos los
// ejemplares (y sus hijas) junto con el catastro. Load debe negarse con un
// error claro en vez de borrar esas filas.
func TestLoadSeNiegaSiHayEjemplares(t *testing.T) {
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
	name := "campus_verde_etl_1c"
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

	if err := gdb.Exec(`INSERT INTO ejemplares DEFAULT VALUES`).Error; err != nil {
		t.Fatal(err)
	}

	geom := json.RawMessage(`{"type":"MultiPolygon","coordinates":[[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.069],[-77.08,-12.07]]]]}`)
	nombre := "Polígono de carga"
	area := Record{FeatureID: "AV-9001", SourceIndex: 1, Nombre: &nombre, Geometry: geom}
	if err := Load(gdb, []Record{area}, nil, map[string][]Record{}); err == nil {
		t.Fatal("Load no debía escribir con ejemplares existentes")
	}

	var nEjemplares, nAreas int
	if err := gdb.Raw(`SELECT count(*) FROM ejemplares`).Scan(&nEjemplares).Error; err != nil {
		t.Fatal(err)
	}
	if nEjemplares != 1 {
		t.Fatalf("ejemplares = %d, se esperaba conservar la fila existente", nEjemplares)
	}
	if err := gdb.Raw(`SELECT count(*) FROM areas_verdes`).Scan(&nAreas).Error; err != nil {
		t.Fatal(err)
	}
	if nAreas != 0 {
		t.Fatalf("areas_verdes = %d, Load no debía haber escrito nada", nAreas)
	}
}

func TestLoadSeNiegaSiHayActividades(t *testing.T) {
	gdb := migrarDBTemporal(t, "campus_verde_etl_guard_act")

	// La base recién migrada ya contiene actividades de demostración (003/006).
	geom := json.RawMessage(`{"type":"MultiPolygon","coordinates":[[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.069],[-77.08,-12.07]]]]}`)
	nombre := "Polígono de carga"
	area := Record{FeatureID: "AV-9001", SourceIndex: 1, Nombre: &nombre, Geometry: geom}

	err := Load(gdb, []Record{area}, nil, map[string][]Record{})
	if err == nil {
		t.Fatal("Load debía negarse si hay actividades")
	}
	if !strings.Contains(err.Error(), "actividades tiene") {
		t.Fatalf("error inesperado: %v", err)
	}
}

func TestLoadSeNiegaSiHayAreasVerdes(t *testing.T) {
	gdb := migrarDBTemporal(t, "campus_verde_etl_guard_av")

	if err := gdb.Exec(`TRUNCATE actividades, actividad_eventos, ordenes_servicio, riego_registros CASCADE`).Error; err != nil {
		t.Fatal(err)
	}

	if err := gdb.Exec(`INSERT INTO areas_verdes (feature_id, source_index, geom) VALUES ('AV-9999', 1, ST_SetSRID(ST_Multi(ST_GeomFromGeoJSON('{"type":"Polygon","coordinates":[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.069],[-77.08,-12.07]]]}')), 4326))`).Error; err != nil {
		t.Fatal(err)
	}

	geom := json.RawMessage(`{"type":"MultiPolygon","coordinates":[[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.069],[-77.08,-12.07]]]]}`)
	nombre := "Polígono de carga"
	area := Record{FeatureID: "AV-9001", SourceIndex: 1, Nombre: &nombre, Geometry: geom}

	err := Load(gdb, []Record{area}, nil, map[string][]Record{})
	if err == nil {
		t.Fatal("Load debía negarse si hay areas_verdes")
	}
	if !strings.Contains(err.Error(), "areas_verdes tiene") {
		t.Fatalf("error inesperado: %v", err)
	}
}

func TestLoadSeNiegaSiHayCambios(t *testing.T) {
	gdb := migrarDBTemporal(t, "campus_verde_etl_guard_cambios")

	if err := gdb.Exec(`TRUNCATE actividades, actividad_eventos, ordenes_servicio, riego_registros CASCADE`).Error; err != nil {
		t.Fatal(err)
	}

	if err := gdb.Exec(`INSERT INTO cambios (entidad, entidad_id, accion, despues) VALUES ('prueba', '1', 'alta', '{}'::jsonb)`).Error; err != nil {
		t.Fatal(err)
	}

	geom := json.RawMessage(`{"type":"MultiPolygon","coordinates":[[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.069],[-77.08,-12.07]]]]}`)
	nombre := "Polígono de carga"
	area := Record{FeatureID: "AV-9001", SourceIndex: 1, Nombre: &nombre, Geometry: geom}

	err := Load(gdb, []Record{area}, nil, map[string][]Record{})
	if err == nil {
		t.Fatal("Load debía negarse si hay cambios")
	}
	if !strings.Contains(err.Error(), "cambios tiene") {
		t.Fatalf("error inesperado: %v", err)
	}
}

func TestLoadInventarioSeNiegaSiHayInventario(t *testing.T) {
	gdb := migrarDBTemporal(t, "campus_verde_etl_guard_inv")

	if err := gdb.Exec(`INSERT INTO inventario (capa, feature_id) VALUES ('arboles', 'ARB-001')`).Error; err != nil {
		t.Fatal(err)
	}

	err := LoadInventario(gdb, map[string][]InvRecord{
		"arboles": {
			{FeatureID: "ARB-002", Capa: "arboles"},
		},
	})
	if err == nil {
		t.Fatal("LoadInventario debía negarse si inventario ya contiene filas")
	}
	if !strings.Contains(err.Error(), "inventario tiene") {
		t.Fatalf("error inesperado: %v", err)
	}
}
