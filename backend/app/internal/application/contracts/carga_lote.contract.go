// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ICargaLoteRepository defines database operations for loading data batches with upserts.
type ICargaLoteRepository interface {
	// CargarLote executes upsert operations for batch source files and records changes in auditoria.
	CargarLote(ctx context.Context, fuentes entities.FuentesLote) (entities.ReporteLote, error)

	// CargarLoteCompleto coordinates full batch source resolution, load, and layer updates.
	CargarLoteCompleto(ctx context.Context, rawDir string, soloLectura bool) (*entities.ReporteLote, error)

	// TextoCargado returns concatenated text columns of the loaded batch for PII verification.
	TextoCargado(ctx context.Context, loteID int64) (string, error)
}
