package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ISesionRepository defines persistence operations for session tokens.
type ISesionRepository interface {
	Crear(ctx context.Context, sesion *entities.Sesion) error
	ObtenerPorTokenHash(ctx context.Context, tokenHash string) (*entities.Usuario, error)
	EliminarPorTokenHash(ctx context.Context, tokenHash string) error
}

// ISesionUseCase defines application operations for authentication and session management.
type ISesionUseCase interface {
	Login(ctx context.Context, usuario, clave string) (string, *dto.UsuarioSesionDTO, error)
	Actual(ctx context.Context, token string) (*dto.UsuarioSesionDTO, error)
	Logout(ctx context.Context, token string)
	Resolver(ctx context.Context, token string) (*dto.UsuarioSesionDTO, error)
}
