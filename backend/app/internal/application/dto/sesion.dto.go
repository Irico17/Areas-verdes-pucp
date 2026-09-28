// Package dto contains data transfer objects for API contracts.
package dto

// UsuarioSesionDTO represents a session user returned in API responses.
type UsuarioSesionDTO struct {
	ID        int64  `json:"id"`
	Usuario   string `json:"usuario"`
	Nombre    string `json:"nombre"`
	Rol       string `json:"rol"`
	RolNombre string `json:"rol_nombre,omitempty"`
	CapatazID string `json:"capataz_id,omitempty"`
}

// PermisoDTO represents a permission item in API responses.
type PermisoDTO struct {
	Rol    string `json:"rol"`
	Accion string `json:"accion"`
}

// UsuariosResponseDTO represents the GET /accesos/usuarios response payload.
type UsuariosResponseDTO struct {
	Usuarios []UsuarioSesionDTO `json:"usuarios"`
	Permisos []PermisoDTO       `json:"permisos"`
	Aviso    string             `json:"aviso"`
}
