// Package mapper translates between database models and domain entities.
package mapper

import (
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// ViveroToEntity maps ViveroRegistroModel to entities.Vivero.
func ViveroToEntity(m *models.ViveroRegistroModel) *entities.Vivero {
	if m == nil {
		return nil
	}
	var f *string
	if m.Fecha != nil {
		s := m.Fecha.Format("2006-01-02")
		f = &s
	}
	return &entities.Vivero{
		ID:            m.ID,
		Fecha:         f,
		Area:          m.Area,
		Subproceso:    m.Subproceso,
		Etapa:         m.Etapa,
		Descripcion:   m.Descripcion,
		Observaciones: m.Observaciones,
		Responsables:  m.Responsables,
		LugarID:       m.LugarID,
		LugarLibre:    m.LugarLibre,
		ArchivadaEn:   m.ArchivadaEn,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// ViveroToModel maps entities.Vivero to ViveroRegistroModel.
func ViveroToModel(e *entities.Vivero) *models.ViveroRegistroModel {
	if e == nil {
		return nil
	}
	var f *time.Time
	if e.Fecha != nil && *e.Fecha != "" {
		if t, err := time.Parse("2006-01-02", *e.Fecha); err == nil {
			f = &t
		}
	}
	return &models.ViveroRegistroModel{
		ID:            e.ID,
		Fecha:         f,
		Area:          e.Area,
		Subproceso:    e.Subproceso,
		Etapa:         e.Etapa,
		Descripcion:   e.Descripcion,
		Observaciones: e.Observaciones,
		Responsables:  e.Responsables,
		LugarID:       e.LugarID,
		LugarLibre:    e.LugarLibre,
		ArchivadaEn:   e.ArchivadaEn,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}
