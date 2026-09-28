package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

// IAreaVerdeRepository defines persistence operations on the areas_verdes table.
type IAreaVerdeRepository interface {
	Fichas(ctx context.Context, q string) ([]dto.FichaDTO, error)
	ObtenerFichaPorFeatureID(ctx context.Context, featureID string) (*dto.FichaDTO, error)
	ActualizarFicha(ctx context.Context, featureID, nombre, uso, riego, referencia string) (*dto.FichaDTO, error)
	CrearSinGeom(ctx context.Context, featureID, nombre, uso string) (*dto.FichaDTO, error)
}

// IAreaVerdeUseCase defines use case operations for managing area fichas.
type IAreaVerdeUseCase interface {
	Fichas(ctx context.Context, q string) ([]dto.FichaDTO, error)
	ActualizarFicha(ctx context.Context, featureID string, req dto.ActualizarFichaDTO, usuarioID *int64) (*dto.FichaDTO, error)
	CrearSinGeom(ctx context.Context, req dto.CrearAreaSinGeomDTO, usuarioID *int64) (*dto.FichaDTO, error)
}
