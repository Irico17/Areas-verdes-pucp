package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IEjemplarRepository defines data access for individual flora specimens.
type IEjemplarRepository interface {
	Listar(ctx context.Context, limit, offset int) ([]entities.Ejemplar, int, error)
	Crear(ctx context.Context, e entities.Ejemplar) (entities.Ejemplar, error)
	Recodificar(ctx context.Context, ejemplarID int64, codigoNuevo string) (entities.CodigoHistorico, error)
	ListarCodigos(ctx context.Context, ejemplarID int64) ([]entities.CodigoHistorico, error)
}

// IEjemplarUseCase defines business logic for individual flora specimens.
type IEjemplarUseCase interface {
	Listar(ctx context.Context, limit, offset int) (dto.EjemplaresPaginadosDTO, error)
	Crear(ctx context.Context, e dto.EjemplarDTO) (dto.EjemplarDTO, error)
	Recodificar(ctx context.Context, ejemplarID int64, req dto.RecodificarDTO) (dto.CodigoHistoricoDTO, error)
	ListarCodigos(ctx context.Context, ejemplarID int64) ([]dto.CodigoHistoricoDTO, error)
}
