// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ICargaInventarioRepository defines database operations for loading inventario layers.
type ICargaInventarioRepository interface {
	// LoadInventario replaces the inventario table with parsed records if empty.
	LoadInventario(ctx context.Context, capas map[string][]entities.InventarioRecord) error
}
