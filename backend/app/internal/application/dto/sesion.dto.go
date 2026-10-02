// Package dto contains data transfer objects for API contracts.
package dto

// UsuarioSesionDTO represents a session user returned in API responses.
type UsuarioSesionDTO struct {
	ID                  int64  `json:"id"`
	Usuario             string `json:"usuario"`
	Nombre              string `json:"nombre"`
	Rol                 string `json:"rol"`
	RolNombre           string `json:"rol_nombre,omitempty"`
	CapatazID           string `json:"capataz_id,omitempty"`
	DebeCambiarPassword bool   `json:"debe_cambiar_password,omitempty"`
}

// PermisoDTO represents a permission item in API responses.
type PermisoDTO struct {
	Rol    string `json:"rol"`
	Accion string `json:"accion"`
}

// CuentaDTO is an account row for the jefatura. It never includes the password hash.
type CuentaDTO struct {
	ID                  int64  `json:"id"`
	Usuario             string `json:"usuario"`
	Nombre              string `json:"nombre"`
	Rol                 string `json:"rol"`
	RolNombre           string `json:"rol_nombre,omitempty"`
	CapatazID           string `json:"capataz_id,omitempty"`
	Activo              bool   `json:"activo"`
	DebeCambiarPassword bool   `json:"debe_cambiar_password"`
}

// RolDTO is a configurable role. The code stays stable.
type RolDTO struct {
	Codigo string `json:"codigo"`
	Nombre string `json:"nombre"`
	Activo bool   `json:"activo"`
}

// CrearCuentaDTO is the payload to open an account with its own password.
type CrearCuentaDTO struct {
	Usuario   string
	Nombre    string
	Rol       string
	Clave     string
	CapatazID string
}

// ActualizarCuentaDTO patches role, name, active flag or password. Nil fields stay.
type ActualizarCuentaDTO struct {
	Nombre    *string
	Rol       *string
	Activo    *bool
	Clave     *string
	CapatazID *string
}

// UsuariosResponseDTO represents the GET /accesos/usuarios response payload.
type UsuariosResponseDTO struct {
	Usuarios []CuentaDTO  `json:"usuarios"`
	Permisos []PermisoDTO `json:"permisos"`
	Roles    []RolDTO     `json:"roles"`
	Aviso    string       `json:"aviso"`
}
