// Package entities defines core domain entities.
package entities

import "time"

// ActividadProperties represents the properties embedded in an activity GeoJSON Feature.
type ActividadProperties struct {
	ID                string  `json:"id"`
	Tipo              string  `json:"tipo"`
	Estado            string  `json:"estado"`
	Titulo            string  `json:"titulo"`
	Detalle           string  `json:"detalle,omitempty"`
	AreaFeatureID     *string `json:"area_feature_id,omitempty"`
	ZonaFeatureID     *string `json:"zona_feature_id,omitempty"`
	AssignedCapatazID *string `json:"assigned_capataz_id,omitempty"`
	Equipo            *string `json:"equipo,omitempty"`
	Ejecutor          string  `json:"ejecutor,omitempty"`
	Archivada         bool    `json:"archivada"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// Intervencion represents the core domain model of an activity / intervention.
type Intervencion struct {
	ID                string
	Tipo              string
	Estado            string
	Titulo            string
	Detalle           string
	AreaFeatureID     *string
	ZonaFeatureID     *string
	AssignedCapatazID *string
	Equipo            *string
	GeometriaGeoJSON  []byte
	ArchivadaEn       *time.Time
	Ejecutor          string
	MotivoArchivo     *string
	LugarID           *int64
	ZonaSupervisionID *int64
	FechaSolicitud    *string
	FechaAtencion     *string
	CuadrillaID       *string
	ClaseCodigo       *string
	TipoCodigo        *string
	Comentario        string
	LugarLibre        string
	OrigenRef         *string
	Origen            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
