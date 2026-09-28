package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type permisoRepository struct {
	db *gorm.DB
}

// NewPermisoRepository creates a new postgres repository for permissions.
func NewPermisoRepository(db *gorm.DB) contracts.IPermisoRepository {
	return &permisoRepository{db: db}
}

func (r *permisoRepository) Listar(ctx context.Context) ([]entities.Permiso, error) {
	var out []entities.Permiso
	err := r.db.WithContext(ctx).Raw(`SELECT rol, accion FROM permisos ORDER BY rol, accion`).Scan(&out).Error
	if out == nil {
		out = []entities.Permiso{}
	}
	return out, err
}

func (r *permisoRepository) Sembrar(ctx context.Context, permisos []entities.Permiso) error {
	for _, p := range permisos {
		err := r.db.WithContext(ctx).Exec(`
			INSERT INTO permisos (rol, accion)
			VALUES ($1, $2)
			ON CONFLICT (rol, accion) DO NOTHING`,
			p.Rol, p.Accion,
		).Error
		if err != nil {
			return err
		}
	}
	return nil
}
