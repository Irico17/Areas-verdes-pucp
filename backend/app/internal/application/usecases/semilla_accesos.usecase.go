package usecases

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/services"
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
	{"norte", "Equipo Norte", enums.RolCapataz.String(), "cap-norte"},
	{"sur", "Equipo Sur", enums.RolCapataz.String(), "cap-sur"},
	{"riego", "Equipo Riego", enums.RolCapataz.String(), "cap-riego"},
	{"coordinacion", "Coordinación", enums.RolCoordinacion.String(), ""},
	{"jefatura", "Jefatura", enums.RolJefatura.String(), ""},
	{"admin", "Administración", enums.RolAdmin.String(), ""},
}

type semillaAccesosUseCase struct {
	db          *gorm.DB
	usuarioRepo contracts.IUsuarioRepository
	permisoRepo contracts.IPermisoRepository
	hasher      contracts.IHasher
}

// NewSemillaAccesosUseCase creates a use case to seed accounts and default permissions.
func NewSemillaAccesosUseCase(
	db *gorm.DB,
	usuarioRepo contracts.IUsuarioRepository,
	permisoRepo contracts.IPermisoRepository,
	hasher contracts.IHasher,
) contracts.ISemillaAccesosUseCase {
	return &semillaAccesosUseCase{
		db:          db,
		usuarioRepo: usuarioRepo,
		permisoRepo: permisoRepo,
		hasher:      hasher,
	}
}

func (uc *semillaAccesosUseCase) Ensure(ctx context.Context, password string) error {
	if uc.db == nil {
		return errors.New("sin base")
	}
	password = strings.TrimSpace(password)
	if password == "" {
		return errors.New("falta CAMPUS_DEV_PASSWORD")
	}

	hash, err := uc.hasher.Hash(password)
	if err != nil {
		return err
	}

	return uc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, s := range semillas {
			var n int64
			if err := tx.Raw(`SELECT count(*) FROM usuarios WHERE usuario = $1`, s.usuario).Scan(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				continue
			}
			if err := tx.Exec(`
				INSERT INTO usuarios (usuario, nombre, rol, capataz_id, password_hash)
				VALUES ($1, $2, $3, NULLIF($4, ''), $5)`,
				s.usuario, s.nombre, s.rol, s.capataz, hash,
			).Error; err != nil {
				return err
			}
		}

		permisosList := make([]entities.Permiso, 0)
		for rol, acciones := range services.MatrizPermisos {
			for _, accion := range acciones {
				permisosList = append(permisosList, entities.Permiso{
					Rol:    rol,
					Accion: accion,
				})
			}
		}

		for _, p := range permisosList {
			if err := tx.Exec(`
				INSERT INTO permisos (rol, accion)
				VALUES ($1, $2)
				ON CONFLICT (rol, accion) DO NOTHING`,
				p.Rol, p.Accion,
			).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
