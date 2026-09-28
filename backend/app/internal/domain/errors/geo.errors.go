// Package errors defines domain error types and sentinel values.
package errors

import "errors"

// Sentinel errors for geo and catastro modules.
var (
	// ErrCapaDesconocida is returned when the requested layer is not in the recognized catalogue.
	ErrCapaDesconocida = errors.New("capa desconocida")

	// ErrBBox indicates an invalid bbox query parameter.
	ErrBBox = errors.New("bbox debe ser minLon,minLat,maxLon,maxLat con min < max")

	// ErrLimit indicates an invalid limit query parameter.
	ErrLimit = errors.New("limit debe ser un entero entre 1 y 10000")

	// ErrFichaNoEncontrada is returned when an area verde with the given feature_id does not exist.
	ErrFichaNoEncontrada = errors.New("área no encontrada")

	// ErrEntrada indicates invalid user input that fails cadastral validation rules.
	ErrEntrada = errors.New("entrada")
)
