// Package usecases contains application business logic orchestrators.
package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

type sectoresUseCase struct {
	sector contracts.ICargaSectorRepository
}

// NewSectoresUseCase creates a new ISectoresUseCase instance.
func NewSectoresUseCase(sector contracts.ICargaSectorRepository) contracts.ISectoresUseCase {
	return &sectoresUseCase{sector: sector}
}

// Ejecutar executes sector generation from jefe_de_grupo.json.
func (u *sectoresUseCase) Ejecutar(ctx context.Context, opt dto.OpcionesSectoresDTO) (*dto.ResultadoSectoresDTO, error) {
	out, sqlStr, err := u.sector.GenerarSectores(ctx, opt.RawDir, opt.V1Dir, opt.Salida, opt.ImprimirSQL, opt.Vivo)
	if err != nil {
		return nil, err
	}

	if opt.ImprimirSQL {
		return &dto.ResultadoSectoresDTO{
			Desde:   out.Fuente,
			Filas:   len(out.Poligonos),
			Conteos: out.Conteos,
			SQL:     sqlStr,
		}, nil
	}

	return &dto.ResultadoSectoresDTO{
		Desde:   out.Fuente,
		Filas:   len(out.Poligonos),
		Conteos: out.Conteos,
	}, nil
}
