// Package entities defines core domain models.
package entities

// Usuario represents a local user account.
type Usuario struct {
	ID                  int64
	Usuario             string
	Nombre              string
	Rol                 string
	RolNombre           string
	CapatazID           string
	PasswordHash        string
	Activo              bool
	DebeCambiarPassword bool
	RolActivo           bool
}
