package main

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestValidateDatabaseTarget_CampusVerdeAborts(t *testing.T) {
	cases := []struct {
		name   string
		dbName string
		dbURL  string
	}{
		{
			name:   "exact dbName campus_verde",
			dbName: "campus_verde",
			dbURL:  "",
		},
		{
			name:   "URL path campus_verde",
			dbName: "",
			dbURL:  "postgres://user:pass@127.0.0.1:5432/campus_verde?sslmode=disable",
		},
		{
			name:   "URL path campus_verde with other dbName",
			dbName: "vp_test",
			dbURL:  "postgres://user:pass@127.0.0.1:5432/campus_verde?sslmode=disable",
		},
		{
			name:   "dbName campus_verde with disposable URL",
			dbName: "campus_verde",
			dbURL:  "postgres://user:pass@127.0.0.1:5432/vp_test?sslmode=disable",
		},
		{
			name:   "DSN containing dbname=campus_verde",
			dbName: "",
			dbURL:  "host=127.0.0.1 dbname=campus_verde user=campus",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateDatabaseTarget(tc.dbName, tc.dbURL)
			if err == nil {
				t.Fatal("esperaba error abortando por campus_verde, obtuve nil")
			}
			if !strings.Contains(err.Error(), "campus_verde") {
				t.Fatalf("mensaje esperado con 'campus_verde', obtuve: %v", err)
			}
		})
	}
}

func TestValidateDatabaseTarget_DisposablePrefix(t *testing.T) {
	invalidCases := []struct {
		name   string
		dbName string
		dbURL  string
	}{
		{
			name:   "dbName without prefix",
			dbName: "areasverdes",
		},
		{
			name:   "dbName with test suffix only",
			dbName: "test_db",
		},
		{
			name:  "URL without prefix",
			dbURL: "postgres://user:pass@127.0.0.1:5432/produccion?sslmode=disable",
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateDatabaseTarget(tc.dbName, tc.dbURL)
			if err == nil {
				t.Fatal("esperaba error por nombre de base de datos no desechable, obtuve nil")
			}
			if !strings.Contains(err.Error(), "vp_") || !strings.Contains(err.Error(), "modelgen_") {
				t.Fatalf("mensaje esperado indicando prefijos permitidos, obtuve: %v", err)
			}
		})
	}

	validCases := []struct {
		name   string
		dbName string
		dbURL  string
	}{
		{
			name:   "dbName with vp_ prefix",
			dbName: "vp_test_temp",
		},
		{
			name:   "dbName with modelgen_ prefix",
			dbName: "modelgen_temp_123",
		},
		{
			name:  "URL with vp_ prefix",
			dbURL: "postgres://user:pass@127.0.0.1:5432/vp_c_test123?sslmode=disable",
		},
		{
			name:  "URL with modelgen_ prefix",
			dbURL: "postgres://user:pass@127.0.0.1:5432/modelgen_introspect?sslmode=disable",
		},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateDatabaseTarget(tc.dbName, tc.dbURL)
			if err != nil {
				t.Fatalf("no esperaba error para base desechable valida, obtuve: %v", err)
			}
		})
	}
}

func TestValidateDatabaseStateCounts(t *testing.T) {
	// Sin schema_migrations debe fallar
	err := ValidateDatabaseStateCounts(false, 0)
	if err == nil {
		t.Fatal("esperaba error cuando falta schema_migrations")
	}
	if !strings.Contains(err.Error(), "schema_migrations") {
		t.Fatalf("mensaje esperado mencionando schema_migrations, obtuve: %v", err)
	}

	// Con datos cargados en areas_verdes debe fallar
	err = ValidateDatabaseStateCounts(true, 521)
	if err == nil {
		t.Fatal("esperaba error cuando areas_verdes tiene datos cargados")
	}
	if !strings.Contains(err.Error(), "areas_verdes") || !strings.Contains(err.Error(), "521") {
		t.Fatalf("mensaje esperado mencionando areas_verdes y conteo de filas, obtuve: %v", err)
	}

	// Con schema_migrations y sin datos cargados debe pasar
	err = ValidateDatabaseStateCounts(true, 0)
	if err != nil {
		t.Fatalf("no esperaba error para base de datos recien migrada y vacia, obtuve: %v", err)
	}
}

func TestIsTypeCompatible(t *testing.T) {
	tests := []struct {
		name       string
		goType     reflect.Type
		pgDataType string
		pgUdtName  string
		want       bool
	}{
		{
			name:       "string with varchar",
			goType:     reflect.TypeOf(""),
			pgDataType: "character varying",
			pgUdtName:  "varchar",
			want:       true,
		},
		{
			name:       "string with uuid",
			goType:     reflect.TypeOf(""),
			pgDataType: "uuid",
			pgUdtName:  "uuid",
			want:       true,
		},
		{
			name:       "pointer string with text",
			goType:     reflect.TypeOf(new(string)),
			pgDataType: "text",
			pgUdtName:  "text",
			want:       true,
		},
		{
			name:       "int64 with bigint",
			goType:     reflect.TypeOf(int64(0)),
			pgDataType: "bigint",
			pgUdtName:  "int8",
			want:       true,
		},
		{
			name:       "int with integer",
			goType:     reflect.TypeOf(int(0)),
			pgDataType: "integer",
			pgUdtName:  "int4",
			want:       true,
		},
		{
			name:       "float64 with double precision",
			goType:     reflect.TypeOf(float64(0)),
			pgDataType: "double precision",
			pgUdtName:  "float8",
			want:       true,
		},
		{
			name:       "bool with boolean",
			goType:     reflect.TypeOf(true),
			pgDataType: "boolean",
			pgUdtName:  "bool",
			want:       true,
		},
		{
			name:       "time.Time with timestamptz",
			goType:     reflect.TypeOf(time.Time{}),
			pgDataType: "timestamp with time zone",
			pgUdtName:  "timestamptz",
			want:       true,
		},
		{
			name:       "time.Time with time without tz",
			goType:     reflect.TypeOf(time.Time{}),
			pgDataType: "time without time zone",
			pgUdtName:  "time",
			want:       true,
		},
		{
			name:       "incompatible int with text",
			goType:     reflect.TypeOf(int(0)),
			pgDataType: "text",
			pgUdtName:  "text",
			want:       false,
		},
		{
			name:       "incompatible bool with integer",
			goType:     reflect.TypeOf(true),
			pgDataType: "integer",
			pgUdtName:  "int4",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTypeCompatible(tt.goType, tt.pgDataType, tt.pgUdtName)
			if got != tt.want {
				t.Errorf("IsTypeCompatible(%s, %s, %s) = %v, want %v",
					tt.goType.String(), tt.pgDataType, tt.pgUdtName, got, tt.want)
			}
		})
	}
}

func TestFormatDriftReport(t *testing.T) {
	items := []DriftItem{
		{
			Model:     "*models.TestModel",
			Table:     "tabla_test",
			Column:    "col_test",
			IssueType: "COLUMNA_FALTANTE",
			Detail:    "columna no encontrada",
		},
	}
	report := FormatDriftReport(items)
	if !strings.Contains(report, "COLUMNA_FALTANTE") || !strings.Contains(report, "tabla_test") {
		t.Fatalf("reporte incompleto: %s", report)
	}
}

// TestDriftCheckAgainstDisposableDB runs a full drift check against a temporary database built from our migrations.
func TestDriftCheckAgainstDisposableDB(t *testing.T) {
	if os.Getenv("MIGRATE_TEST_URL") == "" {
		t.Skip("MIGRATE_TEST_URL no configurada; omitiendo test de base de datos")
	}

	rawDB, gdb := testutil.MigrarDBTemporal(t, "modelgen")
	defer rawDB.Close()

	var dbName string
	if err := gdb.Raw("SELECT current_database()").Scan(&dbName).Error; err != nil {
		t.Fatalf("obtener nombre de base de datos actual: %v", err)
	}

	// Validar que el nombre desechable pase los guards
	if err := ValidateDatabaseTarget(dbName, ""); err != nil {
		t.Fatalf("ValidateDatabaseTarget fallo para base desechable %q: %v", dbName, err)
	}

	// Validar que el estado de la BD (con migraciones aplicadas y sin ETL) pase los guards
	if err := ValidateDatabaseState(gdb, "public"); err != nil {
		t.Fatalf("ValidateDatabaseState fallo para base migrada %q: %v", dbName, err)
	}

	// Ejecutar control de deriva: debe dar 0 discrepancias
	driftItems, err := CheckDrift(gdb, "public")
	if err != nil {
		t.Fatalf("CheckDrift fallo con error: %v", err)
	}

	if len(driftItems) > 0 {
		report := FormatDriftReport(driftItems)
		t.Fatalf("se detectaron discrepancias de deriva inesperadas:\n%s", report)
	}
}

// TestDriftCheckDetectsMissingTable tests that CheckDrift catches when a table is missing.
func TestCheckDrift_DetectsMissingTable(t *testing.T) {
	if os.Getenv("MIGRATE_TEST_URL") == "" {
		t.Skip("MIGRATE_TEST_URL no configurada; omitiendo test de base de datos")
	}

	rawDB, gdb := testutil.MigrarDBTemporal(t, "modelgen_drift")
	defer rawDB.Close()

	// Drop a table to induce drift
	if err := gdb.Exec("DROP TABLE IF EXISTS bebederos CASCADE").Error; err != nil {
		t.Fatalf("error induciendo deriva eliminando tabla bebederos: %v", err)
	}

	driftItems, err := CheckDrift(gdb, "public")
	if err != nil {
		t.Fatalf("CheckDrift fallo: %v", err)
	}

	found := false
	for _, item := range driftItems {
		if item.Table == "bebederos" && item.IssueType == "TABLA_FALTANTE" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("esperaba detectar deriva TABLA_FALTANTE para bebederos, discrepancias encontradas: %v", driftItems)
	}
}
