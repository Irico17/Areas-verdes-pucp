// Package usecases contains application use case implementations.
package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type inventarioUseCase struct {
	repo        contracts.IInventarioRepository
	fotoAdapter contracts.IFotoDiscoAdapter
}

// NewInventarioUseCase creates a new instance of IInventarioUseCase.
func NewInventarioUseCase(repo contracts.IInventarioRepository, fotoAdapter contracts.IFotoDiscoAdapter) contracts.IInventarioUseCase {
	return &inventarioUseCase{
		repo:        repo,
		fotoAdapter: fotoAdapter,
	}
}

// Index returns the catalogue of inventory overlays and currently loaded feature counts.
func (u *inventarioUseCase) Index(ctx context.Context) (dto.IndiceInventarioDTO, error) {
	idx, err := u.repo.Index(ctx)
	if err != nil {
		return dto.IndiceInventarioDTO{}, err
	}
	cargadas := make([]dto.CapaCountDTO, len(idx.Cargadas))
	for i, c := range idx.Cargadas {
		cargadas[i] = dto.CapaCountDTO{
			Capa:     c.Capa,
			Features: c.Features,
		}
	}
	return dto.IndiceInventarioDTO{
		Capas:    idx.Capas,
		Cargadas: cargadas,
	}, nil
}

// Capa returns the GeoJSON FeatureCollection for a given inventory layer.
func (u *inventarioUseCase) Capa(ctx context.Context, capa string) (entities.FeatureCollection, error) {
	return u.repo.Capa(ctx, capa)
}

// Foto resolves the local file path for an inventory photograph.
func (u *inventarioUseCase) Foto(_ context.Context, name string) (string, error) {
	return u.fotoAdapter.RutaFoto(name)
}
