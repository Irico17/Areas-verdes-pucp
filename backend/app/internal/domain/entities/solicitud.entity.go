// Package entities defines domain entities for the application.
package entities

import "time"

// Solicitud represents an external or internal service request / incident.
type Solicitud struct {
	ID            string
	CodigoExterno *string
	Fuente        string
	Titulo        string
	Detalle       string
	Prioridad     string
	Estado        string
	Lugar         *string
	Cantidad      *int
	ActividadID   *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	OrigenRef     *string
	ArchivadaEn   *time.Time
}
