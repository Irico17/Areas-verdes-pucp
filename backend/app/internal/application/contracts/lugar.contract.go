package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ILugarRepository defines data access for campus locations.
type ILugarRepository interface {
	Listar(ctx context.Context) ([]entities.Lugar, error)
	Crear(ctx context.Context, nombre string, lat, lon float64, zonaID *int64) (entities.Lugar, error)
}

// ILugarUseCase defines business logic for campus locations.
type ILugarUseCase interface {
	Listar(ctx context.Context) ([]dto.LugarDTO, error)
	Crear(ctx context.Context, req dto.CrearLugarDTO) (dto.LugarDTO, error)
}
