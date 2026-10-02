// Package requests defines request payload structures for HTTP controllers.
package requests

import "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"

// CrearSolicitudRequest represents HTTP body for creating a solicitud.
type CrearSolicitudRequest struct {
	ID                 string   `json:"id"`
	CodigoExterno      string   `json:"codigo_externo"`
	Fuente             string   `json:"fuente"`
	Titulo             string   `json:"titulo"`
	Detalle            string   `json:"detalle"`
	Prioridad          string   `json:"prioridad"`
	Estado             string   `json:"estado"`
	Lugar              string   `json:"lugar"`
	LugarID            *int64   `json:"lugar_id"`
	Lat                *float64 `json:"lat"`
	Lon                *float64 `json:"lon"`
	Cantidad           int      `json:"cantidad"`
	CantidadSolicitada *int     `json:"cantidad_solicitada"`
	CantidadEjecutada  *int     `json:"cantidad_ejecutada"`
	ActividadID        string   `json:"actividad_id"`
}

// ToDTO converts CrearSolicitudRequest to CrearSolicitudDTO.
func (r CrearSolicitudRequest) ToDTO() dto.CrearSolicitudDTO {
	return dto.CrearSolicitudDTO{
		ID:                 r.ID,
		CodigoExterno:      r.CodigoExterno,
		Fuente:             r.Fuente,
		Titulo:             r.Titulo,
		Detalle:            r.Detalle,
		Prioridad:          r.Prioridad,
		Estado:             r.Estado,
		Lugar:              r.Lugar,
		LugarID:            r.LugarID,
		Lat:                r.Lat,
		Lon:                r.Lon,
		Cantidad:           r.Cantidad,
		CantidadSolicitada: r.CantidadSolicitada,
		CantidadEjecutada:  r.CantidadEjecutada,
		ActividadID:        r.ActividadID,
	}
}

// EditarSolicitudRequest represents HTTP body for editing a solicitud.
type EditarSolicitudRequest struct {
	CodigoExterno      string   `json:"codigo_externo"`
	Titulo             string   `json:"titulo"`
	Detalle            string   `json:"detalle"`
	Prioridad          string   `json:"prioridad"`
	Estado             string   `json:"estado"`
	Lugar              string   `json:"lugar"`
	LugarID            *int64   `json:"lugar_id"`
	Lat                *float64 `json:"lat"`
	Lon                *float64 `json:"lon"`
	CantidadSolicitada *int     `json:"cantidad_solicitada"`
	CantidadEjecutada  *int     `json:"cantidad_ejecutada"`
}

// ToDTO converts EditarSolicitudRequest to EditarSolicitudDTO.
func (r EditarSolicitudRequest) ToDTO(id string) dto.EditarSolicitudDTO {
	return dto.EditarSolicitudDTO{
		ID:                 id,
		CodigoExterno:      r.CodigoExterno,
		Titulo:             r.Titulo,
		Detalle:            r.Detalle,
		Prioridad:          r.Prioridad,
		Estado:             r.Estado,
		Lugar:              r.Lugar,
		LugarID:            r.LugarID,
		Lat:                r.Lat,
		Lon:                r.Lon,
		CantidadSolicitada: r.CantidadSolicitada,
		CantidadEjecutada:  r.CantidadEjecutada,
	}
}
