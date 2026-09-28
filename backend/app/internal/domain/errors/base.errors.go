// Package errors defines domain and application error types.
package errors

import (
	"errors"
	"net/http"
)

var (
	// ErrNoEncontrado indicates that a requested resource was not found.
	ErrNoEncontrado = errors.New("no encontrado")
	// ErrRutaNoEncontrada indicates that the requested route does not exist.
	ErrRutaNoEncontrada = errors.New("ruta no encontrada")
	// ErrNoAutorizado indicates that authentication is required.
	ErrNoAutorizado = errors.New("no autorizado")
	// ErrProhibido indicates insufficient permissions.
	ErrProhibido = errors.New("prohibido")
	// ErrConflicto indicates a conflict with existing state.
	ErrConflicto = errors.New("conflicto")
	// ErrValidacion indicates a validation failure.
	ErrValidacion = errors.New("error de validación")
	// ErrDemasiadosIntentos indicates rate limit exceeded.
	ErrDemasiadosIntentos = errors.New("demasiados intentos de ingreso")
)

// AppError represents an application error with an HTTP status code and user-facing message.
type AppError struct {
	StatusCode int
	Message    string
	Err        error
}

// Error returns the user-facing message.
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "error"
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates an AppError.
func NewAppError(statusCode int, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewNotFoundError creates an AppError with 404 status.
func NewNotFoundError(message string) *AppError {
	if message == "" {
		message = "no encontrado"
	}
	return &AppError{
		StatusCode: http.StatusNotFound,
		Message:    message,
		Err:        ErrNoEncontrado,
	}
}
