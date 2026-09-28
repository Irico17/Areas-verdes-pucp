// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

// IReservasMockAdapter defines access to the mock reservations data source.
type IReservasMockAdapter interface {
	LeerReservas(ctx context.Context) ([]byte, error)
}

// IReservasMockUseCase defines business operations for the mock reservations agenda.
type IReservasMockUseCase interface {
	ObtenerAgenda(ctx context.Context) (dto.ReservasMockResponseDTO, []byte, error)
}
