package usecases

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

var codigoRegex = regexp.MustCompile(`^[a-z0-9_]{2,32}$`)

type catalogoUseCase struct {
	repo contracts.ICatalogoRepository
}

// NewCatalogoUseCase creates a new catalog use case.
func NewCatalogoUseCase(repo contracts.ICatalogoRepository) contracts.ICatalogoUseCase {
	return &catalogoUseCase{repo: repo}
}

// Listar retrieves catalog items according to the filter.
func (uc *catalogoUseCase) Listar(ctx context.Context, filtro dto.FiltroCatalogoDTO) (*dto.CatalogoListResponseDTO, error) {
	items, err := uc.repo.List(ctx, filtro.Clase, filtro.SoloActivos)
	if err != nil {
		return nil, err
	}

	dtoItems := make([]dto.CatalogoItemDTO, 0, len(items))
	for _, it := range items {
		dtoItems = append(dtoItems, dto.CatalogoItemDTO{
			ID:     it.ID,
			Clase:  it.Clase,
			Codigo: it.Codigo,
			Nombre: it.Nombre,
			Activo: it.Activo,
			Orden:  it.Orden,
		})
	}

	return &dto.CatalogoListResponseDTO{
		Items:  dtoItems,
		Clases: enums.ClasesCatalogoValidas(),
	}, nil
}

// Crear validates and creates or reactivates a catalog item.
func (uc *catalogoUseCase) Crear(ctx context.Context, in dto.CrearCatalogoDTO) (*dto.CatalogoItemDTO, error) {
	clase := strings.TrimSpace(in.Clase)
	codigo := strings.TrimSpace(in.Codigo)
	nombre := strings.TrimSpace(in.Nombre)

	if !enums.ClaseCatalogo(clase).EsValida() {
		return nil, apperrors.ErrClaseNoReconocida
	}
	if !codigoRegex.MatchString(codigo) {
		return nil, apperrors.ErrCodigoInvalido
	}
	if nombre == "" || utf8.RuneCountInString(nombre) > 80 {
		return nil, apperrors.ErrNombreObligatorio
	}

	item, err := uc.repo.Create(ctx, clase, codigo, nombre)
	if err != nil {
		return nil, err
	}

	return &dto.CatalogoItemDTO{
		ID:     item.ID,
		Clase:  item.Clase,
		Codigo: item.Codigo,
		Nombre: item.Nombre,
		Activo: item.Activo,
		Orden:  item.Orden,
	}, nil
}

// Desactivar performs a logical deactivation of a catalog item.
func (uc *catalogoUseCase) Desactivar(ctx context.Context, id int64) (*dto.DesactivarCatalogoResponseDTO, error) {
	if err := uc.repo.Deactivate(ctx, id); err != nil {
		return nil, err
	}
	return &dto.DesactivarCatalogoResponseDTO{
		Activo: false,
		ID:     id,
	}, nil
}
