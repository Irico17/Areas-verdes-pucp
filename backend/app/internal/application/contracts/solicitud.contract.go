// Package contracts defines interfaces for application services and repositories.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ISolicitudRepository defines persistence operations for solicitudes.
type ISolicitudRepository interface {
	Listar(ctx context.Context) ([]*entities.Solicitud, error)
	ObtenerPorID(ctx context.Context, id string) (*entities.Solicitud, error)
	Crear(ctx context.Context, in dto.CrearSolicitudDTO) (*entities.Solicitud, error)
	Editar(ctx context.Context, in dto.EditarSolicitudDTO) (*entities.Solicitud, error)
}

// ISolicitudUseCase defines application operations for solicitudes.
type ISolicitudUseCase interface {
	Listar(ctx context.Context) (*dto.SolicitudesResponseDTO, error)
	Crear(ctx context.Context, in dto.CrearSolicitudDTO) (*dto.SolicitudDTO, error)
	Editar(ctx context.Context, in dto.EditarSolicitudDTO) (*dto.SolicitudDTO, error)
}
