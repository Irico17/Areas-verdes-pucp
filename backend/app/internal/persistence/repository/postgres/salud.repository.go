// Package postgres implements PostgreSQL repositories.
package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
)

type saludRepository struct {
	db *gorm.DB
}

// NewSaludRepository creates a new salud repository backed by GORM.
func NewSaludRepository(db *gorm.DB) contracts.ISaludRepository {
	return &saludRepository{db: db}
}

// PostGISVersion checks PostgreSQL connection and queries PostGIS_Version().
func (r *saludRepository) PostGISVersion(ctx context.Context) (string, error) {
	if r.db == nil {
		return "", errors.New("database connection is nil")
	}

	sqlDB, err := r.db.DB()
	if err != nil {
		return "", err
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return "", err
	}

	var version string
	err = r.db.WithContext(ctx).Raw("SELECT PostGIS_Version()").Scan(&version).Error
	return version, err
}
