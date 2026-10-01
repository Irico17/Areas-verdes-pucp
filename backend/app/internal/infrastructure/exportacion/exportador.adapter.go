// Package exportacion implements report exporters for CSV and Excel XML formats.
package exportacion

import (
	"context"
	"io"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type exportadorReporteAdapter struct {
	repo contracts.IReporteRepository
}

// NewExportadorReporteAdapter creates a new IExportadorReporte adapter.
func NewExportadorReporteAdapter(repo contracts.IReporteRepository) contracts.IExportadorReporte {
	return &exportadorReporteAdapter{repo: repo}
}

// ExportarCSV streams report rows in CSV format.
func (e *exportadorReporteAdapter) ExportarCSV(ctx context.Context, f entities.FiltroReporte, w io.Writer) error {
	if err := e.repo.ValidarFiltro(f); err != nil {
		return err
	}
	if err := AbrirCSV(w); err != nil {
		return err
	}
	return e.repo.RecorrerFilas(ctx, f, func(fila entities.FilaReporte) error {
		return EscribirFilaCSV(w, fila)
	})
}

// ExportarXLS streams report rows in Excel XML SpreadsheetML format.
func (e *exportadorReporteAdapter) ExportarXLS(ctx context.Context, f entities.FiltroReporte, w io.Writer) error {
	if err := e.repo.ValidarFiltro(f); err != nil {
		return err
	}
	if err := AbrirExcel(w); err != nil {
		return err
	}
	if err := e.repo.RecorrerFilas(ctx, f, func(fila entities.FilaReporte) error {
		return EscribirFilaExcel(w, fila)
	}); err != nil {
		return err
	}
	return CerrarExcel(w)
}
