// Command modelgen generates GORM models by introspecting PostgreSQL.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	gormschema "gorm.io/gorm/schema"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func main() {
	if err := run(); err != nil {
		log.Error().Err(err).Msg("Error generando modelos GORM")
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.GetConfig()

	// Guarda 1: abortar si DATABASE_NAME es campus_verde o URL apunta a campus_verde
	if cfg.Database.Name == "campus_verde" || strings.Contains(cfg.Database.URL, "campus_verde") {
		return fmt.Errorf("modelgen abortado: no se permite ejecutar modelgen contra la base de datos de datos reales 'campus_verde'")
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

	// Guarda 2: abortar si la BD contiene tablas del sistema de datos existente (schema_migrations o areas_verdes)
	var countDataTables int64
	err = db.Raw(`
		SELECT count(*)
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_name IN ('schema_migrations', 'areas_verdes')
	`, cfg.Database.Schema).Scan(&countDataTables).Error
	if err != nil {
		return fmt.Errorf("verificar tablas del sistema: %w", err)
	}
	if countDataTables > 0 {
		return fmt.Errorf("modelgen abortado: la base de datos contiene tablas del sistema de datos (schema_migrations/areas_verdes). Solo se permite ejecutar contra una base con el esquema v1.1")
	}

	// Guarda 3: verificar que la base contenga el esquema v1.1 (debe existir la tabla intervencion)
	var countIntervencion int64
	err = db.Raw(`
		SELECT count(*)
		FROM information_schema.tables
		WHERE table_schema = ?
		  AND table_name = 'intervencion'
	`, cfg.Database.Schema).Scan(&countIntervencion).Error
	if err != nil {
		return fmt.Errorf("verificar esquema v1.1: %w", err)
	}
	if countIntervencion == 0 {
		return fmt.Errorf("modelgen abortado: la base de datos no contiene el esquema v1.1 (falta la tabla 'intervencion')")
	}

	// Directorio de salida: MODELGEN_OUT o directorio temporal por defecto (nunca en backend/app)
	outDir := strings.TrimSpace(os.Getenv("MODELGEN_OUT"))
	if outDir == "" {
		tmp, err := os.MkdirTemp("", "modelgen_v11_*")
		if err != nil {
			return fmt.Errorf("crear directorio temporal: %w", err)
		}
		outDir = tmp
	} else {
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("crear directorio de salida %q: %w", outDir, err)
		}
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

	log.Info().Int("tables", len(tables)).Str("output", outDir).Msg("GORM models generated successfully")
	fmt.Printf("Modelos generados exitosamente (%d tablas) en: %s\n", len(tables), outDir)
	return nil
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
