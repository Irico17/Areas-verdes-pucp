// Package dto defines data transfer objects for the application layer.
package dto

// OrdenDTO represents a work order item returned in JSON responses.
type OrdenDTO struct {
	ID          string `json:"id"`
	ActividadID string `json:"actividad_id"`
	Empresa     string `json:"empresa"`
	Referencia  string `json:"referencia"`
	Frecuencia  string `json:"frecuencia,omitempty"`
	Estado      string `json:"estado"`
	Conformidad string `json:"conformidad,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// CrearOrdenDTO represents payload to create a new work order.
type CrearOrdenDTO struct {
	ID          string `json:"id"`
	ActividadID string `json:"actividad_id"`
	Empresa     string `json:"empresa"`
	Referencia  string `json:"referencia"`
	Frecuencia  string `json:"frecuencia"`
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
