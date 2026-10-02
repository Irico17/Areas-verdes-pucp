package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IUsuarioRepository defines persistence operations for users.
type IUsuarioRepository interface {
	ObtenerPorUsuario(ctx context.Context, usuario string) (*entities.Usuario, error)
	Listar(ctx context.Context) ([]entities.Usuario, error)
	ExistePorUsuario(ctx context.Context, usuario string) (bool, error)
	Crear(ctx context.Context, usuario *entities.Usuario) error
	Actualizar(ctx context.Context, usuario *entities.Usuario) error
}

// IUsuarioUseCase defines application operations for user management.
type IUsuarioUseCase interface {
	ListarUsuarios(ctx context.Context) (*dto.UsuariosResponseDTO, error)
	Crear(ctx context.Context, actorID int64, in dto.CrearCuentaDTO) (*dto.CuentaDTO, error)
	Actualizar(ctx context.Context, actorID int64, actorUsuario, usuario string, in dto.ActualizarCuentaDTO) (*dto.CuentaDTO, error)
	CambiarClavePropia(ctx context.Context, usuario, actual, nueva string) (*dto.UsuarioSesionDTO, error)
	ActualizarPermiso(ctx context.Context, actorID int64, rol, accion string, concedido bool) error
	CrearRol(ctx context.Context, actorID int64, codigo, nombre string) (*dto.RolDTO, error)
	ActualizarRol(ctx context.Context, actorID int64, codigo string, activo bool) error
}
