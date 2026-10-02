// Package postgres provides PostgreSQL implementations of persistence contracts.
package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/etl"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
)

type cargaInventarioRepository struct {
	db *gorm.DB
}

// NewCargaInventarioRepository creates a new ICargaInventarioRepository instance.
func NewCargaInventarioRepository(db *gorm.DB) contracts.ICargaInventarioRepository {
	return &cargaInventarioRepository{db: db}
}

// LoadInventario replaces the inventario table with parsed records if empty.
func (r *cargaInventarioRepository) LoadInventario(ctx context.Context, capas map[string][]entities.InventarioRecord) error {
	db := database.DBFromContext(ctx, r.db)
	return etl.LoadInventario(db.WithContext(ctx), capas)
}
