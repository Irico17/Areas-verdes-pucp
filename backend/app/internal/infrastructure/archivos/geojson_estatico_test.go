package archivos_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/archivos"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func TestGeoJSONEstaticoAdapter_LeerEdificios(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_edificios.geojson")
	content := []byte(`{"type":"FeatureCollection","name":"edificios","features":[]}`)
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Datos: config.DatosConfig{
			EdificiosPath: filePath,
		},
	}
	adapter := archivos.NewGeoJSONEstaticoAdapter(cfg)

	got, err := adapter.LeerEdificios(context.Background())
	if err != nil {
		t.Fatalf("error leyendo archivo: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("contenido inesperado: %s", string(got))
	}

	// Archivo no existe -> error
	cfgInvalido := &config.Config{
		Datos: config.DatosConfig{
			EdificiosPath: filepath.Join(tmpDir, "no_existe.geojson"),
		},
	}
	adapterInvalido := archivos.NewGeoJSONEstaticoAdapter(cfgInvalido)
	_, err = adapterInvalido.LeerEdificios(context.Background())
	if err == nil {
		t.Fatal("se esperaba error al leer archivo inexistente")
	}
}
