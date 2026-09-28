package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type lugarUseCase struct {
	repo contracts.ILugarRepository
}

// NewLugarUseCase creates a new ILugarUseCase instance.
func NewLugarUseCase(repo contracts.ILugarRepository) contracts.ILugarUseCase {
	return &lugarUseCase{repo: repo}
}

func (uc *lugarUseCase) Listar(ctx context.Context) ([]dto.LugarDTO, error) {
	return uc.repo.Listar(ctx)
}

func (uc *lugarUseCase) Crear(ctx context.Context, req dto.CrearLugarDTO) (dto.LugarDTO, error) {
	l := entities.Lugar{
		Nombre:            req.Nombre,
		Lat:               req.Lat,
		Lon:               req.Lon,
		ZonaSupervisionID: req.ZonaID,
	}
	if err := l.Validar(); err != nil {
		return dto.LugarDTO{}, err
	}
	return uc.repo.Crear(ctx, l.Nombre, l.Lat, l.Lon, l.ZonaSupervisionID)
}
