// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

// ICargaLoteUseCase defines use case operations for running etl-lote.
type ICargaLoteUseCase interface {
	// Ejecutar executes the batch loading pipeline or reports sources in read-only mode.
	Ejecutar(ctx context.Context, opt dto.OpcionesCargaLoteDTO) (*dto.ReporteLoteDTO, error)
}
