package mapper

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// AreaVerdeModelToEntity maps AreaVerdeModel to domain entity AreaVerde.
func AreaVerdeModelToEntity(m *models.AreaVerdeModel) *entities.AreaVerde {
	if m == nil {
		return nil
	}
	return &entities.AreaVerde{
		ID:          m.ID,
		FeatureID:   m.FeatureID,
		SourceIndex: m.SourceIndex,
		Codigo:      m.Codigo,
		Nombre:      m.Nombre,
		Uso:         m.Uso,
		ProyRiego:   m.ProyRiego,
		RiegoAct:    m.RiegoAct,
		Referencia:  m.Referencia,
		PerimetroM:  m.PerimetroM,
		AreaM2:      m.AreaM2,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

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
	}
}
