package database

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var identSQL = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

func openTestDB(name string) (*gorm.DB, error) {
	u := urlDe(name)
	if u == "" {
		return nil, fmt.Errorf("MIGRATE_TEST_URL no configurada o inválida")
	}
	return NewDatabaseConnection(&config.Config{
		Database: config.DatabaseConfig{
			URL: u,
		},
	})
}

func urlDe(name string) string {
	base := os.Getenv("MIGRATE_TEST_URL")
	if base == "" {
		return ""
	}
	u, err := url.Parse(base)
	if err != nil {
		return ""
	}
	u.Path = "/" + name
	return u.String()
}

func dirSerieHistorica() string {
	live := findMigrationsDir()
	hist := filepath.Clean(filepath.Join(live, "..", "referencia", "migraciones-historicas"))
	if _, err := os.Stat(filepath.Join(hist, "014_areas_verdes_zona.sql")); err == nil {
		return hist
	}
	return live
}

func findMigrationsDir() string {
	if env := os.Getenv("MIGRATIONS_DIR"); env != "" {
		if fi, err := os.Stat(env); err == nil && fi.IsDir() {
			return env
		}
	}
	wd, err := os.Getwd()
	if err == nil {
		for dir := wd; ; {
			candidate := filepath.Join(dir, "db", "migrations")
			if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() || (err == nil && fi.IsDir()) {
				if fi.IsDir() {
					return candidate
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return filepath.Join("..", "..", "..", "..", "..", "db", "migrations")
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
	gdb, err := openTestDB(name)
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

func sembrarZona(name string) error {
	gdb, err := openTestDB(name)
	if err != nil {
		return err
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return gdb.Exec(`
		INSERT INTO zonas (feature_id, source_index, nombre, geom)
		VALUES (
		  'Z-9001', 9001, 'Polígono de prueba',
		  ST_SetSRID(ST_GeomFromText('MULTIPOLYGON(((-77.08 -12.07, -77.079 -12.07, -77.079 -12.069, -77.08 -12.069, -77.08 -12.07)))'), 4326)
		)`).Error
}

func assertCopiaZona(t *testing.T, name string) {
	t.Helper()
	gdb, err := openTestDB(name)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	var n int
	if err := gdb.Raw(`SELECT count(*) FROM poligonos_cuadrilla WHERE feature_id = 'Z-9001'`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("%s: la copia del polígono no está en poligonos_cuadrilla", name)
	}
	if err := gdb.Raw(`SELECT count(*) FROM zonas WHERE feature_id = 'Z-9001'`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("%s: la vista zonas no devuelve el polígono", name)
	}
}

func assertEsquema(t *testing.T, name string) {
	t.Helper()
	gdb, err := openTestDB(name)
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
	var fkUsuarios int
	if err := gdb.Raw(`SELECT count(*) FROM pg_constraint WHERE conname = 'usuarios_rol_fkey'`).Scan(&fkUsuarios).Error; err != nil {
		t.Fatal(err)
	}
	if fkUsuarios != 1 {
		t.Fatalf("%s: falta usuarios_rol_fkey", name)
	}
	var chkUsuarios int
	if err := gdb.Raw(`SELECT count(*) FROM pg_constraint WHERE conname = 'usuarios_rol_chk'`).Scan(&chkUsuarios).Error; err != nil {
		t.Fatal(err)
	}
	if chkUsuarios != 0 {
		t.Fatalf("%s: usuarios_rol_chk no debe existir tras migración 046", name)
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
	var relkind string
	if err := gdb.Raw(`SELECT relkind FROM pg_class WHERE relname = 'zonas'`).Scan(&relkind).Error; err != nil {
		t.Fatal(err)
	}
	if relkind != "v" {
		t.Fatalf("%s: zonas debe ser una vista, relkind=%s", name, relkind)
	}
	if err := gdb.Raw(`SELECT relkind FROM pg_class WHERE relname = 'zonas_origen'`).Scan(&relkind).Error; err != nil {
		t.Fatal(err)
	}
	if relkind != "r" {
		t.Fatalf("%s: zonas_origen debe seguir siendo tabla, relkind=%s", name, relkind)
	}
	for _, tabla := range []string{
		"zonas_supervision", "cuadrillas", "poligonos_cuadrilla", "asignaciones_poligono",
		"lugares", "especies", "ejemplares", "codigos_historicos", "medidas_palmera",
		"fauna", "puertas", "playas_estacionamiento", "veredas_riesgo", "xerofiticas", "jardines_reserva",
		"poligonos_sector_ref",
	} {
		var n int
		if err := gdb.Raw(`SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?`, tabla).Scan(&n).Error; err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Fatalf("%s: falta la tabla %s", name, tabla)
		}
	}
	var cuadrillas int
	if err := gdb.Raw(`SELECT count(*) FROM cuadrillas`).Scan(&cuadrillas).Error; err != nil {
		t.Fatal(err)
	}
	if cuadrillas != 9 {
		t.Fatalf("%s: cuadrillas de demostración = %d", name, cuadrillas)
	}
	var geomNull string
	if err := gdb.Raw(`SELECT attnotnull::text FROM pg_attribute WHERE attrelid = 'poligonos_cuadrilla'::regclass AND attname = 'geom'`).Scan(&geomNull).Error; err != nil {
		t.Fatal(err)
	}
	if geomNull != "true" {
		t.Fatalf("%s: poligonos_cuadrilla.geom debería ser NOT NULL cuando la copia no tiene nulos", name)
	}
	var sectorCols int
	if err := gdb.Raw(`SELECT count(*) FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'poligonos_cuadrilla' AND column_name = 'sector'`).Scan(&sectorCols).Error; err != nil {
		t.Fatal(err)
	}
	if sectorCols != 1 {
		t.Fatalf("%s: falta poligonos_cuadrilla.sector", name)
	}
	var sectorEnVista int
	if err := gdb.Raw(`SELECT count(*) FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'zonas' AND column_name = 'sector'`).Scan(&sectorEnVista).Error; err != nil {
		t.Fatal(err)
	}
	if sectorEnVista != 1 {
		t.Fatalf("%s: la vista zonas no expone sector", name)
	}
	if err := gdb.Exec(`
		INSERT INTO zonas (feature_id, source_index, nombre, geom)
		VALUES (
		  'Z-9101', 9101, 'Sin sector',
		  ST_SetSRID(ST_GeomFromText('MULTIPOLYGON(((-77.08 -12.07, -77.079 -12.07, -77.079 -12.069, -77.08 -12.069, -77.08 -12.07)))'), 4326)
		)`).Error; err != nil {
		t.Fatalf("%s: INSERT en zonas sin sector debería seguir funcionando: %v", name, err)
	}
	if err := gdb.Exec(`UPDATE poligonos_cuadrilla SET sector = 'otro' WHERE feature_id = 'Z-9101'`).Error; err != nil {
		t.Fatalf("%s: 060 retiró el CHECK, un sector nuevo debería guardarse: %v", name, err)
	}
	if err := gdb.Exec(`UPDATE poligonos_cuadrilla SET sector = 'cua-valeria' WHERE feature_id = 'Z-9101'`).Error; err != nil {
		t.Fatalf("%s: el CHECK debería aceptar un sector válido: %v", name, err)
	}
	if err := gdb.Exec(`
		INSERT INTO actividades (id, tipo, estado, titulo, detalle, geom)
		VALUES ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', 'riego', 'pendiente', 'Sin ubicación', '', NULL)
	`).Error; err == nil {
		t.Fatalf("%s: actividades aceptó geom NULL sin lugar ni zona", name)
	}
	if err := gdb.Exec(`
		INSERT INTO lugares (nombre, nombre_norm, lat, lon)
		VALUES ('Eje de prueba', 'eje de prueba', -12.07, -77.08)
	`).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Exec(`
		INSERT INTO actividades (id, tipo, estado, titulo, detalle, geom, lugar_id)
		VALUES (
		  'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', 'riego', 'pendiente', 'Con lugar', '', NULL,
		  (SELECT id FROM lugares WHERE nombre_norm = 'eje de prueba')
		)
	`).Error; err != nil {
		t.Fatalf("%s: actividades no aceptó geom NULL con lugar: %v", name, err)
	}
}

func copiarHasta(src, dst, corte string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") || e.Name() >= corte {
			continue
		}
		body, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), body, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func sembrarDuplicados(name string) error {
	gdb, err := openTestDB(name)
	if err != nil {
		return err
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	geom := `ST_SetSRID(ST_GeomFromText('MULTIPOLYGON(((-77.08 -12.07, -77.079 -12.07, -77.079 -12.069, -77.08 -12.069, -77.08 -12.07)))'), 4326)`
	if err := gdb.Exec(`
		INSERT INTO areas_verdes (feature_id, source_index, codigo, nombre, perimetro_m, area_m2, referencia)
		VALUES
		  ('AV-D1', 9101, 'F 26', 'primero', -3, -4, repeat('x', 520)),
		  ('AV-D2', 9102, 'F 26', 'segundo', 10, 20, 'ok')
	`).Error; err != nil {
		return err
	}
	if err := gdb.Exec(`
		INSERT INTO poligonos_cuadrilla (feature_id, source_index, origen_ref, geom)
		VALUES
		  ('Z-D1', 9201, 'jefe:dup', ` + geom + `),
		  ('Z-D2', 9202, 'jefe:dup', ` + geom + `)
	`).Error; err != nil {
		return err
	}
	return gdb.Exec(`
		INSERT INTO capas_auxiliares (capa, feature_id, source_index, area_m2, perimetro_m, referencia)
		VALUES
		  ('jardines_reserva', 'JR-D1', 9301, -8, -1, repeat('y', 600)),
		  ('xerofitica', 'XE-D1', 9302, -2, 4, NULL)
	`).Error
}

func assertDuplicadosResueltos(t *testing.T, name string) {
	t.Helper()
	gdb, err := openTestDB(name)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	var n int
	if err := gdb.Raw(`SELECT count(*) FROM areas_verdes WHERE feature_id IN ('AV-D1', 'AV-D2')`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("se perdieron áreas: %d", n)
	}
	var codigo1, ref1 string
	var area1 sql.NullFloat64
	if err := gdb.Raw(`SELECT codigo, area_m2, referencia FROM areas_verdes WHERE feature_id = 'AV-D1'`).Row().Scan(&codigo1, &area1, &ref1); err != nil {
		t.Fatal(err)
	}
	if codigo1 != "F 26" {
		t.Fatalf("el menor id debía conservar F 26, tiene %q", codigo1)
	}
	if area1.Valid {
		t.Fatalf("area negativa debía quedar NULL, tiene %v", area1.Float64)
	}
	if len(ref1) != 500 {
		t.Fatalf("referencia larga debía recortarse a 500, tiene %d", len(ref1))
	}
	var codigo2 string
	if err := gdb.Raw(`SELECT codigo FROM areas_verdes WHERE feature_id = 'AV-D2'`).Scan(&codigo2).Error; err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(codigo2, "F 26 ·") {
		t.Fatalf("el duplicado debía llevar sufijo, tiene %q", codigo2)
	}
	var cambios int
	if err := gdb.Raw(`SELECT count(*) FROM cambios WHERE entidad = 'areas_verdes' AND accion = 'edicion'`).Scan(&cambios).Error; err != nil {
		t.Fatal(err)
	}
	if cambios < 2 {
		t.Fatalf("faltan ediciones en cambios: %d", cambios)
	}
	var idx int
	if err := gdb.Raw(`SELECT count(*) FROM pg_indexes WHERE indexname = 'areas_verdes_codigo_uidx'`).Scan(&idx).Error; err != nil {
		t.Fatal(err)
	}
	if idx != 1 {
		t.Fatal("no quedó el índice único de codigo")
	}

	var origen string
	if err := gdb.Raw(`SELECT origen_ref FROM poligonos_cuadrilla WHERE feature_id = 'Z-D2'`).Scan(&origen).Error; err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(origen, "jefe:dup ·") {
		t.Fatalf("origen_ref duplicado sin sufijo: %q", origen)
	}
	if err := gdb.Raw(`SELECT count(*) FROM poligonos_cuadrilla WHERE feature_id IN ('Z-D1', 'Z-D2')`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("se perdieron polígonos: %d", n)
	}

	var areaJ sql.NullFloat64
	var refJ string
	if err := gdb.Raw(`SELECT area_m2, coalesce(referencia, '') FROM jardines_reserva WHERE feature_id = 'JR-D1'`).Row().Scan(&areaJ, &refJ); err != nil {
		t.Fatal(err)
	}
	if areaJ.Valid || len(refJ) != 500 {
		t.Fatalf("jardín no se ajustó: area=%v ref=%d", areaJ, len(refJ))
	}
	var areaX sql.NullFloat64
	if err := gdb.Raw(`SELECT area_m2 FROM xerofiticas WHERE feature_id = 'XE-D1'`).Row().Scan(&areaX); err != nil {
		t.Fatal(err)
	}
	if areaX.Valid {
		t.Fatalf("xerofítica conservó area negativa: %v", areaX.Float64)
	}
}

func TestMigraciones001a019EnVacioYSobre008(t *testing.T) {
	base := os.Getenv("MIGRATE_TEST_URL")
	if base == "" {
		t.Skip("MIGRATE_TEST_URL no configurada; omitiendo test de base de datos")
	}
	admin, err := sql.Open("pgx", base)
	if err != nil {
		t.Fatalf("abrir conexion admin: %v", err)
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		t.Fatalf("sin postgres de prueba: %v", err)
	}

	dir := findMigrationsDir()
	vacia := "vp_c_ola1a_vacia"
	actual := "vp_c_ola1a_actual"
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
		if e.IsDir() || e.Name() >= "009" {
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
		t.Fatalf("001-008 sobre base vacía: %v", err)
	}
	if err := sembrarZona(actual); err != nil {
		t.Fatalf("semilla de zona: %v", err)
	}
	if err := aplicar(actual, dir); err != nil {
		t.Fatalf("010-019 sobre la base con 001-008: %v", err)
	}
	assertEsquema(t, actual)
	assertCopiaZona(t, actual)

	gdb, err := openTestDB(actual)
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
	sqlFiles := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			sqlFiles++
		}
	}
	if aplicadas != sqlFiles {
		t.Fatalf("migraciones aplicadas = %d, archivos = %d", aplicadas, sqlFiles)
	}
}

func TestMigracionesConCodigosDuplicados(t *testing.T) {
	base := os.Getenv("MIGRATE_TEST_URL")
	if base == "" {
		t.Skip("MIGRATE_TEST_URL no configurada; omitiendo test de base de datos")
	}
	admin, err := sql.Open("pgx", base)
	if err != nil {
		t.Fatalf("abrir conexion admin: %v", err)
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		t.Fatalf("sin postgres de prueba: %v", err)
	}

	dir := dirSerieHistorica()
	name := "vp_c_dup_codigo"
	recrear(t, admin, name)
	defer func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + name)
	}()

	previas := filepath.Join(t.TempDir(), "hasta013")
	if err := copiarHasta(dir, previas, "014"); err != nil {
		t.Fatal(err)
	}
	if err := aplicar(name, previas); err != nil {
		t.Fatalf("001-013: %v", err)
	}
	if err := sembrarDuplicados(name); err != nil {
		t.Fatalf("semilla duplicada: %v", err)
	}
	if err := aplicar(name, dir); err != nil {
		t.Fatalf("014 en adelante sobre duplicados: %v", err)
	}
	assertDuplicadosResueltos(t, name)
}

func TestComprobarNecesitaETL(t *testing.T) {
	base := os.Getenv("MIGRATE_TEST_URL")
	if base == "" {
		t.Skip("MIGRATE_TEST_URL no configurada; omitiendo test de base de datos")
	}
	admin, err := sql.Open("pgx", base)
	if err != nil {
		t.Fatalf("abrir conexion admin: %v", err)
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		t.Fatalf("sin postgres de prueba: %v", err)
	}

	name := fmt.Sprintf("vp_c_test_necesita_etl_%08x", rand.Uint32())
	recrear(t, admin, name)
	defer func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + name)
	}()

	dir := findMigrationsDir()
	if err := aplicar(name, dir); err != nil {
		t.Fatal(err)
	}

	gdb, err := openTestDB(name)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	// Tras migraciones, las tablas de catastro e inventario están vacías -> necesita=true
	necesita, conDatos, err := ComprobarNecesitaETL(sqlDB)
	if err != nil {
		t.Fatal(err)
	}
	if !necesita || len(conDatos) != 0 {
		t.Fatalf("esperaba necesita=true con 0 tablas con datos, obtuve necesita=%v conDatos=%v", necesita, conDatos)
	}

	// Insertamos un inventario -> necesita=false
	if err := gdb.Exec(`INSERT INTO inventario (capa, feature_id) VALUES ('arboles', 'ARB-001')`).Error; err != nil {
		t.Fatal(err)
	}
	necesita, conDatos, err = ComprobarNecesitaETL(sqlDB)
	if err != nil {
		t.Fatal(err)
	}
	if necesita {
		t.Fatal("esperaba necesita=false tras insertar inventario")
	}
	if len(conDatos) != 1 || !strings.Contains(conDatos[0], "inventario") {
		t.Fatalf("esperaba conDatos con inventario, obtuve: %v", conDatos)
	}
}

func TestBaselineNoSeReaplicaSiLaSerieHistoricaCerro(t *testing.T) {
	base := os.Getenv("MIGRATE_TEST_URL")
	if base == "" {
		t.Skip("MIGRATE_TEST_URL no configurada; omitiendo test de base de datos")
	}
	admin, err := sql.Open("pgx", base)
	if err != nil {
		t.Fatalf("abrir conexion admin: %v", err)
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		t.Fatalf("sin postgres de prueba: %v", err)
	}

	name := fmt.Sprintf("vp_c_baseline_%08x", rand.Uint32())
	recrear(t, admin, name)
	defer func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + name)
	}()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "078_medidas_palmera_baja.sql"), []byte("CREATE TABLE marca_historica (id int);"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := aplicar(name, dir); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "078_medidas_palmera_baja.sql")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "001_esquema_base.sql"), []byte("SELECT 1/0;"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "002_catalogos_base.sql"), []byte("SELECT 1/0;"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := aplicar(name, dir); err != nil {
		t.Fatalf("la serie consolidada no debía ejecutarse: %v", err)
	}

	gdb, err := openTestDB(name)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	var marca, fantasma, versiones int
	if err := gdb.Raw(`SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'marca_historica'`).Scan(&marca).Error; err != nil {
		t.Fatal(err)
	}
	if marca != 1 {
		t.Fatal("se perdió la tabla de la serie histórica")
	}
	if err := gdb.Raw(`SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'no_debe'`).Scan(&fantasma).Error; err != nil {
		t.Fatal(err)
	}
	if fantasma != 0 {
		t.Fatal("la baseline se ejecutó sobre una serie ya cerrada")
	}
	if err := gdb.Raw(`SELECT count(*) FROM schema_migrations WHERE version IN ('078_medidas_palmera_baja.sql', '001_esquema_base.sql', '002_catalogos_base.sql')`).Scan(&versiones).Error; err != nil {
		t.Fatal(err)
	}
	if versiones != 3 {
		t.Fatalf("versiones registradas = %d", versiones)
	}
}

func TestTransicionConservaDatosYRegistraBaseline(t *testing.T) {
	base := os.Getenv("MIGRATE_TEST_URL")
	if base == "" {
		t.Skip("MIGRATE_TEST_URL no configurada; omitiendo test de base de datos")
	}
	admin, err := sql.Open("pgx", base)
	if err != nil {
		t.Fatalf("abrir conexion admin: %v", err)
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		t.Fatalf("sin postgres de prueba: %v", err)
	}

	name := fmt.Sprintf("vp_c_transicion_%08x", rand.Uint32())
	recrear(t, admin, name)
	defer func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + name)
	}()

	if err := aplicar(name, dirSerieHistorica()); err != nil {
		t.Fatalf("serie histórica: %v", err)
	}
	gdb, err := openTestDB(name)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	if err := gdb.Exec(`UPDATE capataces SET equipo = 'MARCA-TRANSICION' WHERE id = 'cap-norte'`).Error; err != nil {
		t.Fatal(err)
	}
	var catalogos, versiones int
	if err := gdb.Raw(`SELECT count(*) FROM catalogos`).Scan(&catalogos).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Raw(`SELECT count(*) FROM schema_migrations`).Scan(&versiones).Error; err != nil {
		t.Fatal(err)
	}

	if err := aplicar(name, findMigrationsDir()); err != nil {
		t.Fatalf("transición: %v", err)
	}
	var equipo string
	if err := gdb.Raw(`SELECT equipo FROM capataces WHERE id = 'cap-norte'`).Scan(&equipo).Error; err != nil {
		t.Fatal(err)
	}
	if equipo != "MARCA-TRANSICION" {
		t.Fatalf("la transición alteró un dato ya cargado: %q", equipo)
	}
	var catalogosDespues, versionesDespues int
	if err := gdb.Raw(`SELECT count(*) FROM catalogos`).Scan(&catalogosDespues).Error; err != nil {
		t.Fatal(err)
	}
	if catalogosDespues != catalogos {
		t.Fatalf("catalogos %d → %d", catalogos, catalogosDespues)
	}
	if err := gdb.Raw(`SELECT count(*) FROM schema_migrations`).Scan(&versionesDespues).Error; err != nil {
		t.Fatal(err)
	}
	if versionesDespues != versiones+2 {
		t.Fatalf("schema_migrations %d → %d; se esperaban las dos consolidadas", versiones, versionesDespues)
	}
	var cubiertas int
	if err := gdb.Raw(`SELECT count(*) FROM schema_migrations WHERE version IN ('078_medidas_palmera_baja.sql', '001_esquema_base.sql', '002_catalogos_base.sql')`).Scan(&cubiertas).Error; err != nil {
		t.Fatal(err)
	}
	if cubiertas != 3 {
		t.Fatalf("faltan versiones de la transición: %d", cubiertas)
	}

	if err := aplicar(name, findMigrationsDir()); err != nil {
		t.Fatalf("segunda pasada: %v", err)
	}
	var otraVez int
	if err := gdb.Raw(`SELECT count(*) FROM schema_migrations`).Scan(&otraVez).Error; err != nil {
		t.Fatal(err)
	}
	if otraVez != versionesDespues {
		t.Fatalf("la segunda pasada registró versiones de más: %d → %d", versionesDespues, otraVez)
	}
}

func TestEsquemaConsolidadoIgualASerieHistorica(t *testing.T) {
	base := os.Getenv("MIGRATE_TEST_URL")
	if base == "" {
		t.Skip("MIGRATE_TEST_URL no configurada; omitiendo test de base de datos")
	}
	admin, err := sql.Open("pgx", base)
	if err != nil {
		t.Fatalf("abrir conexion admin: %v", err)
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		t.Fatalf("sin postgres de prueba: %v", err)
	}

	vieja := fmt.Sprintf("vp_c_eq_vieja_%08x", rand.Uint32())
	nueva := fmt.Sprintf("vp_c_eq_nueva_%08x", rand.Uint32())
	recrear(t, admin, vieja)
	recrear(t, admin, nueva)
	defer func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname IN ($1, $2) AND pid <> pg_backend_pid()`, vieja, nueva)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + vieja)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + nueva)
	}()
	if err := aplicar(vieja, dirSerieHistorica()); err != nil {
		t.Fatalf("serie histórica: %v", err)
	}
	if err := aplicar(nueva, findMigrationsDir()); err != nil {
		t.Fatalf("serie nueva: %v", err)
	}

	consultas := []string{
		`SELECT format('%s.%s|%s|%s|%s', c.table_name, c.column_name, c.data_type, c.udt_name, c.is_nullable, coalesce(c.column_default, ''))
		 FROM information_schema.columns c
		 WHERE c.table_schema = 'public' AND c.table_name <> 'spatial_ref_sys'
		 ORDER BY 1`,
		`SELECT format('%s|%s|%s', r.relname, c.conname, pg_get_constraintdef(c.oid))
		 FROM pg_constraint c
		 JOIN pg_class r ON r.oid = c.conrelid
		 JOIN pg_namespace n ON n.oid = r.relnamespace
		 WHERE n.nspname = 'public' AND r.relname <> 'spatial_ref_sys'
		 ORDER BY 1`,
		`SELECT indexdef FROM pg_indexes WHERE schemaname = 'public' AND tablename <> 'spatial_ref_sys' ORDER BY indexdef`,
		`SELECT format('%s|%s', c.relname, pg_get_viewdef(c.oid))
		 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		 WHERE n.nspname = 'public' AND c.relkind = 'v' ORDER BY 1`,
		`SELECT format('%s|%s', p.proname, pg_get_functiondef(p.oid))
		 FROM pg_proc p
		 JOIN pg_namespace n ON n.oid = p.pronamespace
		 WHERE n.nspname = 'public'
		   AND NOT EXISTS (SELECT 1 FROM pg_depend d WHERE d.objid = p.oid AND d.deptype = 'e')
		 ORDER BY 1`,
		`SELECT pg_get_triggerdef(t.oid)
		 FROM pg_trigger t
		 JOIN pg_class c ON c.oid = t.tgrelid
		 JOIN pg_namespace n ON n.oid = c.relnamespace
		 WHERE n.nspname = 'public' AND NOT t.tgisinternal
		 ORDER BY 1`,
	}
	for i, q := range consultas {
		if diff := diferenciaListas(t, vieja, nueva, q); diff != "" {
			t.Fatalf("consulta %d:\n%s", i, diff)
		}
	}
	if diff := diferenciaListas(t, vieja, nueva, `
		SELECT format('%s|%s', sequencename, last_value)
		FROM pg_sequences WHERE schemaname = 'public' ORDER BY 1`); diff != "" {
		t.Fatalf("secuencias:\n%s", diff)
	}
	for _, tabla := range leerLista(t, vieja, `
		SELECT c.relname
		FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' AND c.relkind = 'r'
		  AND c.relname NOT IN ('schema_migrations', 'spatial_ref_sys')
		ORDER BY 1`) {
		q := consultaDatos(t, vieja, tabla)
		if diff := diferenciaListas(t, vieja, nueva, q); diff != "" {
			t.Fatalf("datos %s:\n%s", tabla, diff)
		}
	}
}

func consultaDatos(t *testing.T, dbName, tabla string) string {
	t.Helper()
	if !identSQL.MatchString(tabla) {
		t.Fatalf("tabla inesperada %s", tabla)
	}
	cols := leerLista(t, dbName, fmt.Sprintf(`
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = '%s'
		  AND data_type <> 'timestamp with time zone'
		ORDER BY ordinal_position`, tabla))
	if len(cols) == 0 {
		return fmt.Sprintf(`SELECT count(*)::text FROM %s`, tabla)
	}
	partes := make([]string, 0, len(cols))
	for _, col := range cols {
		if !identSQL.MatchString(col) || !identSQL.MatchString(tabla) {
			t.Fatalf("identificador inesperado %s.%s", tabla, col)
		}
		partes = append(partes, fmt.Sprintf("coalesce(%s::text, '∅')", col))
	}
	return fmt.Sprintf(`
		SELECT coalesce(string_agg(linea, E'\n' ORDER BY linea), '')
		FROM (SELECT concat_ws('|', %s) AS linea FROM %s) s`,
		strings.Join(partes, ", "), tabla)
}

func diferenciaListas(t *testing.T, a, b, query string) string {
	t.Helper()
	la := leerLista(t, a, query)
	lb := leerLista(t, b, query)
	if strings.Join(la, "\n") == strings.Join(lb, "\n") {
		return ""
	}
	var bld strings.Builder
	fmt.Fprintf(&bld, "vieja=%d nueva=%d\n", len(la), len(lb))
	vistos := map[string]int{}
	for _, s := range la {
		vistos[s]++
	}
	for _, s := range lb {
		vistos[s]--
	}
	n := 0
	for s, d := range vistos {
		if d == 0 {
			continue
		}
		n++
		if n > 12 {
			fmt.Fprintf(&bld, "...\n")
			break
		}
		fmt.Fprintf(&bld, "%+d %s\n", d, s)
	}
	return bld.String()
}

func leerLista(t *testing.T, name, query string) []string {
	t.Helper()
	gdb, err := openTestDB(name)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	var out []string
	if err := gdb.Raw(query).Scan(&out).Error; err != nil {
		t.Fatal(err)
	}
	return out
}
