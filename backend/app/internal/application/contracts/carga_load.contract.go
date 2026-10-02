// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ICargaLoadRepository defines database operations for loading the initial catastro seed.
type ICargaLoadRepository interface {
	// EjecutarETL runs normalization and optional database loading.
	EjecutarETL(ctx context.Context, rawDir, v1Dir string, skipLoad, strict bool) (*entities.ReporteETL, error)

	// Load replaces the seed catastro in a transaction.
	Load(ctx context.Context, areas, zonas []entities.CatastroRecord, capas map[string][]entities.CatastroRecord) error

	// NegarSiHayDependientes verifies that dependent business tables are empty before Load.
	NegarSiHayDependientes(ctx context.Context) error
}
