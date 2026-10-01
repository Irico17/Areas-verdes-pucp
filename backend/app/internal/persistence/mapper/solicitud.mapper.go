// Package mapper translates between database models and domain entities.
package mapper

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// SolicitudToEntity maps SolicitudModel to entities.Solicitud.
func SolicitudToEntity(m *models.SolicitudModel) *entities.Solicitud {
	if m == nil {
		return nil
	}
	return &entities.Solicitud{
		ID:            m.ID,
		CodigoExterno: m.CodigoExterno,
		Fuente:        m.Fuente,
		Titulo:        m.Titulo,
		Detalle:       m.Detalle,
		Prioridad:     m.Prioridad,
		Estado:        m.Estado,
		Lugar:         m.Lugar,
		Cantidad:      m.Cantidad,
		ActividadID:   m.ActividadID,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
		OrigenRef:     m.OrigenRef,
		ArchivadaEn:   m.ArchivadaEn,
	}
}

// SolicitudToModel maps entities.Solicitud to SolicitudModel.
func SolicitudToModel(e *entities.Solicitud) *models.SolicitudModel {
	if e == nil {
		return nil
	}
	return &models.SolicitudModel{
		ID:            e.ID,
		CodigoExterno: e.CodigoExterno,
		Fuente:        e.Fuente,
		Titulo:        e.Titulo,
		Detalle:       e.Detalle,
		Prioridad:     e.Prioridad,
		Estado:        e.Estado,
		Lugar:         e.Lugar,
		Cantidad:      e.Cantidad,
		ActividadID:   e.ActividadID,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
		OrigenRef:     e.OrigenRef,
		ArchivadaEn:   e.ArchivadaEn,
	}
}
