package requests

// CrearUsuarioRequest is the body of POST /accesos/usuarios.
type CrearUsuarioRequest struct {
	Usuario   string `json:"usuario"`
	Nombre    string `json:"nombre"`
	Rol       string `json:"rol"`
	Clave     string `json:"clave"`
	CapatazID string `json:"capataz_id"`
}

// ActualizarUsuarioRequest is the body of PATCH /accesos/usuarios/:usuario.
// Nil fields are left unchanged.
type ActualizarUsuarioRequest struct {
	Nombre    *string `json:"nombre"`
	Rol       *string `json:"rol"`
	Activo    *bool   `json:"activo"`
	Clave     *string `json:"clave"`
	CapatazID *string `json:"capataz_id"`
}

// CambiarClaveRequest is the body of POST /sesion/clave.
type CambiarClaveRequest struct {
	ClaveActual string `json:"clave_actual"`
	ClaveNueva  string `json:"clave_nueva"`
}

// ActualizarPermisoRequest toggles one cell of the permission matrix.
type ActualizarPermisoRequest struct {
	Rol       string `json:"rol"`
	Accion    string `json:"accion"`
	Concedido bool   `json:"concedido"`
}

// CrearRolRequest opens a role without renaming existing codes.
type CrearRolRequest struct {
	Codigo string `json:"codigo"`
	Nombre string `json:"nombre"`
}

// ActualizarRolRequest activates or deactivates a role. It does not rename it.
type ActualizarRolRequest struct {
	Activo *bool `json:"activo"`
}
