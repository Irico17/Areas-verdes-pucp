package archivos_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/archivos"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func TestFotoDiscoAdapter(t *testing.T) {
	dir := t.TempDir()
	samplePhoto := filepath.Join(dir, "bbr63_test.jpg")
	if err := os.WriteFile(samplePhoto, []byte("fake-jpeg-content"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		Datos: config.DatosConfig{
			FotosDir: dir,
		},
	}
	adapter := archivos.NewFotoDiscoAdapter(cfg)

	// 1. Foto existente válida
	path, err := adapter.RutaFoto("bbr63_test.jpg")
	if err != nil {
		t.Fatalf("error inesperado con foto válida: %v", err)
	}
	if path != samplePhoto {
		t.Fatalf("ruta esperada %s, obtenida %s", samplePhoto, path)
	}

	// 2. Foto inexistente pero con extensión válida -> ErrFotografiaNoRecuperada
	_, errRec := adapter.RutaFoto("no_existe.jpg")
	if !errors.Is(errRec, domainErrors.ErrFotografiaNoRecuperada) {
		t.Fatalf("esperado ErrFotografiaNoRecuperada, obtenido %v", errRec)
	}

	// 3. Extensión inválida -> ErrFotografiaNoDisponible
	_, errDispExt := adapter.RutaFoto("archivo.png")
	if !errors.Is(errDispExt, domainErrors.ErrFotografiaNoDisponible) {
		t.Fatalf("esperado ErrFotografiaNoDisponible por extensión, obtenido %v", errDispExt)
	}

	// 4. Path traversal a archivo del sistema sin extensión jpg -> ErrFotografiaNoDisponible
	_, errTrav := adapter.RutaFoto("../../etc/passwd")
	if !errors.Is(errTrav, domainErrors.ErrFotografiaNoDisponible) {
		t.Fatalf("esperado ErrFotografiaNoDisponible por path traversal, obtenido %v", errTrav)
	}

	// 5. Path traversal que termine en .jpg inexistente -> ErrFotografiaNoRecuperada
	_, errTravJpg := adapter.RutaFoto("../../etc/passwd.jpg")
	if !errors.Is(errTravJpg, domainErrors.ErrFotografiaNoRecuperada) {
		t.Fatalf("esperado ErrFotografiaNoRecuperada por traversal con jpg, obtenido %v", errTravJpg)
	}

	// 6. Nombres especiales "." o "/" o vacío
	for _, badName := range []string{".", "/", ""} {
		_, errBad := adapter.RutaFoto(badName)
		if !errors.Is(errBad, domainErrors.ErrFotografiaNoDisponible) {
			t.Fatalf("esperado ErrFotografiaNoDisponible para %q, obtenido %v", badName, errBad)
		}
	}

	// 7. Directorio de fotos no configurado (vacío)
	cfgSinDir := &config.Config{
		Datos: config.DatosConfig{
			FotosDir: "",
		},
	}
	adapterSinDir := archivos.NewFotoDiscoAdapter(cfgSinDir)
	_, errVacio := adapterSinDir.RutaFoto("bbr63_test.jpg")
	if !errors.Is(errVacio, domainErrors.ErrFotografiaNoDisponible) {
		t.Fatalf("esperado ErrFotografiaNoDisponible con FotosDir vacío, obtenido %v", errVacio)
	}
}
