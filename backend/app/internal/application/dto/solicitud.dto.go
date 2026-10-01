// Package dto defines data transfer objects for the application layer.
package dto

// SolicitudDTO represents a service request item returned in JSON responses.
type SolicitudDTO struct {
	ID            string  `json:"id"`
	CodigoExterno *string `json:"codigo_externo,omitempty"`
	Fuente        string  `json:"fuente"`
	Titulo        string  `json:"titulo"`
	Detalle       string  `json:"detalle,omitempty"`
	Prioridad     string  `json:"prioridad"`
	Estado        string  `json:"estado"`
	Lugar         *string `json:"lugar,omitempty"`
	Cantidad      *int    `json:"cantidad,omitempty"`
	ActividadID   *string `json:"actividad_id,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

// CrearSolicitudDTO represents payload to create a new service request.
type CrearSolicitudDTO struct {
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

// EditarSolicitudDTO represents payload to patch an existing service request.
type EditarSolicitudDTO struct {
	ID            string `json:"id"`
	CodigoExterno string `json:"codigo_externo"`
	Titulo        string `json:"titulo"`
	Detalle       string `json:"detalle"`
	Prioridad     string `json:"prioridad"`
	Lugar         string `json:"lugar"`
}

// SolicitudesResponseDTO wraps list of service requests.
type SolicitudesResponseDTO struct {
	Solicitudes []SolicitudDTO `json:"solicitudes"`
}
