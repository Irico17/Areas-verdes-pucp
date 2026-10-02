package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IPermisoRepository defines persistence operations for permissions and roles.
type IPermisoRepository interface {
	Listar(ctx context.Context) ([]entities.Permiso, error)
	Sembrar(ctx context.Context, permisos []entities.Permiso) error
	Concedido(ctx context.Context, rol, accion string) (bool, error)
	Establecer(ctx context.Context, rol, accion string, concedido bool) error
	RolActivo(ctx context.Context, rol string) (bool, error)
	ListarRoles(ctx context.Context) ([]entities.RolCatalogo, error)
	CrearRol(ctx context.Context, rol entities.RolCatalogo) error
	ActualizarRol(ctx context.Context, codigo string, activo bool) error
}

// IPermisosService checks role-based permissions in memory.
type IPermisosService interface {
	Permite(rol, accion string) bool
	Matriz() map[string][]string
}

// ISemillaAccesosUseCase seeds initial users and default permissions.
type ISemillaAccesosUseCase interface {
	Ensure(ctx context.Context, password string) error
}
