package dto

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// Re-export or alias entities as DTOs for catastro maestro responses.
type (
	ZonaSupervisionDTO   = entities.ZonaSupervision
	CuadrillaDTO         = entities.Cuadrilla
	LugarDTO             = entities.Lugar
	EspecieDTO           = entities.Especie
	EjemplarDTO          = entities.Ejemplar
	CodigoHistoricoDTO   = entities.CodigoHistorico
	PoligonoCuadrillaDTO = entities.PoligonoCuadrilla
	CapaFichaDTO         = entities.CapaFicha
)

// CrearZonaSupervisionDTO represents the input data to create a supervision zone.
type CrearZonaSupervisionDTO struct {
	Codigo  string   `json:"codigo"`
	Nombre  string   `json:"nombre"`
	AreaM2  *float64 `json:"area_m2"`
	GeoJSON string   `json:"geojson"`
}

// CrearCuadrillaDTO represents the input data to create a work team.
type CrearCuadrillaDTO struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre_ficticio"`
	Turno  string `json:"turno"`
}

// CrearLugarDTO represents the input data to create a place.
type CrearLugarDTO struct {
	Nombre string  `json:"nombre"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	ZonaID *int64  `json:"zona_supervision_id"`
}

// CrearEspecieDTO represents the input data to create a botanical species.
type CrearEspecieDTO struct {
	Cientifico string `json:"nombre_cientifico"`
	Comun      string `json:"nombre_comun"`
}

// EjemplaresPaginadosDTO represents paginated response for ejemplares.
type EjemplaresPaginadosDTO struct {
	Ejemplares []EjemplarDTO `json:"ejemplares"`
	Total      int           `json:"total"`
	Limit      int           `json:"limit"`
	Offset     int           `json:"offset"`
}

// RecodificarDTO represents input data to recodify a specimen.
type RecodificarDTO struct {
	Codigo string `json:"codigo_nuevo"`
}
