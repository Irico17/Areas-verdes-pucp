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
