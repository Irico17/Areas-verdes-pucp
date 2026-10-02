// Package services contains cross-cutting application services.
package services

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
)

// MatrizPermisos is the seed written by Ensure with ON CONFLICT DO NOTHING.
// Permite reads the permisos table and roles.activo on every call, so an edit
// applies without restarting the process. Startup never deletes rows.
var MatrizPermisos = map[string][]string{
	enums.RolCapataz.String():      {"consultar", "registrar"},
	enums.RolCoordinacion.String(): {"consultar", "registrar", "validar", "solicitudes", "reportes"},
	enums.RolJefatura.String():     {"consultar", "validar", "reportes", "solicitudes", "evidencias", "usuarios"},
	enums.RolAdmin.String():        {"consultar", "registrar", "validar", "reportes", "catalogos", "solicitudes", "usuarios"},
}

type permisosService struct {
	repo contracts.IPermisoRepository
}

// NewPermisosService checks permissions against the repository on each call.
func NewPermisosService(repo contracts.IPermisoRepository) contracts.IPermisosService {
	return &permisosService{repo: repo}
}

// NewPermisosMemoria serves the seed matrix in memory. Route tests use it.
func NewPermisosMemoria() contracts.IPermisosService {
	return NewPermisosService(NuevaMemoriaPermisos(MatrizPermisos))
}

func (s *permisosService) Permite(rol, accion string) bool {
	if s == nil || s.repo == nil {
		return false
	}
	ctx := context.Background()
	activo, err := s.repo.RolActivo(ctx, rol)
	if err != nil || !activo {
		return false
	}
	ok, err := s.repo.Concedido(ctx, rol, accion)
	return err == nil && ok
}

func (s *permisosService) Matriz() map[string][]string {
	out := make(map[string][]string, len(MatrizPermisos))
	for k, v := range MatrizPermisos {
		dup := make([]string, len(v))
		copy(dup, v)
		out[k] = dup
	}
	return out
}
