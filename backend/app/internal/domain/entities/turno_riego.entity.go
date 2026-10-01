// Package entities defines domain entities for the application.
package entities

import "time"

// TurnoRiego represents an irrigation shift log (riego_registros).
type TurnoRiego struct {
	ID                  string
	Sector              string
	Turno               string
	CapatazID           *string
	Equipo              *string
	Fecha               string
	Nota                string
	ZonaSupervisionID   *int64
	ZonaSupervisionCode *string
	Ciclo               string
	SuperficieM2        *float64
	CreatedAt           time.Time
}
