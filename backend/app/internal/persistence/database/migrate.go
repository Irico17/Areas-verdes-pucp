package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// TablasETL lists business tables that must be empty for the initial ETL load.
var TablasETL = []string{
	"areas_verdes",
	"poligonos_cuadrilla",
	"capas_auxiliares",
	"inventario",
	"ejemplares",
	"asignaciones_poligono",
}

// ComprobarNecesitaETL checks if the database requires the initial ETL load.
// Returns:
// - necesita=true, conDatos=nil, err=nil if all tables have 0 rows (exit code 0).
// - necesita=false, conDatos=[...], err=nil if any table has rows (exit code 10).
// - err!=nil if an error occurs querying the database.
func ComprobarNecesitaETL(sqlDB *sql.DB) (bool, []string, error) {
	var conDatos []string
	for _, tabla := range TablasETL {
		var n int
		if err := sqlDB.QueryRow(fmt.Sprintf(`SELECT count(*) FROM %s`, tabla)).Scan(&n); err != nil {
			return false, nil, fmt.Errorf("consultar conteo de %s: %w", tabla, err)
		}
		if n > 0 {
			conDatos = append(conDatos, fmt.Sprintf("%s (%d)", tabla, n))
		}
	}
	if len(conDatos) > 0 {
		return false, conDatos, nil
	}
	return true, nil, nil
}

// Apply executes .sql files in dir in lexicographical order, exactly once each.
func Apply(gdb *gorm.DB, dir string) error {
	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}

	if _, err := sqlDB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("crear schema_migrations: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("leer migraciones %s: %w", dir, err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		var n int
		if err := sqlDB.QueryRow(`SELECT count(*) FROM schema_migrations WHERE version = $1`, name).Scan(&n); err != nil {
			return fmt.Errorf("consultar %s: %w", name, err)
		}
		if n > 0 {
			fmt.Printf("migración ya aplicada: %s\n", name)
			continue
		}

		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}

		tx, err := sqlDB.Begin()
		if err != nil {
			return err
		}
		for i, stmt := range splitSQL(string(body)) {
			if _, err := tx.Exec(stmt); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("aplicar %s sentencia %d: %w\n\n%s", name, i+1, err, ayudaMigracion(name))
			}
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("registrar %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit %s: %w", name, err)
		}
		fmt.Printf("migración aplicada: %s\n", name)
	}
	return nil
}

// splitSQL splits SQL statements respecting $$ blocks (PL/pgSQL functions) and -- comments.
func splitSQL(sql string) []string {
	var stmts []string
	var b strings.Builder
	inDollar := false
	for i := 0; i < len(sql); {
		if !inDollar && strings.HasPrefix(sql[i:], "--") {
			nl := strings.IndexByte(sql[i:], '\n')
			if nl == -1 {
				break
			}
			i += nl + 1
			continue
		}
		if strings.HasPrefix(sql[i:], "$$") {
			inDollar = !inDollar
			b.WriteString("$$")
			i += 2
			continue
		}
		if sql[i] == ';' && !inDollar {
			if s := strings.TrimSpace(b.String()); s != "" {
				stmts = append(stmts, s)
			}
			b.Reset()
			i++
			continue
		}
		b.WriteByte(sql[i])
		i++
	}
	if s := strings.TrimSpace(b.String()); s != "" {
		stmts = append(stmts, s)
	}
	return stmts
}

// ayudaMigracion returns guidance logged when a migration fails and rolls back.
func ayudaMigracion(name string) string {
	return fmt.Sprintf(
		"FALLO DE MIGRACION %s: la transaccion se revirtio y esa version no quedo en schema_migrations. "+
			"Reiniciar el contenedor repite el mismo error y la API sigue en 502. "+
			"Lea el SQLSTATE de arriba, corrija el SQL o el dato que cita (sin TRUNCATE ni borrar filas) y vuelva a desplegar.",
		name,
	)
}
