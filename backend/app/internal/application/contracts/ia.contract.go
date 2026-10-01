// Package contracts defines abstract interfaces between layers.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ISugeridorTipo defines the contract for heuristic type suggestions based on title keywords.
type ISugeridorTipo interface {
	SugerirTipo(titulo string) entities.SugerenciaIA
}

// IIAUseCase defines the application use case for IA-related endpoints.
type IIAUseCase interface {
	Sugerir(ctx context.Context, titulo string) dto.SugerenciaIADTO
}
