package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

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

func findMigrationsDir() string {
	if env := os.Getenv("MIGRATIONS_DIR"); env != "" {
		if fi, err := os.Stat(env); err == nil && fi.IsDir() {
			return env
		}
	}
	wd, err := os.Getwd()
	if err == nil {
		for dir := wd; ; {
			candidate := filepath.Join(dir, "apps", "api", "migrations")
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
	return filepath.Join("..", "..", "..", "..", "..", "apps", "api", "migrations")
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
	if err := gdb.Exec(`UPDATE poligonos_cuadrilla SET sector = 'otro' WHERE feature_id = 'Z-9101'`).Error; err == nil {
		t.Fatalf("%s: el CHECK debería rechazar un sector fuera de la lista", name)
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

	dir := findMigrationsDir()
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

	name := "vp_c_test_necesita_etl"
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
