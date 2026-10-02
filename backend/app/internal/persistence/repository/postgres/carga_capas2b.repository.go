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

type cargaCapas2BRepository struct {
	db *gorm.DB
}

// NewCargaCapas2BRepository creates a new ICargaCapas2BRepository instance.
func NewCargaCapas2BRepository(db *gorm.DB) contracts.ICargaCapas2BRepository {
	return &cargaCapas2BRepository{db: db}
}

// CargarAreasVerdes upserts areas verdes from raw geojson without TRUNCATE.
func (r *cargaCapas2BRepository) CargarAreasVerdes(ctx context.Context, rawDir string) (int, error) {
	db := database.DBFromContext(ctx, r.db)
	return etl.CargarAreasVerdes(db.WithContext(ctx), rawDir)
}

// CargarFrente2B loads tachos, bebederos, puntos, 4.13 layers, and fictional reserves.
func (r *cargaCapas2BRepository) CargarFrente2B(ctx context.Context, rawDir string) (entities.ReporteCapas, error) {
	db := database.DBFromContext(ctx, r.db)
	rep, err := etl.CargarFrente2B(db.WithContext(ctx), rawDir)
	if err != nil {
		return entities.ReporteCapas{}, err
	}
	rech := make([]entities.Rechazo, len(rep.Rechazados))
	for i, rc := range rep.Rechazados {
		rech[i] = entities.Rechazo{
			Fuente: rc.Fuente,
			Fila:   rc.Fila,
			Campo:  rc.Campo,
			Motivo: rc.Motivo,
		}
	}
	return entities.ReporteCapas{
		Cargados:         rep.Cargados,
		Rechazados:       rech,
		Avisos:           rep.Avisos,
		ColumnasOmitidas: rep.ColumnasOmitidas,
	}, nil
}
