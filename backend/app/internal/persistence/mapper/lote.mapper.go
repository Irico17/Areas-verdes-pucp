// Package mapper provides domain-to-model and model-to-domain mappings.
package mapper

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// LoteImportacionModelToEntity converts a GORM model to domain entity.
func LoteImportacionModelToEntity(m models.LoteImportacionModel) entities.LoteImportacion {
	return entities.LoteImportacion{
		ID:            m.ID,
		Entidad:       m.Entidad,
		Estado:        m.Estado,
		UsuarioID:     m.UsuarioID,
		Filas:         m.Filas,
		Contenido:     m.Contenido,
		NombreArchivo: m.NombreArchivo,
		CreatedAt:     m.CreatedAt,
		RevertidoEn:   m.RevertidoEn,
	}
}

// LoteImportacionEntityToModel converts a domain entity to GORM model.
func LoteImportacionEntityToModel(e entities.LoteImportacion) models.LoteImportacionModel {
	return models.LoteImportacionModel{
		ID:            e.ID,
		Entidad:       e.Entidad,
		Estado:        e.Estado,
		UsuarioID:     e.UsuarioID,
		Filas:         e.Filas,
		Contenido:     e.Contenido,
		NombreArchivo: e.NombreArchivo,
		CreatedAt:     e.CreatedAt,
		RevertidoEn:   e.RevertidoEn,
	}
}
