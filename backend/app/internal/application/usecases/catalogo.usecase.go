package usecases

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
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
		dtoItems = append(dtoItems, catalogoItemDTO(it))
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

	out := catalogoItemDTO(*item)
	return &out, nil
}

// Desactivar performs a logical deactivation of a catalog item.
func (uc *catalogoUseCase) Desactivar(ctx context.Context, id int64, usuarioID int64) (*dto.DesactivarCatalogoResponseDTO, error) {
	if err := uc.repo.Deactivate(ctx, id, usuarioID); err != nil {
		return nil, err
	}
	return &dto.DesactivarCatalogoResponseDTO{
		Activo: false,
		ID:     id,
	}, nil
}

// Renombrar corrects the visible name and keeps the previous row.
func (uc *catalogoUseCase) Renombrar(ctx context.Context, in dto.RenombrarCatalogoDTO) (*dto.CatalogoItemDTO, error) {
	nombre := strings.TrimSpace(in.Nombre)
	if nombre == "" || utf8.RuneCountInString(nombre) > 80 {
		return nil, apperrors.ErrNombreObligatorio
	}
	if in.ID <= 0 {
		return nil, apperrors.ErrItemNoExiste
	}
	item, err := uc.repo.Renombrar(ctx, in.ID, nombre, in.UsuarioID)
	if err != nil {
		return nil, err
	}
	out := catalogoItemDTO(*item)
	return &out, nil
}

func catalogoItemDTO(it entities.CatalogoItem) dto.CatalogoItemDTO {
	return dto.CatalogoItemDTO{
		ID:          it.ID,
		Clase:       it.Clase,
		Codigo:      it.Codigo,
		Nombre:      it.Nombre,
		Activo:      it.Activo,
		Orden:       it.Orden,
		Provisional: it.Provisional,
	}
}
