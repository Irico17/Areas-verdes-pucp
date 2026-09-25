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
	if len(doc.Paths) != 50 {
		t.Fatalf("paths = %d", len(doc.Paths))
	}
	for _, key := range []string{"/health", "/api/v1/geo/areas", "/api/v1/operacion/actividades", "/api/v1/sesion", "/api/v1/ia/sugerir-tipo", "/api/v1/evidencias", "/api/v1/evidencias/{id}/archivo"} {
		if _, ok := doc.Paths[key]; !ok {
			t.Errorf("falta %s", key)
		}
	}
	raw := string(body)
	if strings.Contains(raw, "TRUNCATE") {
		t.Fatal("el contrato servido no debe mencionar TRUNCATE")
	}
}

func TestContratoRechazaPathRepetido(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "openapi.yaml")
	if err := os.WriteFile(root, []byte("openapi: 3.0.3\npaths: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	part := filepath.Join(dir, "openapi")
	if err := os.Mkdir(part, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "/api/v1/sesion:\n  post:\n    summary: uno\n"
	if err := os.WriteFile(filepath.Join(part, "a.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(part, "b.yaml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := UnirContrato(root)
	if err == nil || !strings.Contains(err.Error(), "repetido") {
		t.Fatalf("se esperaba error de path repetido, fue %v", err)
	}
}

func TestMigracionesNuevasSinTruncate(t *testing.T) {
	dir := filepath.Join("..", "..", "migrations")
	for _, name := range []string{"007_auditoria.sql", "008_fk_minimas.sql", "020_labores.sql", "021_poda.sql", "022_vivero.sql", "023_riego_zona.sql", "024_ordenes_solicitud.sql", "041_lotes.sql"} {
		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(strings.ToUpper(string(body)), "TRUNCATE") {
			t.Fatalf("%s tiene TRUNCATE", name)
		}
	}
}
