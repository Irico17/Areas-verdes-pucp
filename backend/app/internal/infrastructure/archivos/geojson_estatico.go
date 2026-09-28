package archivos

import (
	"context"
	"os"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

type geoJSONEstaticoAdapter struct {
	edificiosPath string
}

// NewGeoJSONEstaticoAdapter creates a new static GeoJSON file adapter.
func NewGeoJSONEstaticoAdapter(cfg *config.Config) contracts.IArchivoEstaticoAdapter {
	return &geoJSONEstaticoAdapter{
		edificiosPath: cfg.Datos.EdificiosPath,
	}
}

// LeerEdificios reads the static buildings extract from disk.
func (a *geoJSONEstaticoAdapter) LeerEdificios(_ context.Context) ([]byte, error) {
	return os.ReadFile(a.edificiosPath)
}
