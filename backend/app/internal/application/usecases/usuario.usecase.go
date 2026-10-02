package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

const avisoCuentas = "Cuentas de VerdePUCP. Las administra la jefatura de sección."

type usuarioUseCase struct {
	usuarioRepo contracts.IUsuarioRepository
	permisoRepo contracts.IPermisoRepository
}

// NewUsuarioUseCase creates a new user usecase implementation.
func NewUsuarioUseCase(
	usuarioRepo contracts.IUsuarioRepository,
	permisoRepo contracts.IPermisoRepository,
) contracts.IUsuarioUseCase {
	return &usuarioUseCase{
		usuarioRepo: usuarioRepo,
		permisoRepo: permisoRepo,
	}
}

func (uc *usuarioUseCase) ListarUsuarios(ctx context.Context) (*dto.UsuariosResponseDTO, error) {
	usuarios, err := uc.usuarioRepo.Listar(ctx)
	if err != nil {
		return nil, err
	}

	permisos, err := uc.permisoRepo.Listar(ctx)
	if err != nil {
		return nil, err
	}

	usuariosDTO := make([]dto.UsuarioSesionDTO, 0, len(usuarios))
	for _, u := range usuarios {
		usuariosDTO = append(usuariosDTO, dto.UsuarioSesionDTO{
			ID:        u.ID,
			Usuario:   u.Usuario,
			Nombre:    u.Nombre,
			Rol:       u.Rol,
			RolNombre: u.RolNombre,
			CapatazID: u.CapatazID,
		})
	}

	permisosDTO := make([]dto.PermisoDTO, 0, len(permisos))
	for _, p := range permisos {
		permisosDTO = append(permisosDTO, dto.PermisoDTO{
			Rol:    p.Rol,
			Accion: p.Accion,
		})
	}

	return &dto.UsuariosResponseDTO{
		Usuarios: usuariosDTO,
		Permisos: permisosDTO,
		Aviso:    avisoCuentas,
	}, nil
}
