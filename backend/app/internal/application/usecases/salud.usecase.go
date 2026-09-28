// Package usecases contains application use cases.
package usecases

import (
	"context"
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

// ISaludUseCase defines health-checking use cases.
type ISaludUseCase interface {
	VerificarSalud(ctx context.Context) (*dto.EstadoSaludDTO, error)
}

type saludUseCase struct {
	repo contracts.ISaludRepository
}

// NewSaludUseCase creates a new salud use case.
func NewSaludUseCase(repo contracts.ISaludRepository) ISaludUseCase {
	return &saludUseCase{repo: repo}
}

// VerificarSalud queries the repository with a timeout to verify database and PostGIS health.
func (u *saludUseCase) VerificarSalud(ctx context.Context) (*dto.EstadoSaludDTO, error) {
	if u.repo == nil {
		return &dto.EstadoSaludDTO{Database: "down"}, nil
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	version, err := u.repo.PostGISVersion(ctxTimeout)
	if err != nil {
		return &dto.EstadoSaludDTO{Database: "down"}, err
	}

	return &dto.EstadoSaludDTO{
		Database: "up",
		PostGIS:  version,
	}, nil
}
