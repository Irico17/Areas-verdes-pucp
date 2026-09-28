package dto

// RegistrarCambioDTO holds parameters to record a change in the audit trail.
type RegistrarCambioDTO struct {
	Entidad   string
	EntidadID string
	Accion    string
	Antes     any
	Despues   any
	UsuarioID *int64
	LoteID    *int64
}
