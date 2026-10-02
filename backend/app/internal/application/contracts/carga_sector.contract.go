// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ICargaSectorRepository defines database operations for polygon sector assignment.
type ICargaSectorRepository interface {
	// AplicarSectores populates the sector column for polygons from reference data where NULL.
	AplicarSectores(ctx context.Context) (int64, error)

	// GenerarSectores generates sector groupings from jefe_de_grupo.json and writes JSON or returns SQL.
	GenerarSectores(ctx context.Context, rawDir, v1Dir, salida string, imprimirSQL, vivo bool) (*entities.ArchivoSectores, string, error)
}
