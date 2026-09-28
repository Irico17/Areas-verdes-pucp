// Package contracts defines application interfaces.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

// ISaludRepository provides health check operations against PostgreSQL/PostGIS.
type ISaludRepository interface {
	PostGISVersion(ctx context.Context) (string, error)
}

// ISaludUseCase defines health-checking use cases.
type ISaludUseCase interface {
	VerificarSalud(ctx context.Context) (*dto.EstadoSaludDTO, error)
}
