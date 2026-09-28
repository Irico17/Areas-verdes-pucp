package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ICambioRepository defines persistence operations on the cambios table.
type ICambioRepository interface {
	Crear(ctx context.Context, cambio *entities.Cambio) error
}

// IAuditoriaService defines high-level change recording services.
type IAuditoriaService interface {
	RegistrarCambio(ctx context.Context, req dto.RegistrarCambioDTO) error
}
