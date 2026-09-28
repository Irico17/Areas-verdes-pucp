package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ICatalogoRepository defines persistence operations for catalog items.
type ICatalogoRepository interface {
	List(ctx context.Context, clase string, soloActivos bool) ([]entities.CatalogoItem, error)
	Activo(ctx context.Context, clase, codigo string) (bool, error)
	Create(ctx context.Context, clase, codigo, nombre string) (*entities.CatalogoItem, error)
	Deactivate(ctx context.Context, id int64) error
}

// ICatalogoUseCase defines business use cases for catalog management.
type ICatalogoUseCase interface {
	Listar(ctx context.Context, filtro dto.FiltroCatalogoDTO) (*dto.CatalogoListResponseDTO, error)
	Crear(ctx context.Context, in dto.CrearCatalogoDTO) (*dto.CatalogoItemDTO, error)
	Desactivar(ctx context.Context, id int64) (*dto.DesactivarCatalogoResponseDTO, error)
}
