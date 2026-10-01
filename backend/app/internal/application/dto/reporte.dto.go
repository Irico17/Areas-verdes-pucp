// Package dto contains data transfer objects for presentation and application layers.
package dto

// FilaReporteDTO represents a single activity row serialized for JSON responses.
type FilaReporteDTO struct {
	ID             string `json:"id"`
	Titulo         string `json:"titulo"`
	Tipo           string `json:"tipo"`
	Estado         string `json:"estado"`
	Ejecutor       string `json:"ejecutor"`
	Equipo         string `json:"equipo,omitempty"`
	Zona           string `json:"zona,omitempty"`
	CodigoExterno  string `json:"codigo_externo,omitempty"`
	Fuente         string `json:"fuente,omitempty"`
	CreatedAt      string `json:"created_at"`
	Clase          string `json:"clase,omitempty"`
	Lugar          string `json:"lugar,omitempty"`
	Cuadrilla      string `json:"cuadrilla,omitempty"`
	FechaSolicitud string `json:"fecha_solicitud,omitempty"`
	FechaAtencion  string `json:"fecha_atencion,omitempty"`
}

// ConteoReporteDTO represents activity count aggregated by status.
type ConteoReporteDTO struct {
	Estado string `json:"estado"`
	N      int    `json:"n"`
}

// HuecoIndicadorDTO represents an indicator without an agreed formula.
type HuecoIndicadorDTO struct {
	Clave  string `json:"clave"`
	Nombre string `json:"nombre"`
	Estado string `json:"estado"`
	Nota   string `json:"nota"`
}

// ReporteResponseDTO is the complete JSON response returned by GET /reportes/labores.
type ReporteResponseDTO struct {
	Aviso      string              `json:"aviso"`
	PorEstado  []ConteoReporteDTO  `json:"por_estado"`
	Filas      []FilaReporteDTO    `json:"filas"`
	Pendientes []HuecoIndicadorDTO `json:"pendientes"`
}

// FiltroReporteDTO represents input filter criteria for the basic report.
type FiltroReporteDTO struct {
	Estado    string
	Desde     string
	Hasta     string
	Zona      string
	Cuadrilla string
	Origen    string
}
