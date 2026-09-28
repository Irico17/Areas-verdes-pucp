package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type zonaSupervisionUseCase struct {
	repo contracts.IZonaSupervisionRepository
}

// NewZonaSupervisionUseCase creates a new IZonaSupervisionUseCase instance.
func NewZonaSupervisionUseCase(repo contracts.IZonaSupervisionRepository) contracts.IZonaSupervisionUseCase {
	return &zonaSupervisionUseCase{repo: repo}
}

func (uc *zonaSupervisionUseCase) Listar(ctx context.Context) ([]dto.ZonaSupervisionDTO, error) {
	return uc.repo.Listar(ctx)
}

func (uc *zonaSupervisionUseCase) Crear(ctx context.Context, req dto.CrearZonaSupervisionDTO) (dto.ZonaSupervisionDTO, error) {
	z := entities.ZonaSupervision{
		Codigo: req.Codigo,
		Nombre: req.Nombre,
		AreaM2: req.AreaM2,
	}
	if err := z.Validar(req.GeoJSON); err != nil {
		return dto.ZonaSupervisionDTO{}, err
	}
	return uc.repo.Crear(ctx, z.Codigo, z.Nombre, req.GeoJSON, z.AreaM2)
}
