package errors

import "errors"

var (
	// ErrConfirmacion is returned when subsequent edits exist and confirmation is required to revert.
	ErrConfirmacion = errors.New("confirmacion")

	// ErrLoteNoEncontrado is returned when the requested batch does not exist.
	ErrLoteNoEncontrado = errors.New("lote no encontrado")

	// ErrLoteYaRevertido is returned when trying to revert an already reverted batch.
	ErrLoteYaRevertido = errors.New("el lote ya fue revertido")
)
