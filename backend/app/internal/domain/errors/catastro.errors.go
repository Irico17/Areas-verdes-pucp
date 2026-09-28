// Package errors defines domain error types and sentinel values.
package errors

import "errors"

// Sentinel errors for catastro maestro.
var (
	// ErrNoEncontrado indicates that a record does not exist or has been deactivated.
	ErrNoEncontrado = errors.New("no encontrado")
)
