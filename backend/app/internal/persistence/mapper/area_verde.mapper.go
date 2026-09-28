package mapper

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
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

// AreaVerdeEntityToFichaDTO converts an AreaVerde entity to FichaDTO.
func AreaVerdeEntityToFichaDTO(e *entities.AreaVerde, conGeom bool) dto.FichaDTO {
	if e == nil {
		return dto.FichaDTO{}
	}
	var nombre, uso, riego, ref string
	if e.Nombre != nil {
		nombre = *e.Nombre
	}
	if e.Uso != nil {
		uso = *e.Uso
	}
	if e.RiegoAct != nil {
		riego = *e.RiegoAct
	}
	if e.Referencia != nil {
		ref = *e.Referencia
	}
	return dto.FichaDTO{
		FeatureID:  e.FeatureID,
		Nombre:     nombre,
		Uso:        uso,
		RiegoAct:   riego,
		Referencia: ref,
		AreaM2:     e.AreaM2,
		ConGeom:    conGeom,
	}
}
