package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ICuadrillaRepository defines data access for operational work teams.
type ICuadrillaRepository interface {
	Listar(ctx context.Context) ([]entities.Cuadrilla, error)
	Crear(ctx context.Context, id, nombre, turno string) (entities.Cuadrilla, error)
}

// ICuadrillaUseCase defines business logic for operational work teams.
type ICuadrillaUseCase interface {
	Listar(ctx context.Context) ([]dto.CuadrillaDTO, error)
	Crear(ctx context.Context, req dto.CrearCuadrillaDTO) (dto.CuadrillaDTO, error)
}
