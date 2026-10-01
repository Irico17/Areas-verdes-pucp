// Package errors defines domain error types and sentinel values.
package errors

import "errors"

var (
	// ErrArchivoNoDisponible is returned when an evidence file cannot be found or read from storage.
	ErrArchivoNoDisponible = errors.New("archivo no disponible")
)
