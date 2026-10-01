// Package contracts defines interfaces for application services and repositories.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IServicioTercerizadoRepository defines persistence operations for work orders (ordenes_servicio).
type IServicioTercerizadoRepository interface {
	Listar(ctx context.Context) ([]*entities.ServicioTercerizado, error)
	ObtenerPorID(ctx context.Context, id string) (*entities.ServicioTercerizado, error)
	Crear(ctx context.Context, in dto.CrearOrdenDTO, actorRol, capatazID string) (*entities.ServicioTercerizado, error)
	Editar(ctx context.Context, in dto.EditarOrdenDTO, actorRol, capatazID string) (*entities.ServicioTercerizado, error)
}

// IServicioTercerizadoUseCase defines application operations for work orders.
type IServicioTercerizadoUseCase interface {
	Listar(ctx context.Context) (*dto.OrdenesResponseDTO, error)
	Crear(ctx context.Context, in dto.CrearOrdenDTO, actorRol, capatazID string) (*dto.OrdenDTO, error)
	Editar(ctx context.Context, in dto.EditarOrdenDTO, actorRol, capatazID string) (*dto.EditarOrdenResponseDTO, error)
}
