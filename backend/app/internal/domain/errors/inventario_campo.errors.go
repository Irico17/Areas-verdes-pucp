// Package errors defines domain error types and sentinel values.
package errors

import "errors"

// Sentinel errors for field inventory frente 2B.
var (
	// ErrPatchVacio is returned when a PATCH body is empty, null, or has no valid keys.
	ErrPatchVacio = errors.New("patch vacío")

	// ErrRegistroNoEncontrado is returned when a record is not found or already inactive.
	ErrRegistroNoEncontrado = errors.New("no se encontró el registro")

	// ErrTachoInvalido is returned when a tacho fails validation.
	ErrTachoInvalido = errors.New("tacho inválido")

	// ErrBebederoInvalido is returned when a bebedero fails validation.
	ErrBebederoInvalido = errors.New("bebedero inválido")

	// ErrPuntoInvalido is returned when a point fails validation.
	ErrPuntoInvalido = errors.New("punto inválido")

	// ErrPuntoContacto is returned when personal contact data is detected in a point payload.
	ErrPuntoContacto = errors.New("no se guardan teléfono, placeId ni website")

	// ErrReservaInvalida is returned when a reservation fails validation.
	ErrReservaInvalida = errors.New("reserva inválida")

	// ErrReservaOrigen is returned when a reservation has an origin other than 'ficticio'.
	ErrReservaOrigen = errors.New("mientras la hoja responda 401 solo se acepta origen ficticio")

	// ErrFichaInvalida is returned when a layer record fails validation.
	ErrFichaInvalida = errors.New("ficha inválida")

	// ErrCSVInvalido is returned when CSV content is malformed.
	ErrCSVInvalido = errors.New("CSV inválido")
)
