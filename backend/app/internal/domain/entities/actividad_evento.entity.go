// Package entities defines core domain entities.
package entities

import "time"

// ActividadEvento represents an event log row in the timeline of an activity.
type ActividadEvento struct {
	ID          int64
	ActividadID string
	Tipo        string
	Estado      *string
	CapatazID   *string
	Equipo      *string
	ActorRol    string
	UsuarioID   *int64
	Usuario     string
	Nombre      string
	Nota        string
	CreatedAt   time.Time
}
