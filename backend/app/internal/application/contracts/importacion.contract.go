// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

// IImportacionUseCase defines use case operations for file imports.
type IImportacionUseCase interface {
	// Entidades returns the list of importable entity names.
	Entidades(ctx context.Context) []string

	// Previsualizar parses and validates an uploaded file, saving a batch preview.
	Previsualizar(ctx context.Context, entidad, nombre string, body []byte, usuarioID int64) (*dto.VistaPreviaResponseDTO, error)

	// Confirmar confirms and writes an import batch to the database.
	Confirmar(ctx context.Context, loteID, usuarioID int64) (*dto.ConfirmarImportacionResponseDTO, error)
}
