// Package contracts defines interfaces for application services and repositories.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IRiegoRepository defines persistence operations for irrigation logs (riego_registros).
type IRiegoRepository interface {
	Listar(ctx context.Context, capatazID string) ([]*entities.TurnoRiego, error)
	Crear(ctx context.Context, in entities.NuevoTurnoRiego) error
}

// IRiegoUseCase defines application operations for irrigation logs.
type IRiegoUseCase interface {
	Listar(ctx context.Context, capatazID string) (*dto.RiegoResponseDTO, error)
	Crear(ctx context.Context, in dto.CrearRiegoDTO, actorRol, capatazID string) (*dto.CrearRiegoResponseDTO, error)
}
