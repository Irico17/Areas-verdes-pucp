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

// String returns the technical string representation of the role.
func (r Rol) String() string {
	return string(r)
}

// EsValido checks if the role is a recognized technical role.
func (r Rol) EsValido() bool {
	switch r {
	case RolCapataz, RolCoordinacion, RolJefatura, RolAdmin:
		return true
	default:
		return false
	}
}

// RolesValidos returns the list of all supported technical roles.
func RolesValidos() []Rol {
	return []Rol{
		RolCapataz,
		RolCoordinacion,
		RolJefatura,
		RolAdmin,
	}
}
