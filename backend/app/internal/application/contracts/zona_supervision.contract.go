package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IZonaSupervisionRepository defines data access for supervision zones.
type IZonaSupervisionRepository interface {
	Listar(ctx context.Context) ([]entities.ZonaSupervision, error)
	Crear(ctx context.Context, codigo, nombre, geojson string, area *float64) (entities.ZonaSupervision, error)
	Actualizar(ctx context.Context, codigo, nombre, geojson string, area *float64, usuarioID *int64) (entities.ZonaSupervision, error)
	Baja(ctx context.Context, codigo string, usuarioID *int64) (entities.ZonaSupervision, error)
}

// IZonaSupervisionUseCase defines business logic for supervision zones.
type IZonaSupervisionUseCase interface {
	Listar(ctx context.Context) ([]dto.ZonaSupervisionDTO, error)
	Crear(ctx context.Context, req dto.CrearZonaSupervisionDTO) (dto.ZonaSupervisionDTO, error)
	Actualizar(ctx context.Context, codigo string, req dto.ActualizarZonaSupervisionDTO, usuarioID *int64) (dto.ZonaSupervisionDTO, error)
	Baja(ctx context.Context, codigo string, usuarioID *int64) (dto.ZonaSupervisionDTO, error)
}
