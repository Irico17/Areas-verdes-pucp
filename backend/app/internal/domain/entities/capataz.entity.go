// Package entities defines core domain entities.
package entities

// Capataz represents a demonstration field team.
type Capataz struct {
	ID     string `json:"id"`
	Equipo string `json:"equipo"`
	Turno  string `json:"turno"`
	Activo bool   `json:"activo,omitempty"`
}
