// Package entities defines core domain entities.
package entities

import "time"

// PersonalLabor represents a member of staff assigned to a specific activity.
type PersonalLabor struct {
	ID             string
	ActividadID    string
	NombreFicticio string
	RolCampo       string
	CreatedAt      time.Time
}
