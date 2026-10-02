package mapper

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// AreaVerdeModelToFicha converts an AreaVerdeModel to entities.AreaVerdeFicha.
func AreaVerdeModelToFicha(m *models.AreaVerdeModel, conGeom bool) entities.AreaVerdeFicha {
	if m == nil {
		return entities.AreaVerdeFicha{}
	}
	var nombre, uso, riego, ref string
	if m.Nombre != nil {
		nombre = *m.Nombre
	}
	if m.Uso != nil {
		uso = *m.Uso
	}
	if m.RiegoAct != nil {
		riego = *m.RiegoAct
	}
	if m.Referencia != nil {
		ref = *m.Referencia
	}
	return entities.AreaVerdeFicha{
		FeatureID:  m.FeatureID,
		Nombre:     nombre,
		Uso:        uso,
		RiegoAct:   riego,
		Referencia: ref,
		AreaM2:     m.AreaM2,
		ConGeom:    conGeom,
		Activo:     m.Activo,
	}
}
