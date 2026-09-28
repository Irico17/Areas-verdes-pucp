package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ICatastroReferenciaRepository defines data access for polygons and auxiliary reference layers.
type ICatastroReferenciaRepository interface {
	ListarPoligonos(ctx context.Context) ([]entities.PoligonoCuadrilla, error)
	ListarCapa(ctx context.Context, tabla string) ([]entities.CapaFicha, error)
}

// ICatastroReferenciaUseCase defines business logic for polygons and auxiliary reference layers.
type ICatastroReferenciaUseCase interface {
	ListarPoligonos(ctx context.Context) ([]dto.PoligonoCuadrillaDTO, error)
	ListarCapa(ctx context.Context, tabla string) ([]dto.CapaFichaDTO, error)
}
