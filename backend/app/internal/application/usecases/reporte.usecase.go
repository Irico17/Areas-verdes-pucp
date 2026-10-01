// Package usecases contains application workflow logic.
package usecases

import (
	"context"
	"fmt"
	"io"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type reporteUseCase struct {
	repo       contracts.IReporteRepository
	exportador contracts.IExportadorReporte
}

// NewReporteUseCase creates a new IReporteUseCase instance.
func NewReporteUseCase(
	repo contracts.IReporteRepository,
	exportador contracts.IExportadorReporte,
) contracts.IReporteUseCase {
	return &reporteUseCase{
		repo:       repo,
		exportador: exportador,
	}
}

// ObtenerReporte retrieves counts, rows, and pending indicators for the given filter.
func (uc *reporteUseCase) ObtenerReporte(ctx context.Context, f dto.FiltroReporteDTO) (*dto.ReporteResponseDTO, error) {
	filtro := entities.FiltroReporte{
		Estado:    f.Estado,
		Desde:     f.Desde,
		Hasta:     f.Hasta,
		Zona:      f.Zona,
		Cuadrilla: f.Cuadrilla,
		Origen:    f.Origen,
	}
	rep, err := uc.repo.ObtenerReporte(ctx, filtro)
	if err != nil {
		return nil, err
	}

	porEstado := make([]dto.ConteoReporteDTO, 0, len(rep.PorEstado))
	for _, c := range rep.PorEstado {
		porEstado = append(porEstado, dto.ConteoReporteDTO{
			Estado: c.Estado,
			N:      c.N,
		})
	}

	filas := make([]dto.FilaReporteDTO, 0, len(rep.Filas))
	for _, row := range rep.Filas {
		filas = append(filas, dto.FilaReporteDTO{
			ID:             row.ID,
			Titulo:         row.Titulo,
			Tipo:           row.Tipo,
			Estado:         row.Estado,
			Ejecutor:       row.Ejecutor,
			Equipo:         row.Equipo,
			Zona:           row.Zona,
			CodigoExterno:  row.CodigoExterno,
			Fuente:         row.Fuente,
			CreatedAt:      row.CreatedAt,
			Clase:          row.Clase,
			Lugar:          row.Lugar,
			Cuadrilla:      row.Cuadrilla,
			FechaSolicitud: row.FechaSolicitud,
			FechaAtencion:  row.FechaAtencion,
		})
	}

	pendientes := make([]dto.HuecoIndicadorDTO, 0, len(rep.Pendientes))
	for _, p := range rep.Pendientes {
		pendientes = append(pendientes, dto.HuecoIndicadorDTO{
			Clave:  p.Clave,
			Nombre: p.Nombre,
			Estado: p.Estado,
			Nota:   p.Nota,
		})
	}

	return &dto.ReporteResponseDTO{
		Aviso:      rep.Aviso,
		PorEstado:  porEstado,
		Filas:      filas,
		Pendientes: pendientes,
	}, nil
}

// Exportar streams report rows in CSV or XLS (SpreadsheetML) format.
func (uc *reporteUseCase) Exportar(ctx context.Context, f dto.FiltroReporteDTO, formato string, w io.Writer) error {
	filtro := entities.FiltroReporte{
		Estado:    f.Estado,
		Desde:     f.Desde,
		Hasta:     f.Hasta,
		Zona:      f.Zona,
		Cuadrilla: f.Cuadrilla,
		Origen:    f.Origen,
	}
	switch formato {
	case "csv":
		return uc.exportador.ExportarCSV(ctx, filtro, w)
	case "xls":
		return uc.exportador.ExportarXLS(ctx, filtro, w)
	default:
		return fmt.Errorf("formato")
	}
}
