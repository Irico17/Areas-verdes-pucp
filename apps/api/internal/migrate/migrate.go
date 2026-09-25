package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// Apply ejecuta los .sql de dir en orden lexicográfico, una sola vez cada uno.
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

// splitSQL separa sentencias respetando bloques $$ (funciones PL/pgSQL) y comentarios --.
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

// ayudaMigracion es el texto que tiene que verse en el log cuando el contenedor
// entra en bucle: reintentar no aplica la migración que acaba de revertirse.
func ayudaMigracion(name string) string {
	return fmt.Sprintf(
		"FALLO DE MIGRACION %s: la transaccion se revirtio y esa version no quedo en schema_migrations. "+
			"Reiniciar el contenedor repite el mismo error y la API sigue en 502. "+
			"Lea el SQLSTATE de arriba, corrija el SQL o el dato que cita (sin TRUNCATE ni borrar filas) y vuelva a desplegar.",
		name,
	)
}
