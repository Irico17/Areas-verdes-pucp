// Package mappers converts between persistence models and domain entities.
package mappers

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// EvidenciaModelToEntity converts an EvidenciaModel to a domain Evidencia entity.
func EvidenciaModelToEntity(m *models.EvidenciaModel) entities.Evidencia {
	if m == nil {
		return entities.Evidencia{}
	}
	actID := ""
	if m.ActividadID != nil {
		actID = *m.ActividadID
	}
	return entities.Evidencia{
		ID:          m.ID,
		ActividadID: actID,
		SolicitudID: m.SolicitudID,
		OrdenID:     m.OrdenID,
		Nombre:      m.Nombre,
		Mime:        m.Mime,
		Bytes:       m.Bytes,
		Ruta:        m.Ruta,
		Nota:        m.Nota,
		SHA256:      m.SHA256,
		Lat:         m.Lat,
		Lon:         m.Lon,
		Exif:        m.Exif,
		CreatedAt:   m.CreatedAt,
	}
}

// EvidenciaEntityToModel converts a domain Evidencia entity to a persistence EvidenciaModel.
func EvidenciaEntityToModel(e entities.Evidencia) *models.EvidenciaModel {
	var actID *string
	if e.ActividadID != "" {
		actID = &e.ActividadID
	}
	return &models.EvidenciaModel{
		ID:          e.ID,
		ActividadID: actID,
		SolicitudID: e.SolicitudID,
		OrdenID:     e.OrdenID,
		Nombre:      e.Nombre,
		Mime:        e.Mime,
		Bytes:       e.Bytes,
		Ruta:        e.Ruta,
		Nota:        e.Nota,
		CreatedAt:   e.CreatedAt,
		SHA256:      e.SHA256,
		Lat:         e.Lat,
		Lon:         e.Lon,
		Exif:        e.Exif,
	}
}
