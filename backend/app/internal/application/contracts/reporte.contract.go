// Package contracts defines abstract interfaces between layers.
package contracts

import (
	"context"
	"io"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IReporteRepository defines database access for generating reports and streaming rows.
type IReporteRepository interface {
	ObtenerReporte(ctx context.Context, f entities.FiltroReporte) (*entities.Reporte, error)
	ValidarFiltro(f entities.FiltroReporte) error
	RecorrerFilas(ctx context.Context, f entities.FiltroReporte, fn func(entities.FilaReporte) error) error
}

// IReporteUseCase defines the application use case for report retrieval and streaming export.
type IReporteUseCase interface {
	ObtenerReporte(ctx context.Context, f dto.FiltroReporteDTO) (*dto.ReporteResponseDTO, error)
	Exportar(ctx context.Context, f dto.FiltroReporteDTO, formato string, w io.Writer) error
}
