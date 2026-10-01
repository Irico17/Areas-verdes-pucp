// Package mapper translates between database models and domain entities.
package mapper

import (
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// PodaToEntity maps PodaModel to entities.Poda.
func PodaToEntity(m *models.PodaModel) *entities.Poda {
	if m == nil {
		return nil
	}
	var fRep, fEjec *string
	if m.FechaReporte != nil {
		s := m.FechaReporte.Format("2006-01-02")
		fRep = &s
	}
	if m.FechaEjecucion != nil {
		s := m.FechaEjecucion.Format("2006-01-02")
		fEjec = &s
	}
	return &entities.Poda{
		ID:                m.ID,
		Codigo:            m.Codigo,
		CodigoExterno:     m.CodigoExterno,
		Tipo:              m.Tipo,
		TipoActividad:     m.TipoActividad,
		FechaReporte:      fRep,
		FechaEjecucion:    fEjec,
		Personal:          m.PersonalFicticio,
		Ubicacion:         m.Ubicacion,
		Unidad:            m.Unidad,
		CantidadPedida:    m.CantidadPedida,
		CantidadEjecutada: m.CantidadEjecutada,
		Prioridad:         m.Prioridad,
		Comentario:        m.Comentario,
		NombreComun:       m.NombreComun,
		NombreCientifico:  m.NombreCientifico,
		ArchivadaEn:       m.ArchivadaEn,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

// PodaToModel maps entities.Poda to PodaModel.
func PodaToModel(e *entities.Poda) *models.PodaModel {
	if e == nil {
		return nil
	}
	var fRep, fEjec *time.Time
	if e.FechaReporte != nil && *e.FechaReporte != "" {
		if t, err := time.Parse("2006-01-02", *e.FechaReporte); err == nil {
			fRep = &t
		}
	}
	if e.FechaEjecucion != nil && *e.FechaEjecucion != "" {
		if t, err := time.Parse("2006-01-02", *e.FechaEjecucion); err == nil {
			fEjec = &t
		}
	}
	return &models.PodaModel{
		ID:                e.ID,
		Codigo:            e.Codigo,
		CodigoExterno:     e.CodigoExterno,
		Tipo:              e.Tipo,
		TipoActividad:     e.TipoActividad,
		FechaReporte:      fRep,
		FechaEjecucion:    fEjec,
		PersonalFicticio:  e.Personal,
		Ubicacion:         e.Ubicacion,
		Unidad:            e.Unidad,
		CantidadPedida:    e.CantidadPedida,
		CantidadEjecutada: e.CantidadEjecutada,
		Prioridad:         e.Prioridad,
		Comentario:        e.Comentario,
		NombreComun:       e.NombreComun,
		NombreCientifico:  e.NombreCientifico,
		ArchivadaEn:       e.ArchivadaEn,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}
}
