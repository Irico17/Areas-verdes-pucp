// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ICargaCapas2BRepository defines database operations for loading auxiliary layers and areas verdes via upsert.
type ICargaCapas2BRepository interface {
	// CargarAreasVerdes upserts areas verdes from raw geojson without TRUNCATE.
	CargarAreasVerdes(ctx context.Context, rawDir string) (int, error)

	// CargarFrente2B loads tachos, bebederos, puntos, 4.13 layers, and fictional reserves.
	CargarFrente2B(ctx context.Context, rawDir string) (entities.ReporteCapas, error)
}
