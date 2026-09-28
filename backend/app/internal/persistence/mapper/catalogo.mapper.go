package mapper

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// CatalogoToEntity maps CatalogoModel to domain CatalogoItem entity.
func CatalogoToEntity(m *models.CatalogoModel) *entities.CatalogoItem {
	if m == nil {
		return nil
	}
	return &entities.CatalogoItem{
		ID:     m.ID,
		Clase:  m.Clase,
		Codigo: m.Codigo,
		Nombre: m.Nombre,
		Activo: m.Activo,
		Orden:  m.Orden,
	}
}

// CatalogoToModel maps domain CatalogoItem entity to CatalogoModel.
func CatalogoToModel(e *entities.CatalogoItem) *models.CatalogoModel {
	if e == nil {
		return nil
	}
	return &models.CatalogoModel{
		ID:     e.ID,
		Clase:  e.Clase,
		Codigo: e.Codigo,
		Nombre: e.Nombre,
		Activo: e.Activo,
		Orden:  e.Orden,
	}
}
