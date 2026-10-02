package database

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

var reSemillaDestructiva = regexp.MustCompile(`(?i)\b(?:DROP\s+TABLE|DROP\s+COLUMN|TRUNCATE|DELETE\s+FROM)\b`)

// AplicarSemillaFicticia inserts the fictional develop/qa seed only when the
// cadastre tables are empty. It never deletes or updates loaded rows.
func AplicarSemillaFicticia(gdb *gorm.DB, path string) error {
	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}
	necesita, conDatos, err := ComprobarNecesitaETL(sqlDB)
	if err != nil {
		return err
	}
	if !necesita {
		fmt.Printf("BD con datos (%s): no se aplica la semilla ficticia\n", strings.Join(conDatos, ", "))
		return nil
	}

	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("leer semilla ficticia %s: %w", path, err)
	}
	if reSemillaDestructiva.Match(body) {
		return fmt.Errorf("semilla ficticia %s contiene DROP, TRUNCATE o DELETE FROM", path)
	}

	stmts := splitSQL(string(body))
	if len(stmts) == 0 {
		return fmt.Errorf("semilla ficticia %s no tiene sentencias", path)
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		return err
	}
	for i, stmt := range stmts {
		if _, err := tx.Exec(stmt); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("semilla ficticia sentencia %d: %w", i+1, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Printf("semilla ficticia aplicada (%d sentencias) desde %s\n", len(stmts), path)
	return nil
}

// SemillaFicticiaEsSegura reports whether path exists and has no destructive SQL.
func SemillaFicticiaEsSegura(path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if reSemillaDestructiva.Match(body) {
		return fmt.Errorf("semilla destructiva: %s", path)
	}
	if len(splitSQL(string(body))) == 0 {
		return fmt.Errorf("semilla vacía: %s", path)
	}
	return nil
}
