package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

type catastroReferenciaUseCase struct {
	repo contracts.ICatastroReferenciaRepository
}

// NewCatastroReferenciaUseCase creates a new ICatastroReferenciaUseCase instance.
func NewCatastroReferenciaUseCase(repo contracts.ICatastroReferenciaRepository) contracts.ICatastroReferenciaUseCase {
	return &catastroReferenciaUseCase{repo: repo}
}

func (uc *catastroReferenciaUseCase) ListarPoligonos(ctx context.Context) ([]dto.PoligonoCuadrillaDTO, error) {
	rows, err := uc.repo.ListarPoligonos(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.PoligonoCuadrillaDTO, len(rows))
	for i, r := range rows {
		out[i] = dto.PoligonoCuadrillaDTO{
			ID:                r.ID,
			FeatureID:         r.FeatureID,
			Codigo:            r.Codigo,
			Nombre:            r.Nombre,
			CuadrillaID:       r.CuadrillaID,
			ZonaSupervisionID: r.ZonaSupervisionID,
			ConGeom:           r.ConGeom,
			Activo:            r.Activo,
		}
	}
	return out, nil
}

func (uc *catastroReferenciaUseCase) ListarCapa(ctx context.Context, tabla string) ([]dto.CapaFichaDTO, error) {
	rows, err := uc.repo.ListarCapa(ctx, tabla)
	if err != nil {
		return nil, err
	}
	out := make([]dto.CapaFichaDTO, len(rows))
	for i, r := range rows {
		out[i] = dto.CapaFichaDTO{
			ID:         r.ID,
			FeatureID:  r.FeatureID,
			Nombre:     r.Nombre,
			Codigo:     r.Codigo,
			Nota:       r.Nota,
			Clase:      r.Clase,
			Riego:      r.Riego,
			Pertenecen: r.Pertenecen,
			Activo:     r.Activo,
		}
	}
	return out, nil
}
