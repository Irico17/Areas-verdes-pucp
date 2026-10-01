// Package mapper translates between database models and domain entities.
package mapper

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// OrdenServicioToEntity maps OrdenServicioModel to entities.ServicioTercerizado.
func OrdenServicioToEntity(m *models.OrdenServicioModel) *entities.ServicioTercerizado {
	if m == nil {
		return nil
	}
	return &entities.ServicioTercerizado{
		ID:               m.ID,
		ActividadID:      m.ActividadID,
		Empresa:          m.Empresa,
		Referencia:       m.Referencia,
		Frecuencia:       m.Frecuencia,
		Estado:           m.Estado,
		Conformidad:      m.Conformidad,
		CreatedAt:        m.CreatedAt,
		PeriodoInicio:    m.PeriodoInicio,
		PeriodoFin:       m.PeriodoFin,
		ReporteProveedor: m.ReporteProveedor,
	}
}

// OrdenServicioToModel maps entities.ServicioTercerizado to OrdenServicioModel.
func OrdenServicioToModel(e *entities.ServicioTercerizado) *models.OrdenServicioModel {
	if e == nil {
		return nil
	}
	return &models.OrdenServicioModel{
		ID:               e.ID,
		ActividadID:      e.ActividadID,
		Empresa:          e.Empresa,
		Referencia:       e.Referencia,
		Frecuencia:       e.Frecuencia,
		Estado:           e.Estado,
		Conformidad:      e.Conformidad,
		CreatedAt:        e.CreatedAt,
		PeriodoInicio:    e.PeriodoInicio,
		PeriodoFin:       e.PeriodoFin,
		ReporteProveedor: e.ReporteProveedor,
	}
}
