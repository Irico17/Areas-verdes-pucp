// Package mapper converts between persistence models and domain entities.
package mapper

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// TachoModelToEntity converts a TachoModel to a domain Tacho entity.
func TachoModelToEntity(m *models.TachoModel) *entities.Tacho {
	if m == nil {
		return nil
	}
	var nota, lugar, espacios, accion, tachoActual, tachoNuevo, recomendaciones string
	if m.Nota != nil {
		nota = *m.Nota
	}
	if m.Lugar != nil {
		lugar = *m.Lugar
	}
	if m.Espacios != nil {
		espacios = *m.Espacios
	}
	if m.Accion != nil {
		accion = *m.Accion
	}
	if m.TachoActual != nil {
		tachoActual = *m.TachoActual
	}
	if m.TachoNuevo != nil {
		tachoNuevo = *m.TachoNuevo
	}
	if m.Recomendaciones != nil {
		recomendaciones = *m.Recomendaciones
	}

	return &entities.Tacho{
		ID:                  m.ID,
		Codigo:              m.Codigo,
		Lat:                 m.Lat,
		Lon:                 m.Lon,
		Nota:                nota,
		Lugar:               lugar,
		Espacios:            espacios,
		Accion:              accion,
		TachoActual:         tachoActual,
		TachoNuevo:          tachoNuevo,
		Recomendaciones:     recomendaciones,
		NoAprovechables:     m.NoAprovechables,
		PapelCarton:         m.PapelCarton,
		Plastico:            m.Plastico,
		Vidrio:              m.Vidrio,
		Pilas:               m.Pilas,
		Peligrosos:          m.Peligrosos,
		RAEE:                m.RAEE,
		Metales:             m.Metales,
		Aniquem:             m.Aniquem,
		IntermediosPlastico: m.IntermediosPlastico,
		IntermediosMetal:    m.IntermediosMetal,
		Activo:              m.Activo,
	}
}

// BebederoModelToEntity converts a BebederoModel to a domain Bebedero entity.
func BebederoModelToEntity(m *models.BebederoModel) *entities.Bebedero {
	if m == nil {
		return nil
	}
	var sede string
	if m.Sede != nil {
		sede = *m.Sede
	}
	return &entities.Bebedero{
		ID:      m.ID,
		Codigo:  m.Codigo,
		Subtipo: m.Subtipo,
		Estado:  m.Estado,
		Sede:    sede,
		Lat:     m.Lat,
		Lon:     m.Lon,
		Activo:  m.Activo,
	}
}

// PuntoPUCPModelToEntity converts a PuntoPUCPModel to a domain PuntoPUCP entity.
func PuntoPUCPModelToEntity(m *models.PuntoPUCPModel) *entities.PuntoPUCP {
	if m == nil {
		return nil
	}
	var url string
	if m.URL != nil {
		url = *m.URL
	}
	return &entities.PuntoPUCP{
		ID:     m.ID,
		Titulo: m.Titulo,
		Lat:    m.Lat,
		Lon:    m.Lon,
		URL:    url,
		Activo: m.Activo,
	}
}

// FichaCapaModelToEntity converts a FichaCapaModel to a domain FichaCapa entity.
func FichaCapaModelToEntity(m *models.FichaCapaModel) *entities.FichaCapa {
	if m == nil {
		return nil
	}
	return &entities.FichaCapa{
		ID:         m.ID,
		FeatureID:  m.FeatureID,
		Nombre:     m.Nombre,
		Codigo:     m.Codigo,
		Nota:       m.Nota,
		Clase:      m.Clase,
		Riego:      m.Riego,
		AreaM2:     m.AreaM2,
		PerimetroM: m.PerimetroM,
		Pertenecen: m.Pertenecen,
		Uso:        m.Uso,
		GeoJSON:    m.GeoJSON,
		Activo:     m.Activo,
	}
}
