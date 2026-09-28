package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IEspecieRepository defines data access for botanical species.
type IEspecieRepository interface {
	Listar(ctx context.Context) ([]entities.Especie, error)
	Crear(ctx context.Context, cientifico, comun string) (entities.Especie, error)
}

// IEspecieUseCase defines business logic for botanical species.
type IEspecieUseCase interface {
	Listar(ctx context.Context) ([]dto.EspecieDTO, error)
	Crear(ctx context.Context, req dto.CrearEspecieDTO) (dto.EspecieDTO, error)
}
