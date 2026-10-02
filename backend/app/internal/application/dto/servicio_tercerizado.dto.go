// Package dto defines data transfer objects for the application layer.
package dto

// EvidenciaOrdenDTO is an evidence file already stored for a work order.
type EvidenciaOrdenDTO struct {
	ID        string `json:"id"`
	Nombre    string `json:"nombre"`
	Mime      string `json:"mime"`
	Bytes     int    `json:"bytes"`
	Nota      string `json:"nota,omitempty"`
	CreatedAt string `json:"created_at"`
}

// OrdenDTO represents a work order item returned in JSON responses.
type OrdenDTO struct {
	ID                 string              `json:"id"`
	ActividadID        string              `json:"actividad_id"`
	Empresa            string              `json:"empresa"`
	EmpresaID          *int64              `json:"empresa_id,omitempty"`
	EmpresaCatalogo    string              `json:"empresa_catalogo,omitempty"`
	EmpresaCodigo      string              `json:"empresa_codigo,omitempty"`
	Referencia         string              `json:"referencia"`
	Frecuencia         string              `json:"frecuencia,omitempty"`
	FrecuenciaID       *int64              `json:"frecuencia_id,omitempty"`
	FrecuenciaCatalogo string              `json:"frecuencia_catalogo,omitempty"`
	FrecuenciaCodigo   string              `json:"frecuencia_codigo,omitempty"`
	Estado             string              `json:"estado"`
	Conformidad        string              `json:"conformidad,omitempty"`
	CreatedAt          string              `json:"created_at"`
	Evidencias         []EvidenciaOrdenDTO `json:"evidencias"`
}

// CrearOrdenDTO represents payload to create a new work order.
type CrearOrdenDTO struct {
	ID           string `json:"id"`
	ActividadID  string `json:"actividad_id"`
	Empresa      string `json:"empresa"`
	EmpresaID    int64  `json:"empresa_id"`
	Referencia   string `json:"referencia"`
	Frecuencia   string `json:"frecuencia"`
	FrecuenciaID int64  `json:"frecuencia_id"`
	Conformidad  string `json:"conformidad"`
}

// EvidenciasOrdenDTO lists evidence already linked to one work order.
type EvidenciasOrdenDTO struct {
	OrdenID    string              `json:"orden_id"`
	Evidencias []EvidenciaOrdenDTO `json:"evidencias"`
}

// EditarOrdenDTO represents payload to edit a work order.
type EditarOrdenDTO struct {
	ID               string `json:"id"`
	Conformidad      string `json:"conformidad"`
	PeriodoInicio    string `json:"periodo_inicio"`
	PeriodoFin       string `json:"periodo_fin"`
	ReporteProveedor string `json:"reporte_proveedor"`
	Estado           string `json:"estado"`
}

// OrdenesResponseDTO wraps list of work orders.
type OrdenesResponseDTO struct {
	Aviso   string     `json:"aviso"`
	Ordenes []OrdenDTO `json:"ordenes"`
}

// EditarOrdenResponseDTO wraps updated work order and advisory notice.
type EditarOrdenResponseDTO struct {
	Orden OrdenDTO `json:"orden"`
	Aviso string   `json:"aviso"`
}
