// Package mapper translates between database models and domain entities.
package mapper

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// RiegoRegistroToEntity maps RiegoRegistroModel and joined fields to entities.TurnoRiego.
func RiegoRegistroToEntity(m *models.RiegoRegistroModel, capatazEquipo, zonaCodigo *string) *entities.TurnoRiego {
	if m == nil {
		return nil
	}
	return &entities.TurnoRiego{
		ID:                  m.ID,
		Sector:              m.Sector,
		Turno:               m.Turno,
		CapatazID:           m.CapatazID,
		Equipo:              capatazEquipo,
		Fecha:               m.Fecha.Format("2006-01-02"),
		Nota:                m.Nota,
		ZonaSupervisionID:   m.ZonaSupervisionID,
		ZonaSupervisionCode: zonaCodigo,
		Ciclo:               m.Ciclo,
		SuperficieM2:        m.SuperficieM2,
		CreatedAt:           m.CreatedAt,
	}
}
