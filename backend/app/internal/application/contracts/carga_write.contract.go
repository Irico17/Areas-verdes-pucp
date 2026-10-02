// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"context"
)

// ICargaWriteRepository defines persistence operations for import previews and confirmations.
type ICargaWriteRepository interface {
	// GuardarVistaPrevia records an import batch in 'vista_previa' state.
	GuardarVistaPrevia(ctx context.Context, entidad string, usuarioID int64, validas int, contenido []byte, nombreArchivo string) (int64, error)

	// Confirmar executes the confirmation and persistence of an import preview batch.
	Confirmar(ctx context.Context, loteID, usuarioID int64) (int64, int, error)
}
