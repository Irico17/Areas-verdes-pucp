// Package entities defines domain entities for the application.
package entities

import "time"

// EvidenciaOrden is an evidence row already linked to a work order.
type EvidenciaOrden struct {
	ID        string
	Nombre    string
	Mime      string
	Bytes     int
	Nota      string
	CreatedAt time.Time
}

// ServicioTercerizado represents an outsourced service order (ordenes_servicio).
type ServicioTercerizado struct {
	ID                 string
	ActividadID        string
	Empresa            string
	EmpresaID          *int64
	EmpresaCatalogo    string
	EmpresaCodigo      string
	Referencia         string
	Frecuencia         string
	FrecuenciaID       *int64
	FrecuenciaCatalogo string
	FrecuenciaCodigo   string
	Estado             string
	Conformidad        string
	CreatedAt          time.Time
	PeriodoInicio      *time.Time
	PeriodoFin         *time.Time
	ReporteProveedor   string
	Evidencias         []EvidenciaOrden
}

// NuevaOrdenServicio contains domain input parameters to create a new work order.
type NuevaOrdenServicio struct {
	ID           string
	ActividadID  string
	Empresa      string
	EmpresaID    *int64
	Referencia   string
	Frecuencia   string
	FrecuenciaID *int64
	Conformidad  string
}

// EditarOrdenServicio contains domain input parameters to edit a work order.
type EditarOrdenServicio struct {
	ID               string
	Conformidad      string
	PeriodoInicio    string
	PeriodoFin       string
	ReporteProveedor string
	Estado           string
}
