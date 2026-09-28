// Package usecases contains application use cases.
package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// CapasConocidas defines the auxiliary reference layers.
var CapasConocidas = []string{"jardines_reserva", "xerofitica"}

type geoUseCase struct {
	repo          contracts.IGeoRepository
	staticAdapter contracts.IArchivoEstaticoAdapter
}

// NewGeoUseCase creates a new GeoUseCase instance.
func NewGeoUseCase(repo contracts.IGeoRepository, staticAdapter contracts.IArchivoEstaticoAdapter) contracts.IGeoUseCase {
	return &geoUseCase{
		repo:          repo,
		staticAdapter: staticAdapter,
	}
}

func (u *geoUseCase) Areas(ctx context.Context, f dto.FiltroGeoDTO) (entities.FeatureCollection, error) {
	return u.repo.Areas(ctx, filtroGeoDTOToEntity(f))
}

func (u *geoUseCase) Zonas(ctx context.Context, f dto.FiltroGeoDTO) (entities.FeatureCollection, error) {
	return u.repo.Zonas(ctx, filtroGeoDTOToEntity(f))
}

func (u *geoUseCase) Capa(ctx context.Context, capa string, f dto.FiltroGeoDTO) (entities.FeatureCollection, error) {
	if !isKnownCapa(capa) {
		return entities.FeatureCollection{}, domainErrors.ErrCapaDesconocida
	}
	return u.repo.Capa(ctx, capa, filtroGeoDTOToEntity(f))
}

func (u *geoUseCase) Capas(ctx context.Context) (dto.CapasIndexDTO, error) {
	idx, err := u.repo.Capas(ctx)
	if err != nil {
		return dto.CapasIndexDTO{}, err
	}
	cargadas := make([]dto.CapaCountDTO, len(idx.Cargadas))
	for i, c := range idx.Cargadas {
		cargadas[i] = dto.CapaCountDTO{
			Capa:     c.Capa,
			Features: c.Features,
		}
	}
	return dto.CapasIndexDTO{
		CapasConocidas: CapasConocidas,
		Cargadas:       cargadas,
	}, nil
}

func (u *geoUseCase) Resumen(ctx context.Context) (dto.ResumenDTO, error) {
	res, err := u.repo.Resumen(ctx)
	if err != nil {
		return dto.ResumenDTO{}, err
	}
	capas := make([]dto.CapaCountDTO, len(res.Capas))
	for i, c := range res.Capas {
		capas[i] = dto.CapaCountDTO{
			Capa:     c.Capa,
			Features: c.Features,
		}
	}
	return dto.ResumenDTO{
		CRS:               res.CRS,
		Areas:             res.Areas,
		AreasConGeometria: res.AreasConGeometria,
		Zonas:             res.Zonas,
		ZonasConGeometria: res.ZonasConGeometria,
		Capas:             capas,
	}, nil
}

func filtroGeoDTOToEntity(f dto.FiltroGeoDTO) entities.FiltroGeo {
	var bbox *entities.BBox
	if f.BBox != nil {
		bbox = &entities.BBox{
			MinX: f.BBox.MinX,
			MinY: f.BBox.MinY,
			MaxX: f.BBox.MaxX,
			MaxY: f.BBox.MaxY,
		}
	}
	return entities.FiltroGeo{
		BBox:  bbox,
		Limit: f.Limit,
	}
}

func (u *geoUseCase) Edificios(ctx context.Context) ([]byte, error) {
	data, err := u.staticAdapter.LeerEdificios(ctx)
	if err != nil || len(data) == 0 {
		return []byte(`{"type":"FeatureCollection","name":"edificios","features":[]}`), nil
	}
	return data, nil
}

func isKnownCapa(name string) bool {
	for _, c := range CapasConocidas {
		if c == name {
			return true
		}
	}
	return false
}
