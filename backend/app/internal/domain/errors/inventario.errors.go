// Package errors defines domain error types and sentinel values.
package errors

import "errors"

// Sentinel errors for inventario heredado and reservas mock modules.
var (
	// ErrCapaInventarioDesconocida is returned when the requested inventory layer is unknown.
	ErrCapaInventarioDesconocida = errors.New("capa de inventario desconocida")

	// ErrFotografiaNoDisponible is returned when the photo request has an invalid name or extension.
	ErrFotografiaNoDisponible = errors.New("fotografía no disponible")

	// ErrFotografiaNoRecuperada is returned when the photo file cannot be found or read from disk.
	ErrFotografiaNoRecuperada = errors.New("fotografía no recuperada")

	// ErrLeerAgendaFicticia is returned when reading or decoding the mock agenda fails.
	ErrLeerAgendaFicticia = errors.New("no se pudo leer la agenda ficticia")

	// ErrAgendaFicticiaReferenciaExterna is returned if mock agenda still references an external spreadsheet.
	ErrAgendaFicticiaReferenciaExterna = errors.New("la agenda ficticia todavía arrastra una referencia externa")
)
