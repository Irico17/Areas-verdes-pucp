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
	// ErrUsuarioYaExiste is returned when the login name is already taken.
	ErrUsuarioYaExiste = errors.New("ya existe una cuenta con ese usuario")
	// ErrCuentaNoEncontrada is returned when the account does not exist.
	ErrCuentaNoEncontrada = errors.New("no existe esa cuenta")
	// ErrClaveInvalida is returned when a password is missing or too short.
	ErrClaveInvalida = errors.New("la clave debe tener al menos 10 caracteres")
	// ErrNombreInvalido is returned when the display name is empty.
	ErrNombreInvalido = errors.New("indique el nombre de la persona")
	// ErrUsuarioInvalido is returned when the login name has an invalid shape.
	ErrUsuarioInvalido = errors.New("el usuario solo admite minúsculas, números, punto, guion y guion bajo")
	// ErrRolNoDisponible is returned when the role is missing or inactive.
	ErrRolNoDisponible = errors.New("ese rol no está disponible")
	// ErrNoBajaPropia is returned when someone tries to deactivate their own account.
	ErrNoBajaPropia = errors.New("no puede dar de baja su propia cuenta")
	// ErrRolYaExiste is returned when a role code is already in the catalog.
	ErrRolYaExiste = errors.New("ya existe un rol con ese código")
	// ErrCodigoRolInvalido is returned when a new role code is not a slug.
	ErrCodigoRolInvalido = errors.New("el código del rol solo admite minúsculas, números y guion bajo")
	// ErrAccionInvalida is returned when a permission action is not a slug.
	ErrAccionInvalida = errors.New("la acción del permiso no es válida")
	// ErrCuadrillaNoExiste is returned when capataz_id does not match a crew.
	ErrCuadrillaNoExiste = errors.New("esa cuadrilla no existe")
)
