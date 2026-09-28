package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type cuadrillaUseCase struct {
	repo contracts.ICuadrillaRepository
}

// NewCuadrillaUseCase creates a new ICuadrillaUseCase instance.
func NewCuadrillaUseCase(repo contracts.ICuadrillaRepository) contracts.ICuadrillaUseCase {
	return &cuadrillaUseCase{repo: repo}
}

func (uc *cuadrillaUseCase) Listar(ctx context.Context) ([]dto.CuadrillaDTO, error) {
	return uc.repo.Listar(ctx)
}

func (uc *cuadrillaUseCase) Crear(ctx context.Context, req dto.CrearCuadrillaDTO) (dto.CuadrillaDTO, error) {
	c := entities.Cuadrilla{
		ID:             req.ID,
		NombreFicticio: req.Nombre,
		Turno:          req.Turno,
	}
	if err := c.Validar(); err != nil {
		return dto.CuadrillaDTO{}, err
	}
	return uc.repo.Crear(ctx, c.ID, c.NombreFicticio, c.Turno)
}
