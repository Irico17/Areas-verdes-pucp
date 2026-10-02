package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IEjemplarRepository defines data access for individual flora specimens.
type IEjemplarRepository interface {
	Listar(ctx context.Context, limit, offset int, q string) ([]entities.Ejemplar, int, error)
	Crear(ctx context.Context, e entities.Ejemplar) (entities.Ejemplar, error)
	Actualizar(ctx context.Context, id int64, p entities.ParcheEjemplar, usuarioID *int64) (entities.Ejemplar, error)
	Recodificar(ctx context.Context, ejemplarID int64, codigoNuevo string, usuarioID *int64) (entities.CodigoHistorico, error)
	ListarCodigos(ctx context.Context, ejemplarID int64) ([]entities.CodigoHistorico, error)
}

// IEjemplarUseCase defines business logic for individual flora specimens.
type IEjemplarUseCase interface {
	Listar(ctx context.Context, limit, offset int, q string) (dto.EjemplaresPaginadosDTO, error)
	Crear(ctx context.Context, e dto.EjemplarDTO) (dto.EjemplarDTO, error)
	Actualizar(ctx context.Context, id int64, req dto.ActualizarEjemplarDTO, usuarioID *int64) (dto.EjemplarDTO, error)
	Recodificar(ctx context.Context, ejemplarID int64, req dto.RecodificarDTO, usuarioID *int64) (dto.CodigoHistoricoDTO, error)
	ListarCodigos(ctx context.Context, ejemplarID int64) ([]dto.CodigoHistoricoDTO, error)
}
