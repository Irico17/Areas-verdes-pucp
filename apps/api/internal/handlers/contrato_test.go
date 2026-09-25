package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestContratoUneLosTags(t *testing.T) {
	root := filepath.Join("..", "..", "openapi.yaml")
	body, err := UnirContrato(root)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Paths map[string]map[string]struct {
			Tags []string `yaml:"tags"`
		} `yaml:"paths"`
	}
	if err := yaml.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Paths) != 34 {
		t.Fatalf("paths = %d", len(doc.Paths))
	}
	for _, key := range []string{"/health", "/api/v1/geo/areas", "/api/v1/operacion/actividades", "/api/v1/sesion", "/api/v1/ia/sugerir-tipo"} {
		if _, ok := doc.Paths[key]; !ok {
			t.Errorf("falta %s", key)
		}
	}
	raw := string(body)
	if strings.Contains(raw, "TRUNCATE") {
		t.Fatal("el contrato servido no debe mencionar TRUNCATE")
	}
}

func TestMigracionesNuevasSinTruncate(t *testing.T) {
	dir := filepath.Join("..", "..", "migrations")
	for _, name := range []string{"007_auditoria.sql", "008_fk_minimas.sql"} {
		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(strings.ToUpper(string(body)), "TRUNCATE") {
			t.Fatalf("%s tiene TRUNCATE", name)
		}
	}
}
