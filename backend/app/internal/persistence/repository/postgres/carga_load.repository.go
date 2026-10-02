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

type cargaLoadRepository struct {
	db *gorm.DB
}

// NewCargaLoadRepository creates a new ICargaLoadRepository instance.
func NewCargaLoadRepository(db *gorm.DB) contracts.ICargaLoadRepository {
	return &cargaLoadRepository{db: db}
}

// EjecutarETL runs normalization and optional database loading.
func (r *cargaLoadRepository) EjecutarETL(ctx context.Context, rawDir, v1Dir string, skipLoad, strict bool) (*entities.ReporteETL, error) {
	db := database.DBFromContext(ctx, r.db)
	if !skipLoad {
		if err := r.NegarSiHayDependientes(ctx); err != nil {
			return nil, err
		}
	}

	etlOpt := etl.Options{
		RawDir:   rawDir,
		V1Dir:    v1Dir,
		SkipLoad: skipLoad,
		Strict:   strict,
	}
	if !skipLoad {
		etlOpt.DB = db
	}

	rep, err := etl.Run(etlOpt)
	if err != nil {
		return nil, err
	}

	return &entities.ReporteETL{
		Areas:      rep.Areas,
		Zonas:      rep.Zonas,
		Capas:      rep.Capas,
		Inventario: rep.Inventario,
	}, nil
}

// Load replaces the seed catastro in a transaction.
func (r *cargaLoadRepository) Load(ctx context.Context, areas, zonas []entities.CatastroRecord, capas map[string][]entities.CatastroRecord) error {
	db := database.DBFromContext(ctx, r.db)
	return etl.Load(db.WithContext(ctx), areas, zonas, capas)
}

// NegarSiHayDependientes verifies that dependent business tables are empty before Load.
func (r *cargaLoadRepository) NegarSiHayDependientes(ctx context.Context) error {
	db := database.DBFromContext(ctx, r.db)
	return etl.NegarSiHayDependientes(db.WithContext(ctx))
}
