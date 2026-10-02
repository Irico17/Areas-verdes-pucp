// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

// ISectoresUseCase defines use case operations for sector groupings and commands.
type ISectoresUseCase interface {
	// Ejecutar executes sector generation from jefe_de_grupo.json.
	Ejecutar(ctx context.Context, opt dto.OpcionesSectoresDTO) (*dto.ResultadoSectoresDTO, error)
}
