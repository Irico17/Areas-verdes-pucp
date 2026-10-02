package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSemillaFicticiaNoBorraDatos(t *testing.T) {
	path := buscarSemilla(t)
	if err := SemillaFicticiaEsSegura(path); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	stmts := splitSQL(string(body))
	if len(stmts) < 2 {
		t.Fatalf("se esperaban al menos 2 INSERT, hay %d", len(stmts))
	}
	for _, stmt := range stmts {
		if len(stmt) < 6 || stmt[:6] != "INSERT" {
			preview := stmt
			if len(preview) > 40 {
				preview = preview[:40]
			}
			t.Fatalf("la semilla solo puede insertar, obtuvo: %s", preview)
		}
	}
}

func buscarSemilla(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		candidate := filepath.Join(dir, "deploy", "seed", "ficticio.sql")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no está deploy/seed/ficticio.sql")
		}
		dir = parent
	}
}
