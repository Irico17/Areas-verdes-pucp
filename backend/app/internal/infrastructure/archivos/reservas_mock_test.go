package archivos_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/archivos"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func TestReservasMockAdapter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "reservas.json")
	content := []byte(`{"fake": true, "reservas": []}`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Datos: config.DatosConfig{
			ReservasPath: path,
		},
	}
	adapter := archivos.NewReservasMockAdapter(cfg)

	bytes, err := adapter.LeerReservas(context.Background())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if string(bytes) != string(content) {
		t.Fatalf("contenido inesperado: %s", string(bytes))
	}
}
