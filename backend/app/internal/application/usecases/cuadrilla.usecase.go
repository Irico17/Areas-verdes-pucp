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
	rows, err := uc.repo.Listar(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.CuadrillaDTO, len(rows))
	for i, c := range rows {
		out[i] = cuadrillaEntityToDTO(c)
	}
	return out, nil
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
	res, err := uc.repo.Crear(ctx, c.ID, c.NombreFicticio, c.Turno)
	if err != nil {
		return dto.CuadrillaDTO{}, err
	}
	return cuadrillaEntityToDTO(res), nil
}

func cuadrillaEntityToDTO(c entities.Cuadrilla) dto.CuadrillaDTO {
	return dto.CuadrillaDTO{
		ID:             c.ID,
		NombreFicticio: c.NombreFicticio,
		Turno:          c.Turno,
		Activo:         c.Activo,
	}
}
