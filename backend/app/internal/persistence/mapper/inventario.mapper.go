package mapper

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// InventarioModelToProps maps InventarioModel to domain InventarioProperties.
func InventarioModelToProps(m *models.InventarioModel) entities.InventarioProperties {
	if m == nil {
		return entities.InventarioProperties{}
	}
	return entities.InventarioProperties{
		FeatureID: m.FeatureID,
		Capa:      m.Capa,
		Nombre:    m.Nombre,
		Subtipo:   m.Subtipo,
		Detalle:   m.Detalle,
		Lugar:     m.Lugar,
		Foto:      m.Foto,
	}
}
