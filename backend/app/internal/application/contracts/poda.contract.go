// Package contracts defines interfaces for application services and repositories.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IPodaRepository defines persistence operations for poda records.
type IPodaRepository interface {
	Listar(ctx context.Context) ([]*entities.Poda, error)
	Guardar(ctx context.Context, in entities.GuardarPoda) (*entities.Poda, error)
	Archivar(ctx context.Context, id string) error
}

// IPodaUseCase defines application operations for poda records.
type IPodaUseCase interface {
	Listar(ctx context.Context) (*dto.PodasResponseDTO, error)
	Crear(ctx context.Context, in dto.GuardarPodaDTO) (*dto.PodaDTO, error)
	Editar(ctx context.Context, in dto.GuardarPodaDTO) (*dto.PodaDTO, error)
	Archivar(ctx context.Context, id string) error
}
