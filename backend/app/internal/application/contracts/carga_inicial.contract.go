// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

// ICargaInicialUseCase defines use case operations for running the initial ETL pipeline.
type ICargaInicialUseCase interface {
	// Ejecutar executes raw to v1 normalization and optional initial database load.
	Ejecutar(ctx context.Context, opt dto.OpcionesETLDTO) (*dto.ReporteETLDTO, error)
}
