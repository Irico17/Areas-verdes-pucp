// Package contracts defines abstract interfaces between layers.
package contracts

import (
	"context"
	"io"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IExportadorReporte defines the interface for streaming reports in CSV or Excel XML format.
type IExportadorReporte interface {
	ExportarCSV(ctx context.Context, f entities.FiltroReporte, w io.Writer) error
	ExportarXLS(ctx context.Context, f entities.FiltroReporte, w io.Writer) error
}
