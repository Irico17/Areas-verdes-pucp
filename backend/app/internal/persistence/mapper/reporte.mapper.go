// Package mapper converts between persistence models and domain entities.
package mapper

import (
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// ToDomainFilaReporte converts a persistence FilaReporteModel to a domain FilaReporte.
func ToDomainFilaReporte(m models.FilaReporteModel) entities.FilaReporte {
	f := entities.FilaReporte{
		ID:            m.ID,
		Titulo:        m.Titulo,
		Tipo:          m.Tipo,
		Estado:        m.Estado,
		Ejecutor:      m.Ejecutor,
		Equipo:        m.Equipo,
		Zona:          m.Zona,
		CodigoExterno: m.CodigoExterno,
		Fuente:        m.Fuente,
		CreatedAt:     m.CreatedAt.UTC().Format(time.RFC3339),
		Clase:         m.Clase,
		Lugar:         m.Lugar,
		Cuadrilla:     m.Cuadrilla,
	}
	if m.FechaSolicitud.Valid {
		f.FechaSolicitud = m.FechaSolicitud.Time.Format("2006-01-02")
	}
	if m.FechaAtencion.Valid {
		f.FechaAtencion = m.FechaAtencion.Time.Format("2006-01-02")
	}
	return f
}

// ToDomainConteoReporte converts a persistence ConteoReporteModel to a domain ConteoReporte.
func ToDomainConteoReporte(m models.ConteoReporteModel) entities.ConteoReporte {
	return entities.ConteoReporte{
		Estado: m.Estado,
		N:      m.N,
	}
}
