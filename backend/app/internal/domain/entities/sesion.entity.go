package entities

import "time"

// Sesion represents an active opaque session token record.
type Sesion struct {
	TokenHash string
	UsuarioID int64
	ExpiresAt time.Time
}
