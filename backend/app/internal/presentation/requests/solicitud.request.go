// Package requests defines request payload structures for HTTP controllers.
package requests

import "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"

// CrearSolicitudRequest represents HTTP body for creating a solicitud.
type CrearSolicitudRequest struct {
	ID            string `json:"id"`
	CodigoExterno string `json:"codigo_externo"`
	Fuente        string `json:"fuente"`
	Titulo        string `json:"titulo"`
	Detalle       string `json:"detalle"`
	Prioridad     string `json:"prioridad"`
	Lugar         string `json:"lugar"`
	Cantidad      int    `json:"cantidad"`
	ActividadID   string `json:"actividad_id"`
}

// ToDTO converts CrearSolicitudRequest to CrearSolicitudDTO.
func (r CrearSolicitudRequest) ToDTO() dto.CrearSolicitudDTO {
	return dto.CrearSolicitudDTO{
		ID:            r.ID,
		CodigoExterno: r.CodigoExterno,
		Fuente:        r.Fuente,
		Titulo:        r.Titulo,
		Detalle:       r.Detalle,
		Prioridad:     r.Prioridad,
		Lugar:         r.Lugar,
		Cantidad:      r.Cantidad,
		ActividadID:   r.ActividadID,
	}
}

// EditarSolicitudRequest represents HTTP body for editing a solicitud.
type EditarSolicitudRequest struct {
	CodigoExterno string `json:"codigo_externo"`
	Titulo        string `json:"titulo"`
	Detalle       string `json:"detalle"`
	Prioridad     string `json:"prioridad"`
	Lugar         string `json:"lugar"`
}

// ToDTO converts EditarSolicitudRequest to EditarSolicitudDTO.
func (r EditarSolicitudRequest) ToDTO(id string) dto.EditarSolicitudDTO {
	return dto.EditarSolicitudDTO{
		ID:            id,
		CodigoExterno: r.CodigoExterno,
		Titulo:        r.Titulo,
		Detalle:       r.Detalle,
		Prioridad:     r.Prioridad,
		Lugar:         r.Lugar,
	}
}
