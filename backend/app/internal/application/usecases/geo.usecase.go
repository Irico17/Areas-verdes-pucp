// Package usecases contains application use cases.
package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

var capasConocidas = []string{"jardines_reserva", "xerofitica"}

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
	return u.repo.Areas(ctx, f)
}

func (u *geoUseCase) Zonas(ctx context.Context, f dto.FiltroGeoDTO) (entities.FeatureCollection, error) {
	return u.repo.Zonas(ctx, f)
}

func (u *geoUseCase) Capa(ctx context.Context, capa string, f dto.FiltroGeoDTO) (entities.FeatureCollection, error) {
	if !isKnownCapa(capa) {
		return entities.FeatureCollection{}, domainErrors.ErrCapaDesconocida
	}
	return u.repo.Capa(ctx, capa, f)
}

func (u *geoUseCase) Capas(ctx context.Context) (dto.CapasIndexDTO, error) {
	idx, err := u.repo.Capas(ctx)
	if err != nil {
		return dto.CapasIndexDTO{}, err
	}
	idx.CapasConocidas = capasConocidas
	if idx.Cargadas == nil {
		idx.Cargadas = []dto.CapaCountDTO{}
	}
	return idx, nil
}

func (u *geoUseCase) Resumen(ctx context.Context) (dto.ResumenDTO, error) {
	return u.repo.Resumen(ctx)
}

func (u *geoUseCase) Edificios(ctx context.Context) ([]byte, error) {
	data, err := u.staticAdapter.LeerEdificios(ctx)
	if err != nil || len(data) == 0 {
		return []byte(`{"type":"FeatureCollection","name":"edificios","features":[]}`), nil
	}
	return data, nil
}

func isKnownCapa(name string) bool {
	for _, c := range capasConocidas {
		if c == name {
			return true
		}
	}
	return false
}
