package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
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
	if rows == nil {
		rows = []entities.PoligonoCuadrilla{}
	}
	return rows, nil
}

func (uc *catastroReferenciaUseCase) ListarCapa(ctx context.Context, tabla string) ([]dto.CapaFichaDTO, error) {
	rows, err := uc.repo.ListarCapa(ctx, tabla)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []entities.CapaFicha{}
	}
	return rows, nil
}
