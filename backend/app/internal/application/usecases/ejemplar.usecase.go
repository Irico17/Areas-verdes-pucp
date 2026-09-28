package usecases

import (
	"context"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
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
	ejemplares := make([]dto.EjemplarDTO, len(rows))
	for i, r := range rows {
		ejemplares[i] = ejemplarEntityToDTO(r)
	}
	return dto.EjemplaresPaginadosDTO{
		Ejemplares: ejemplares,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (uc *ejemplarUseCase) Crear(ctx context.Context, req dto.EjemplarDTO) (dto.EjemplarDTO, error) {
	e := ejemplarDTOToEntity(req)
	if err := e.Validar(); err != nil {
		return dto.EjemplarDTO{}, err
	}
	creado, err := uc.repo.Crear(ctx, e)
	if err != nil {
		return dto.EjemplarDTO{}, err
	}
	return ejemplarEntityToDTO(creado), nil
}

func (uc *ejemplarUseCase) Recodificar(ctx context.Context, ejemplarID int64, req dto.RecodificarDTO) (dto.CodigoHistoricoDTO, error) {
	codigo := strings.TrimSpace(req.Codigo)
	if ejemplarID < 1 || codigo == "" {
		return dto.CodigoHistoricoDTO{}, domainErrors.ErrEntrada
	}
	ch, err := uc.repo.Recodificar(ctx, ejemplarID, codigo)
	if err != nil {
		return dto.CodigoHistoricoDTO{}, err
	}
	return codigoHistoricoEntityToDTO(ch), nil
}

func (uc *ejemplarUseCase) ListarCodigos(ctx context.Context, ejemplarID int64) ([]dto.CodigoHistoricoDTO, error) {
	rows, err := uc.repo.ListarCodigos(ctx, ejemplarID)
	if err != nil {
		return nil, err
	}
	codigos := make([]dto.CodigoHistoricoDTO, len(rows))
	for i, r := range rows {
		codigos[i] = codigoHistoricoEntityToDTO(r)
	}
	return codigos, nil
}

func ejemplarEntityToDTO(e entities.Ejemplar) dto.EjemplarDTO {
	return dto.EjemplarDTO{
		ID:                 e.ID,
		NumeroOrigen:       e.NumeroOrigen,
		Codigo:             e.Codigo,
		EspecieID:          e.EspecieID,
		NombreComun:        e.NombreComun,
		TipoVegetacion:     e.TipoVegetacion,
		Cantidad:           e.Cantidad,
		UbicacionLugarID:   e.UbicacionLugarID,
		Referencia:         e.Referencia,
		Lat:                e.Lat,
		Lon:                e.Lon,
		ObservacionFen2026: e.ObservacionFen2026,
		Salud:              e.Salud,
		Activo:             e.Activo,
	}
}

func ejemplarDTOToEntity(d dto.EjemplarDTO) entities.Ejemplar {
	return entities.Ejemplar{
		ID:                 d.ID,
		NumeroOrigen:       d.NumeroOrigen,
		Codigo:             d.Codigo,
		EspecieID:          d.EspecieID,
		NombreComun:        d.NombreComun,
		TipoVegetacion:     d.TipoVegetacion,
		Cantidad:           d.Cantidad,
		UbicacionLugarID:   d.UbicacionLugarID,
		Referencia:         d.Referencia,
		Lat:                d.Lat,
		Lon:                d.Lon,
		ObservacionFen2026: d.ObservacionFen2026,
		Salud:              d.Salud,
		Activo:             d.Activo,
	}
}

func codigoHistoricoEntityToDTO(c entities.CodigoHistorico) dto.CodigoHistoricoDTO {
	return dto.CodigoHistoricoDTO{
		ID:             c.ID,
		EjemplarID:     c.EjemplarID,
		CodigoAnterior: c.CodigoAnterior,
		CodigoNuevo:    c.CodigoNuevo,
	}
}
