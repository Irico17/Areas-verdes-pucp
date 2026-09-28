// Package services contains cross-cutting application services.
package services

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
)

// MatrizPermisos defines default role-action permissions.
// jefatura includes 'evidencias' per current apps/api specification.
var MatrizPermisos = map[string][]string{
	enums.RolCapataz.String():      {"consultar", "registrar"},
	enums.RolCoordinacion.String(): {"consultar", "registrar", "validar", "solicitudes", "reportes"},
	enums.RolJefatura.String():     {"consultar", "validar", "reportes", "solicitudes", "evidencias"},
	enums.RolAdmin.String():        {"consultar", "registrar", "validar", "reportes", "catalogos", "solicitudes"},
}

type permisosService struct{}

// NewPermisosService creates an in-memory permissions checking service.
func NewPermisosService() contracts.IPermisosService {
	return &permisosService{}
}

func (s *permisosService) Permite(rol, accion string) bool {
	acciones, ok := MatrizPermisos[rol]
	if !ok {
		return false
	}
	for _, a := range acciones {
		if a == accion {
			return true
		}
	}
	return false
}

func (s *permisosService) PermiteAlguno(rol string, acciones ...string) bool {
	for _, accion := range acciones {
		if s.Permite(rol, accion) {
			return true
		}
	}
	return false
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
