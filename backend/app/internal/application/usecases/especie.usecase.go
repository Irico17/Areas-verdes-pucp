package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type especieUseCase struct {
	repo contracts.IEspecieRepository
}

// NewEspecieUseCase creates a new IEspecieUseCase instance.
func NewEspecieUseCase(repo contracts.IEspecieRepository) contracts.IEspecieUseCase {
	return &especieUseCase{repo: repo}
}

func (uc *especieUseCase) Listar(ctx context.Context) ([]dto.EspecieDTO, error) {
	rows, err := uc.repo.Listar(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.EspecieDTO, len(rows))
	for i, e := range rows {
		out[i] = especieEntityToDTO(e)
	}
	return out, nil
}

func (uc *especieUseCase) Crear(ctx context.Context, req dto.CrearEspecieDTO) (dto.EspecieDTO, error) {
	e := entities.Especie{
		NombreCientifico: req.Cientifico,
		NombreComun:      req.Comun,
	}
	if err := e.Validar(); err != nil {
		return dto.EspecieDTO{}, err
	}
	res, err := uc.repo.Crear(ctx, e.NombreCientifico, e.NombreComun)
	if err != nil {
		return dto.EspecieDTO{}, err
	}
	return especieEntityToDTO(res), nil
}

func especieEntityToDTO(e entities.Especie) dto.EspecieDTO {
	return dto.EspecieDTO{
		ID:               e.ID,
		NombreCientifico: e.NombreCientifico,
		NombreComun:      e.NombreComun,
		Activo:           e.Activo,
	}
}
