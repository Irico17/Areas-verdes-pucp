package usecases

import (
	"context"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type zonaSupervisionUseCase struct {
	repo contracts.IZonaSupervisionRepository
}

// NewZonaSupervisionUseCase creates a new IZonaSupervisionUseCase instance.
func NewZonaSupervisionUseCase(repo contracts.IZonaSupervisionRepository) contracts.IZonaSupervisionUseCase {
	return &zonaSupervisionUseCase{repo: repo}
}

func (uc *zonaSupervisionUseCase) Listar(ctx context.Context) ([]dto.ZonaSupervisionDTO, error) {
	rows, err := uc.repo.Listar(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ZonaSupervisionDTO, len(rows))
	for i, z := range rows {
		out[i] = zonaSupervisionEntityToDTO(z)
	}
	return out, nil
}

func (uc *zonaSupervisionUseCase) Crear(ctx context.Context, req dto.CrearZonaSupervisionDTO) (dto.ZonaSupervisionDTO, error) {
	z := entities.ZonaSupervision{
		Codigo: req.Codigo,
		Nombre: req.Nombre,
		AreaM2: req.AreaM2,
	}
	if err := z.Validar(req.GeoJSON); err != nil {
		return dto.ZonaSupervisionDTO{}, err
	}
	res, err := uc.repo.Crear(ctx, z.Codigo, z.Nombre, req.GeoJSON, z.AreaM2)
	if err != nil {
		return dto.ZonaSupervisionDTO{}, err
	}
	return zonaSupervisionEntityToDTO(res), nil
}

func (uc *zonaSupervisionUseCase) Actualizar(ctx context.Context, codigo string, req dto.ActualizarZonaSupervisionDTO, usuarioID *int64) (dto.ZonaSupervisionDTO, error) {
	codigo = strings.TrimSpace(codigo)
	nombre := strings.TrimSpace(req.Nombre)
	if err := entities.ValidarCodigoZona(codigo); err != nil || nombre == "" {
		return dto.ZonaSupervisionDTO{}, domainErrors.ErrEntrada
	}
	if req.AreaM2 != nil && *req.AreaM2 < 0 {
		return dto.ZonaSupervisionDTO{}, domainErrors.ErrEntrada
	}
	res, err := uc.repo.Actualizar(ctx, codigo, nombre, strings.TrimSpace(req.GeoJSON), req.AreaM2, usuarioID)
	if err != nil {
		return dto.ZonaSupervisionDTO{}, err
	}
	return zonaSupervisionEntityToDTO(res), nil
}

func (uc *zonaSupervisionUseCase) Baja(ctx context.Context, codigo string, usuarioID *int64) (dto.ZonaSupervisionDTO, error) {
	codigo = strings.TrimSpace(codigo)
	if err := entities.ValidarCodigoZona(codigo); err != nil {
		return dto.ZonaSupervisionDTO{}, err
	}
	res, err := uc.repo.Baja(ctx, codigo, usuarioID)
	if err != nil {
		return dto.ZonaSupervisionDTO{}, err
	}
	return zonaSupervisionEntityToDTO(res), nil
}

func zonaSupervisionEntityToDTO(z entities.ZonaSupervision) dto.ZonaSupervisionDTO {
	return dto.ZonaSupervisionDTO{
		ID:      z.ID,
		Codigo:  z.Codigo,
		Nombre:  z.Nombre,
		AreaM2:  z.AreaM2,
		ConGeom: z.ConGeom,
		Activo:  z.Activo,
	}
}
