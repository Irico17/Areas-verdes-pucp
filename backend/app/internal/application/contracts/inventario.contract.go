// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IInventarioRepository defines database operations for legacy inventory overlays.
type IInventarioRepository interface {
	Index(ctx context.Context) (dto.IndiceInventarioDTO, error)
	Capa(ctx context.Context, capa string) (entities.FeatureCollection, error)
}

// IInventarioUseCase defines business operations for legacy inventory overlays.
type IInventarioUseCase interface {
	Index(ctx context.Context) (dto.IndiceInventarioDTO, error)
	Capa(ctx context.Context, capa string) (entities.FeatureCollection, error)
	Foto(ctx context.Context, name string) (string, error)
}
