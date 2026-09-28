// Package errors defines domain error types and sentinel values.
package errors

import "errors"

var (
	// ErrOperacionProhibido is returned when the user's role cannot perform the action.
	ErrOperacionProhibido = errors.New("prohibido")

	// ErrLaborNoEncontrada is returned when the requested activity does not exist.
	ErrLaborNoEncontrada = errors.New("no encontrada")

	// ErrLaborConflicto is returned when an ID already exists with different payload.
	ErrLaborConflicto = errors.New("conflicto")

	// ErrValidacionOperacion marks an InputError.
	ErrValidacionOperacion = errors.New("validacion")
)

// InputError represents a client-side validation error (HTTP 400).
type InputError struct {
	Reason string
}

func (e InputError) Error() string { return e.Reason }
func (e InputError) Unwrap() error { return ErrValidacionOperacion }

// ForbiddenError represents a forbidden operation with a custom user-facing message (HTTP 403).
type ForbiddenError struct {
	Reason string
}

func (e ForbiddenError) Error() string { return e.Reason }
func (e ForbiddenError) Unwrap() error { return ErrOperacionProhibido }
