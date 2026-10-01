// Command modelgen generates GORM models and checks schema drift against PostgreSQL.
package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	gormschema "gorm.io/gorm/schema"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func main() {
	if err := run(); err != nil {
		log.Error().Err(err).Msg("Error en modelgen / control de deriva")
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.GetConfig()
	return runWithConfig(cfg)
}

func runWithConfig(cfg *config.Config) error {
	// Guarda 1 y 2: validar nombre y prefijo desechable antes de conectar
	if err := ValidateDatabaseTarget(cfg.Database.Name, cfg.Database.URL); err != nil {
		return err
	}

	db, err := database.NewDatabaseConnection(cfg)
	if err != nil {
		return fmt.Errorf("conectar a base de datos: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("obtener conexion SQL: %w", err)
	}
	defer sqlDB.Close()

	// Guarda 3 y 4: validar que sea construida por migraciones y sin datos cargados
	if err := ValidateDatabaseState(db, cfg.Database.Schema); err != nil {
		return err
	}

	// Directorio de salida: MODELGEN_OUT o directorio temporal por defecto (nunca bajo backend/app)
	outDir := strings.TrimSpace(os.Getenv("MODELGEN_OUT"))
	isTempDir := false
	if outDir == "" {
		tmp, err := os.MkdirTemp("", "modelgen_*")
		if err != nil {
			return fmt.Errorf("crear directorio temporal: %w", err)
		}
		outDir = tmp
		isTempDir = true
	} else {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("crear directorio de salida %q: %w", outDir, err)
		}
	}
	if isTempDir {
		defer os.RemoveAll(outDir)
	}

	tables, err := applicationTables(db, cfg.Database.Schema)
	if err != nil {
		return fmt.Errorf("listar tablas del esquema %q: %w", cfg.Database.Schema, err)
	}
	if len(tables) == 0 {
		return fmt.Errorf("el esquema %q no contiene tablas de aplicacion", cfg.Database.Schema)
	}

	indexOptions, err := modelIndexOptions(db, cfg.Database.Schema)
	if err != nil {
		return fmt.Errorf("leer indices del esquema %q: %w", cfg.Database.Schema, err)
	}

	generator := gen.NewGenerator(gen.Config{
		OutPath:             filepath.Join(outDir, "query"),
		ModelPkgPath:        outDir,
		FieldNullable:       true,
		FieldWithTypeTag:    true,
		FieldWithDefaultTag: true,
		UseAny:              true,
	})
	generator.UseDB(db)
	generator.WithModelNameStrategy(
		gormschema.NamingStrategy{SingularTable: true}.SchemaName,
	)

	generatedFiles := make(map[string]struct{}, len(tables))
	for _, table := range tables {
		model := generator.GenerateModel(table, indexOptions[table]...)
		generatedFiles[model.FileName+".gen.go"] = struct{}{}
	}

	generator.Execute()

	if err := removeStaleGeneratedFiles(outDir, generatedFiles); err != nil {
		return err
	}

	log.Info().Int("tables", len(tables)).Str("output", outDir).Msg("Modelos GORM generados exitosamente")
	fmt.Printf("Modelos generados exitosamente (%d tablas) en: %s\n", len(tables), outDir)

	// Control de deriva contra nuestros modelos hand-written en internal/persistence/models
	driftItems, err := CheckDrift(db, cfg.Database.Schema)
	if err != nil {
		return fmt.Errorf("ejecutar control de deriva: %w", err)
	}

	if len(driftItems) > 0 {
		report := FormatDriftReport(driftItems)
		fmt.Fprintln(os.Stderr, report)
		return fmt.Errorf("control de deriva fallido: se encontraron %d discrepancias", len(driftItems))
	}

	fmt.Println("Control de deriva exitoso: todos los modelos coinciden con las tablas de la base de datos.")
	return nil
}

// IsCampusVerde checks if the target database name or URL points to the production/dev dataset database.
func IsCampusVerde(dbName string, dbURL string) bool {
	if dbName == "campus_verde" {
		return true
	}
	if dbURL == "" {
		return false
	}
	u, err := url.Parse(dbURL)
	if err == nil {
		cleanPath := strings.TrimPrefix(u.Path, "/")
		if cleanPath == "campus_verde" {
			return true
		}
	}
	if strings.Contains(dbURL, "/campus_verde") || strings.Contains(dbURL, "dbname=campus_verde") {
		return true
	}
	return false
}

// EffectiveDatabaseName resolves the actual database name from URL or database config name.
func EffectiveDatabaseName(dbName string, dbURL string) string {
	if dbURL != "" {
		u, err := url.Parse(dbURL)
		if err == nil {
			cleanPath := strings.TrimPrefix(u.Path, "/")
			if cleanPath != "" {
				return cleanPath
			}
		}
		for _, part := range strings.Fields(dbURL) {
			if strings.HasPrefix(part, "dbname=") {
				return strings.TrimPrefix(part, "dbname=")
			}
		}
	}
	return dbName
}

// ValidateDatabaseTarget verifies that the database is disposable and not campus_verde.
func ValidateDatabaseTarget(dbName string, dbURL string) error {
	if IsCampusVerde(dbName, dbURL) {
		return fmt.Errorf("modelgen abortado: no se permite ejecutar modelgen contra la base de datos de datos reales 'campus_verde'")
	}

	target := EffectiveDatabaseName(dbName, dbURL)
	if !strings.HasPrefix(target, "vp_") && !strings.HasPrefix(target, "modelgen_") {
		return fmt.Errorf("modelgen abortado: la base de datos debe ser desechable y comenzar con 'vp_' o 'modelgen_' (obtenido: '%s')", target)
	}
	return nil
}

// ValidateDatabaseStateCounts verifies database state counts.
func ValidateDatabaseStateCounts(hasSchemaMigrations bool, areasVerdesCount int64) error {
	if !hasSchemaMigrations {
		return fmt.Errorf("modelgen abortado: la base de datos no fue construida con nuestras migraciones (falta la tabla 'schema_migrations')")
	}
	if areasVerdesCount > 0 {
		return fmt.Errorf("modelgen abortado: la base de datos contiene datos cargados (areas_verdes tiene %d filas). Solo se permite ejecutar contra una base desechable vacía construida por cmd/migrate sin ETL", areasVerdesCount)
	}
	return nil
}

// ValidateDatabaseState queries the database to verify it was built with migrations and has no loaded data.
func ValidateDatabaseState(db *gorm.DB, schemaName string) error {
	var countMigrationsTable int64
	err := db.Raw(`
		SELECT count(*)
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_name = 'schema_migrations'
	`, schemaName).Scan(&countMigrationsTable).Error
	if err != nil {
		return fmt.Errorf("verificar tabla schema_migrations: %w", err)
	}

	var hasAreasVerdes int64
	err = db.Raw(`
		SELECT count(*)
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_name = 'areas_verdes'
	`, schemaName).Scan(&hasAreasVerdes).Error
	if err != nil {
		return fmt.Errorf("verificar tabla areas_verdes: %w", err)
	}

	var areasVerdesCount int64
	if hasAreasVerdes > 0 {
		err = db.Raw(`SELECT count(*) FROM areas_verdes`).Scan(&areasVerdesCount).Error
		if err != nil {
			return fmt.Errorf("contar filas en areas_verdes: %w", err)
		}
	}

	return ValidateDatabaseStateCounts(countMigrationsTable > 0, areasVerdesCount)
}

// DriftItem represents a schema discrepancy between hand-written models and PostgreSQL schema.
type DriftItem struct {
	Model     string
	Table     string
	Column    string
	IssueType string
	Detail    string
}

// ColumnMetadata holds column data type information from information_schema.
type ColumnMetadata struct {
	DataType string
	UDTName  string
}

// CheckDrift compares internal/persistence/models against the PostgreSQL database schema.
func CheckDrift(db *gorm.DB, schemaName string) ([]DriftItem, error) {
	var tableNames []string
	err := db.Raw(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_type = 'BASE TABLE'
	`, schemaName).Scan(&tableNames).Error
	if err != nil {
		return nil, fmt.Errorf("listar tablas existentes: %w", err)
	}

	existingTables := make(map[string]bool, len(tableNames))
	for _, t := range tableNames {
		existingTables[t] = true
	}

	type colRow struct {
		TableName  string `gorm:"column:table_name"`
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
		UDTName    string `gorm:"column:udt_name"`
	}

	var colRows []colRow
	err = db.Raw(`
		SELECT table_name, column_name, data_type, udt_name
		FROM information_schema.columns
		WHERE table_schema = ?
	`, schemaName).Scan(&colRows).Error
	if err != nil {
		return nil, fmt.Errorf("listar columnas existentes: %w", err)
	}

	tableColumns := make(map[string]map[string]ColumnMetadata)
	for _, r := range colRows {
		if _, ok := tableColumns[r.TableName]; !ok {
			tableColumns[r.TableName] = make(map[string]ColumnMetadata)
		}
		tableColumns[r.TableName][r.ColumnName] = ColumnMetadata{
			DataType: r.DataType,
			UDTName:  r.UDTName,
		}
	}

	allModels := models.AllModels()
	namer := gormschema.NamingStrategy{}
	cache := &sync.Map{}
	var drift []DriftItem

	for _, m := range allModels {
		s, err := gormschema.Parse(m, cache, namer)
		if err != nil {
			return nil, fmt.Errorf("analizar modelo %T: %w", m, err)
		}
		tableName := s.Table
		modelName := fmt.Sprintf("%T", m)

		if !existingTables[tableName] {
			drift = append(drift, DriftItem{
				Model:     modelName,
				Table:     tableName,
				IssueType: "TABLA_FALTANTE",
				Detail:    fmt.Sprintf("la tabla %q no existe en la base de datos", tableName),
			})
			continue
		}

		cols := tableColumns[tableName]
		for _, f := range s.Fields {
			if f.DBName == "" {
				continue
			}
			ci, ok := cols[f.DBName]
			if !ok {
				// Column not in table: check if it's a read-only joined/virtual field
				if _, isReadOnly := f.TagSettings["->"]; isReadOnly {
					continue
				}
				drift = append(drift, DriftItem{
					Model:     modelName,
					Table:     tableName,
					Column:    f.DBName,
					IssueType: "COLUMNA_FALTANTE",
					Detail:    fmt.Sprintf("columna %q mapeada por campo %s no existe en tabla %q", f.DBName, f.Name, tableName),
				})
				continue
			}

			if !IsTypeCompatible(f.FieldType, ci.DataType, ci.UDTName) {
				drift = append(drift, DriftItem{
					Model:     modelName,
					Table:     tableName,
					Column:    f.DBName,
					IssueType: "TIPO_INCOMPATIBLE",
					Detail: fmt.Sprintf("campo %s (Go: %s) es incompatible con columna %s.%s (BD: %s / %s)",
						f.Name, f.FieldType.String(), tableName, f.DBName, ci.DataType, ci.UDTName),
				})
			}
		}
	}

	return drift, nil
}

// IsTypeCompatible checks whether a Go reflect.Type is compatible with PostgreSQL column data types.
func IsTypeCompatible(goType reflect.Type, pgDataType string, pgUdtName string) bool {
	if goType.Kind() == reflect.Pointer {
		goType = goType.Elem()
	}

	dt := strings.ToLower(pgDataType)
	udt := strings.ToLower(pgUdtName)

	if goType == reflect.TypeOf(time.Time{}) {
		return strings.Contains(dt, "time") || strings.Contains(dt, "date") ||
			strings.Contains(udt, "time") || strings.Contains(udt, "date")
	}

	switch goType.Kind() {
	case reflect.String:
		return dt == "character varying" || dt == "text" || dt == "uuid" ||
			dt == "user-defined" || dt == "jsonb" || dt == "json" ||
			dt == "date" || strings.Contains(dt, "time") ||
			udt == "varchar" || udt == "text" || udt == "uuid" ||
			udt == "geometry" || udt == "jsonb" || udt == "date" || udt == "time"
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8,
		reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return dt == "integer" || dt == "bigint" || dt == "smallint" || dt == "numeric" ||
			udt == "int2" || udt == "int4" || udt == "int8" || udt == "numeric"
	case reflect.Float32, reflect.Float64:
		return dt == "double precision" || dt == "real" || dt == "numeric" ||
			udt == "float4" || udt == "float8" || udt == "numeric"
	case reflect.Bool:
		return dt == "boolean" || udt == "bool"
	case reflect.Slice:
		if goType.Elem().Kind() == reflect.Uint8 {
			return dt == "bytea" || dt == "text" || dt == "jsonb" || dt == "json" ||
				udt == "bytea" || udt == "text" || udt == "jsonb"
		}
	case reflect.Map:
		return dt == "jsonb" || dt == "json" || udt == "jsonb" || udt == "json"
	}
	return false
}

// FormatDriftReport formats drift issues into a human-readable Spanish report.
func FormatDriftReport(items []DriftItem) string {
	var b strings.Builder
	b.WriteString("\n=== REPORTE DE CONTROL DE DERIVA ===\n")
	b.WriteString(fmt.Sprintf("Discrepancias detectadas: %d\n", len(items)))
	for _, item := range items {
		b.WriteString(fmt.Sprintf("- [%s] Modelo %s (tabla: %s", item.IssueType, item.Model, item.Table))
		if item.Column != "" {
			b.WriteString(fmt.Sprintf(", columna: %s", item.Column))
		}
		b.WriteString(fmt.Sprintf("): %s\n", item.Detail))
	}
	return b.String()
}

// applicationTables excludes tables installed and owned by PostgreSQL extensions.
func applicationTables(db *gorm.DB, schema string) ([]string, error) {
	const query = `
		SELECT table_info.table_name
		FROM information_schema.tables AS table_info
		WHERE table_info.table_schema = ?
		  AND table_info.table_type = 'BASE TABLE'
		  AND NOT EXISTS (
			SELECT 1
			FROM pg_catalog.pg_depend AS dependency
			JOIN pg_catalog.pg_class AS relation ON relation.oid = dependency.objid
			JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid = relation.relnamespace
			JOIN pg_catalog.pg_extension AS extension ON extension.oid = dependency.refobjid
			WHERE namespace.nspname = table_info.table_schema
			  AND relation.relname = table_info.table_name
		  )
		ORDER BY table_info.table_name`

	var tables []string
	if err := db.Raw(query, schema).Scan(&tables).Error; err != nil {
		return nil, err
	}
	return tables, nil
}

type indexMetadata struct {
	TableName   string  `gorm:"column:table_name"`
	IndexName   string  `gorm:"column:index_name"`
	ColumnName  string  `gorm:"column:column_name"`
	IndexMethod string  `gorm:"column:index_method"`
	Priority    int     `gorm:"column:priority"`
	IsUnique    bool    `gorm:"column:is_unique"`
	Predicate   *string `gorm:"column:predicate"`
}

func modelIndexOptions(db *gorm.DB, schema string) (map[string][]gen.ModelOpt, error) {
	const query = `
		SELECT
			table_relation.relname AS table_name,
			index_relation.relname AS index_name,
			column_info.attname AS column_name,
			access_method.amname AS index_method,
			index_column.ordinality AS priority,
			index_info.indisunique AS is_unique,
			pg_get_expr(index_info.indpred, index_info.indrelid) AS predicate
		FROM pg_catalog.pg_index AS index_info
		JOIN pg_catalog.pg_class AS table_relation
		  ON table_relation.oid = index_info.indrelid
		JOIN pg_catalog.pg_namespace AS namespace
		  ON namespace.oid = table_relation.relnamespace
		JOIN pg_catalog.pg_class AS index_relation
		  ON index_relation.oid = index_info.indexrelid
		JOIN pg_catalog.pg_am AS access_method
		  ON access_method.oid = index_relation.relam
		CROSS JOIN LATERAL unnest(index_info.indkey)
		  WITH ORDINALITY AS index_column(attribute_number, ordinality)
		JOIN pg_catalog.pg_attribute AS column_info
		  ON column_info.attrelid = table_relation.oid
		 AND column_info.attnum = index_column.attribute_number
		WHERE namespace.nspname = ?
		  AND NOT index_info.indisprimary
		ORDER BY table_relation.relname, index_relation.relname, index_column.ordinality`

	var indexes []indexMetadata
	if err := db.Raw(query, schema).Scan(&indexes).Error; err != nil {
		return nil, err
	}

	options := make(map[string][]gen.ModelOpt)
	for _, index := range indexes {
		index := index
		options[index.TableName] = append(options[index.TableName], gen.FieldGORMTag(
			index.ColumnName,
			func(tag field.GormTag) field.GormTag {
				tagName := field.TagKeyGormIndex
				if index.IsUnique {
					tagName = field.TagKeyGormUniqueIndex
				}

				value := fmt.Sprintf("%s,priority:%d,type:%s", index.IndexName, index.Priority, index.IndexMethod)
				if index.Predicate != nil && strings.TrimSpace(*index.Predicate) != "" {
					value += ",where:" + *index.Predicate
				}
				tag.Append(tagName, value)
				return tag
			},
		))
	}
	return options, nil
}

func removeStaleGeneratedFiles(directory string, expected map[string]struct{}) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("read generated models directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".gen.go") {
			continue
		}
		if _, ok := expected[entry.Name()]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(directory, entry.Name())); err != nil {
			return fmt.Errorf("remove stale generated model %q: %w", entry.Name(), err)
		}
	}
	return nil
}
