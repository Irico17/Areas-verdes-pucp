// Package usecases contains application business logic orchestrators.
package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

type cargaInicialUseCase struct {
	load contracts.ICargaLoadRepository
}

// NewCargaInicialUseCase creates a new ICargaInicialUseCase instance.
func NewCargaInicialUseCase(load contracts.ICargaLoadRepository) contracts.ICargaInicialUseCase {
	return &cargaInicialUseCase{
		load: load,
	}
}

// Ejecutar executes raw to v1 normalization and optional initial database load.
func (u *cargaInicialUseCase) Ejecutar(ctx context.Context, opt dto.OpcionesETLDTO) (*dto.ReporteETLDTO, error) {
	rep, err := u.load.EjecutarETL(ctx, opt.RawDir, opt.V1Dir, opt.SkipLoad, opt.Strict)
	if err != nil {
		return nil, err
	}

	return &dto.ReporteETLDTO{
		Areas:      rep.Areas,
		Zonas:      rep.Zonas,
		Capas:      rep.Capas,
		Inventario: rep.Inventario,
	}, nil
}
