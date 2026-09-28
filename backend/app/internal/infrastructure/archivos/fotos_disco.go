// Package archivos provides file-system based adapters for static data.
package archivos

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

type fotoDiscoAdapter struct {
	fotosDir string
}

// NewFotoDiscoAdapter creates a new disk-based photo adapter.
func NewFotoDiscoAdapter(cfg *config.Config) contracts.IFotoDiscoAdapter {
	return &fotoDiscoAdapter{
		fotosDir: cfg.Datos.FotosDir,
	}
}

// RutaFoto returns the validated local filesystem path for a requested photo name.
// Path traversal and non-JPEG extensions return ErrFotografiaNoDisponible.
// Non-existent files return ErrFotografiaNoRecuperada.
func (a *fotoDiscoAdapter) RutaFoto(name string) (string, error) {
	base := filepath.Base(name)
	lower := strings.ToLower(base)
	if base == "." || base == "/" || (!strings.HasSuffix(lower, ".jpg") && !strings.HasSuffix(lower, ".jpeg")) {
		return "", domainErrors.ErrFotografiaNoDisponible
	}
	if a.fotosDir == "" {
		return "", domainErrors.ErrFotografiaNoDisponible
	}
	path := filepath.Join(a.fotosDir, base)
	if _, err := os.Stat(path); err != nil {
		return "", domainErrors.ErrFotografiaNoRecuperada
	}
	return path, nil
}
