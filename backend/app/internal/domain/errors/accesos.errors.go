package errors

import "errors"

var (
	// ErrCredencialesInvalidas is returned when user or password do not match.
	ErrCredencialesInvalidas = errors.New("usuario o clave incorrectos")
	// ErrSinSesion is returned when a session cookie is missing or invalid.
	ErrSinSesion = errors.New("sin sesión")
	// ErrSoloAdmin is returned when a non-admin tries to access admin-only data.
	ErrSoloAdmin = errors.New("solo administración ve las cuentas")
	// ErrSinPermiso is returned when a user lacks the required permission.
	ErrSinPermiso = errors.New("su rol no tiene ese permiso")
	// ErrBaseDeDatosNoDisponible is returned when database connection is unavailable.
	ErrBaseDeDatosNoDisponible = errors.New("base de datos no disponible")
)
