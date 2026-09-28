package mapper

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// PermisoToEntity maps PermisoModel to domain Permiso entity.
func PermisoToEntity(m *models.PermisoModel) entities.Permiso {
	if m == nil {
		return entities.Permiso{}
	}
	return entities.Permiso{
		Rol:    m.Rol,
		Accion: m.Accion,
	}
}

// PermisoToModel maps domain Permiso entity to PermisoModel.
func PermisoToModel(e entities.Permiso) models.PermisoModel {
	return models.PermisoModel{
		Rol:    e.Rol,
		Accion: e.Accion,
	}
}
