// Package errors defines domain and business errors.
package errors

import "errors"

var (
	// ErrEntidad is returned when the requested entity is not importable.
	ErrEntidad = errors.New("entidad no importable")

	// ErrSinValidas is returned when an uploaded file has no valid rows.
	ErrSinValidas = errors.New("ninguna fila válida")

	// ErrLote is returned when the batch is not in preview state.
	ErrLote = errors.New("lote no confirmable")

	// ErrLoteNoConfirmable is an alias for ErrLote.
	ErrLoteNoConfirmable = ErrLote
)
