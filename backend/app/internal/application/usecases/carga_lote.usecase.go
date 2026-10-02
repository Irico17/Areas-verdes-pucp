// Package usecases contains application business logic orchestrators.
package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

type cargaLoteUseCase struct {
	lote contracts.ICargaLoteRepository
}

// NewCargaLoteUseCase creates a new ICargaLoteUseCase instance.
func NewCargaLoteUseCase(lote contracts.ICargaLoteRepository) contracts.ICargaLoteUseCase {
	return &cargaLoteUseCase{
		lote: lote,
	}
}

// Ejecutar executes the batch loading pipeline or reports sources in read-only mode.
func (u *cargaLoteUseCase) Ejecutar(ctx context.Context, opt dto.OpcionesCargaLoteDTO) (*dto.ReporteLoteDTO, error) {
	rep, err := u.lote.CargarLoteCompleto(ctx, opt.RawDir, opt.SoloLectura)
	if err != nil {
		return nil, err
	}

	rechDTO := make([]dto.RechazoDTO, len(rep.Rechazados))
	for i, r := range rep.Rechazados {
		rechDTO[i] = dto.RechazoDTO{
			Fuente: r.Fuente,
			Fila:   r.Fila,
			Campo:  r.Campo,
			Motivo: r.Motivo,
		}
	}

	return &dto.ReporteLoteDTO{
		LoteID:     rep.LoteID,
		Origen:     rep.Origen,
		Cargados:   rep.Cargados,
		Rechazados: rechDTO,
		Avisos:     rep.Avisos,
	}, nil
}
