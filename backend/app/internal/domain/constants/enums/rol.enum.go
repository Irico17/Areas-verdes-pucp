// Package enums defines domain enumeration types and values.
package enums

// Rol identifies technical role codes.
type Rol string

const (
	RolCapataz      Rol = "capataz"
	RolCoordinacion Rol = "coordinacion"
	RolJefatura     Rol = "jefatura"
	RolAdmin        Rol = "admin"
)
