// Package contracts defines interfaces for application use cases and persistence repositories.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IIntervencionRepository defines persistence operations for interventions and work logs.
type IIntervencionRepository interface {
	Capataces(ctx context.Context) ([]entities.Capataz, error)
	List(ctx context.Context, f dto.FiltroIntervencionesDTO) (entities.FeatureCollection, error)
	Create(ctx context.Context, in dto.CrearIntervencionDTO) (entities.Feature, bool, error)
	Assign(ctx context.Context, in dto.AsignarIntervencionDTO) (entities.Feature, error)
	SetEstado(ctx context.Context, in dto.CambiarEstadoDTO) (entities.Feature, error)
	Archive(ctx context.Context, in dto.ArchivarIntervencionDTO) error
	Timeline(ctx context.Context, id string) (dto.TimelineResponseDTO, error)
	GuardarFicha(ctx context.Context, in dto.FichaIntervencionDTO) error
	CrearAvance(ctx context.Context, in dto.CrearAvanceDTO) error
	One(ctx context.Context, id string) (entities.Feature, error)
}

// IIntervencionUseCase defines business operations for interventions, lifecycle, and progress.
type IIntervencionUseCase interface {
	ListarCapataces(ctx context.Context) ([]dto.CapatazDTO, error)
	ListarActividades(ctx context.Context, f dto.FiltroIntervencionesDTO) (entities.FeatureCollection, error)
	CrearActividad(ctx context.Context, in dto.CrearIntervencionDTO) (dto.CrearIntervencionResponseDTO, error)
	AsignarActividad(ctx context.Context, in dto.AsignarIntervencionDTO) (entities.Feature, error)
	CambiarEstado(ctx context.Context, in dto.CambiarEstadoDTO) (entities.Feature, error)
	ArchivarActividad(ctx context.Context, in dto.ArchivarIntervencionDTO) (dto.ArchivarIntervencionResponseDTO, error)
	Timeline(ctx context.Context, id string) (dto.TimelineResponseDTO, error)
	GuardarFicha(ctx context.Context, in dto.FichaIntervencionDTO) error
	CrearAvance(ctx context.Context, in dto.CrearAvanceDTO) error
}
