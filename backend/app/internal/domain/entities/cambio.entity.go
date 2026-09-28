package entities

import "time"

// Cambio represents an audited change in the system (tabla cambios).
type Cambio struct {
	ID        int64
	Entidad   string
	EntidadID string
	Accion    string
	Antes     *string
	Despues   *string
	UsuarioID *int64
	LoteID    *int64
	CreatedAt time.Time
}
