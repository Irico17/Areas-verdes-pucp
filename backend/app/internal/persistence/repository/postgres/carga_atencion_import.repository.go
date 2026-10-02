// Package postgres provides PostgreSQL implementations of persistence contracts.
package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/etl"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
)

type cargaAtencionImportRepository struct {
	db *gorm.DB
}

// NewCargaAtencionImportRepository creates a new ICargaAtencionImportRepository instance.
func NewCargaAtencionImportRepository(db *gorm.DB) contracts.ICargaAtencionImportRepository {
	return &cargaAtencionImportRepository{db: db}
}

// MapaLugares returns a map of normalized place name to place ID for active places.
func (r *cargaAtencionImportRepository) MapaLugares(ctx context.Context) (map[string]string, error) {
	db := database.DBFromContext(ctx, r.db)
	return etl.MapaLugares(db.WithContext(ctx))
}
