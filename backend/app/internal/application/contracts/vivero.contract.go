// Package contracts defines interfaces for application services and repositories.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IViveroRepository defines persistence operations for nursery activity records.
type IViveroRepository interface {
	Listar(ctx context.Context, mes string) ([]*entities.Vivero, error)
	Guardar(ctx context.Context, in dto.GuardarViveroDTO) (*entities.Vivero, error)
	Archivar(ctx context.Context, id string) error
}

// IViveroUseCase defines application operations for nursery activity records.
type IViveroUseCase interface {
	Listar(ctx context.Context, mes string) (*dto.ViveroResponseDTO, error)
	Crear(ctx context.Context, in dto.GuardarViveroDTO) (*dto.ViveroDTO, error)
	Editar(ctx context.Context, in dto.GuardarViveroDTO) (*dto.ViveroDTO, error)
	Archivar(ctx context.Context, id string) error
}
