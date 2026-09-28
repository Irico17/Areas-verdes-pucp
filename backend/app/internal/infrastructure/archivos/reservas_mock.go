// Package archivos provides file-system based adapters for static data.
package archivos

import (
	"context"
	"os"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

type reservasMockAdapter struct {
	path string
}

// NewReservasMockAdapter creates a new file-based mock reservations adapter.
func NewReservasMockAdapter(cfg *config.Config) contracts.IReservasMockAdapter {
	return &reservasMockAdapter{
		path: cfg.Datos.ReservasPath,
	}
}

// LeerReservas reads the mock reservations JSON file from disk.
func (a *reservasMockAdapter) LeerReservas(_ context.Context) ([]byte, error) {
	return os.ReadFile(a.path)
}
