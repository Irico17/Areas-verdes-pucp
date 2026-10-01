// Package entities defines domain entities for the application.
package entities

import "time"

// ServicioTercerizado represents an outsourced service order (ordenes_servicio).
type ServicioTercerizado struct {
	ID               string
	ActividadID      string
	Empresa          string
	Referencia       string
	Frecuencia       string
	Estado           string
	Conformidad      string
	CreatedAt        time.Time
	PeriodoInicio    *time.Time
	PeriodoFin       *time.Time
	ReporteProveedor string
}

// NuevaOrdenServicio contains domain input parameters to create a new work order.
type NuevaOrdenServicio struct {
	ID          string
	ActividadID string
	Empresa     string
	Referencia  string
	Frecuencia  string
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
