// Package errors defines domain and application error types.
package errors

import (
	"errors"
)

var (
	// ErrRutaNoEncontrada indicates that the requested route does not exist.
	ErrRutaNoEncontrada = errors.New("ruta no encontrada")
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
