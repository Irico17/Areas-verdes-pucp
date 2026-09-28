// Package entities defines core domain entities.
package entities

import "time"

// Avance represents a progress log entry for an activity.
type Avance struct {
	ID            string
	ActividadID   string
	Fecha         string
	Nota          string
	AreaFeatureID *string
	EjemplarRef   *string
	CreatedAt     time.Time
}
