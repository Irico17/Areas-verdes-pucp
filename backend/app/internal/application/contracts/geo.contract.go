package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IGeoRepository defines database reads for cadastral and auxiliary geometries.
type IGeoRepository interface {
	Areas(ctx context.Context, f entities.FiltroGeo) (entities.FeatureCollection, error)
	Zonas(ctx context.Context, f entities.FiltroGeo) (entities.FeatureCollection, error)
	Capa(ctx context.Context, capa string, f entities.FiltroGeo) (entities.FeatureCollection, error)
	Capas(ctx context.Context) (entities.CapasIndex, error)
	Resumen(ctx context.Context) (entities.ResumenCatastro, error)
}

// IGeoUseCase defines use case operations for reading cadastral geometries.
type IGeoUseCase interface {
	Areas(ctx context.Context, f dto.FiltroGeoDTO) (entities.FeatureCollection, error)
	Zonas(ctx context.Context, f dto.FiltroGeoDTO) (entities.FeatureCollection, error)
	Capa(ctx context.Context, capa string, f dto.FiltroGeoDTO) (entities.FeatureCollection, error)
	Capas(ctx context.Context) (dto.CapasIndexDTO, error)
	Resumen(ctx context.Context) (dto.ResumenDTO, error)
	Edificios(ctx context.Context) ([]byte, error)
}
