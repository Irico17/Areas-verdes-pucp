package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IPermisoRepository defines persistence operations for permissions.
type IPermisoRepository interface {
	Listar(ctx context.Context) ([]entities.Permiso, error)
	Sembrar(ctx context.Context, permisos []entities.Permiso) error
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
