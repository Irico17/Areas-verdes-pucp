package usecases

import (
	"context"
	"errors"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type seedAccount struct {
	usuario string
	nombre  string
	rol     string
	capataz string
}

var semillas = []seedAccount{
	{"norte", "Elsa Quispe", enums.RolCapataz.String(), "cap-norte"},
	{"sur", "Iván Paredes", enums.RolCapataz.String(), "cap-sur"},
	{"riego", "Nora Beltrán", enums.RolCapataz.String(), "cap-riego"},
	{"coordinacion", "Coordinación", enums.RolCoordinacion.String(), ""},
	{"jefatura", "Jefatura", enums.RolJefatura.String(), ""},
	{"admin", "Administración", enums.RolAdmin.String(), ""},
}

type semillaAccesosUseCase struct {
	transaccion contracts.ITransaccion
	usuarioRepo contracts.IUsuarioRepository
	permisoRepo contracts.IPermisoRepository
	permisosSvc contracts.IPermisosService
	hasher      contracts.IHasher
}

// NewSemillaAccesosUseCase creates a use case to seed accounts and default permissions.
func NewSemillaAccesosUseCase(
	transaccion contracts.ITransaccion,
	usuarioRepo contracts.IUsuarioRepository,
	permisoRepo contracts.IPermisoRepository,
	permisosSvc contracts.IPermisosService,
	hasher contracts.IHasher,
) contracts.ISemillaAccesosUseCase {
	return &semillaAccesosUseCase{
		transaccion: transaccion,
		usuarioRepo: usuarioRepo,
		permisoRepo: permisoRepo,
		permisosSvc: permisosSvc,
		hasher:      hasher,
	}
}

func (uc *semillaAccesosUseCase) Ensure(ctx context.Context, password string) error {
	password = strings.TrimSpace(password)
	if password == "" {
		return errors.New("falta CAMPUS_DEV_PASSWORD")
	}

	hash, err := uc.hasher.Hash(password)
	if err != nil {
		return err
	}

	return uc.transaccion.Ejecutar(ctx, func(txCtx context.Context) error {
		for _, s := range semillas {
			existe, err := uc.usuarioRepo.ExistePorUsuario(txCtx, s.usuario)
			if err != nil {
				return err
			}
			if existe {
				continue
			}
			u := &entities.Usuario{
				Usuario:      s.usuario,
				Nombre:       s.nombre,
				Rol:          s.rol,
				CapatazID:    s.capataz,
				PasswordHash: hash,
				Activo:       true,
			}
			if err := uc.usuarioRepo.Crear(txCtx, u); err != nil {
				return err
			}
		}

		permisosList := make([]entities.Permiso, 0)
		for rol, acciones := range uc.permisosSvc.Matriz() {
			for _, accion := range acciones {
				permisosList = append(permisosList, entities.Permiso{
					Rol:    rol,
					Accion: accion,
				})
			}
		}

		return uc.permisoRepo.Sembrar(txCtx, permisosList)
	})
}
