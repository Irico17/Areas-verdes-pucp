package errors

import "errors"

var (
	// ErrClaseNoReconocida is returned when the catalog class is not recognized.
	ErrClaseNoReconocida = errors.New("clase de catálogo no reconocida")
	// ErrCodigoInvalido is returned when the catalog code does not match the required pattern.
	ErrCodigoInvalido = errors.New("el código usa minúsculas, números y guion bajo")
	// ErrNombreObligatorio is returned when the catalog name is empty or longer than 80 runes.
	ErrNombreObligatorio = errors.New("el nombre es obligatorio y de hasta 80 caracteres")
	// ErrItemNoExiste is returned when deactivating a non-existent catalog item.
	ErrItemNoExiste = errors.New("no existe ese ítem")
	// ErrIDInvalido is returned when a catalog item ID cannot be parsed.
	ErrIDInvalido = errors.New("id inválido")
)
