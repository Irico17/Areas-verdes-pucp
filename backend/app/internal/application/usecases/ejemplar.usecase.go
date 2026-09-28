package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type ejemplarUseCase struct {
	repo contracts.IEjemplarRepository
}

// NewEjemplarUseCase creates a new IEjemplarUseCase instance.
func NewEjemplarUseCase(repo contracts.IEjemplarRepository) contracts.IEjemplarUseCase {
	return &ejemplarUseCase{repo: repo}
}

func (uc *ejemplarUseCase) Listar(ctx context.Context, limit, offset int) (dto.EjemplaresPaginadosDTO, error) {
	rows, total, err := uc.repo.Listar(ctx, limit, offset)
	if err != nil {
		return dto.EjemplaresPaginadosDTO{}, err
	}
	if rows == nil {
		rows = []entities.Ejemplar{}
	}
	return dto.EjemplaresPaginadosDTO{
		Ejemplares: rows,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (uc *ejemplarUseCase) Crear(ctx context.Context, req dto.EjemplarDTO) (dto.EjemplarDTO, error) {
	e := entities.Ejemplar(req)
	if err := e.Validar(); err != nil {
		return dto.EjemplarDTO{}, err
	}
	return uc.repo.Crear(ctx, e)
}

func (uc *ejemplarUseCase) Recodificar(ctx context.Context, ejemplarID int64, req dto.RecodificarDTO) (dto.CodigoHistoricoDTO, error) {
	return uc.repo.Recodificar(ctx, ejemplarID, req.Codigo)
}

func (uc *ejemplarUseCase) ListarCodigos(ctx context.Context, ejemplarID int64) ([]dto.CodigoHistoricoDTO, error) {
	rows, err := uc.repo.ListarCodigos(ctx, ejemplarID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []dto.CodigoHistoricoDTO{}
	}
	return rows, nil
}
