package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

type permisoRepository struct {
	db *gorm.DB
}

// NewPermisoRepository creates a new postgres repository for permissions.
func NewPermisoRepository(db *gorm.DB) contracts.IPermisoRepository {
	return &permisoRepository{db: db}
}

func (r *permisoRepository) dbWithCtx(ctx context.Context) *gorm.DB {
	return database.DBFromContext(ctx, r.db).WithContext(ctx)
}

func (r *permisoRepository) Listar(ctx context.Context) ([]entities.Permiso, error) {
	var rows []models.PermisoModel
	err := r.dbWithCtx(ctx).Raw(`SELECT rol, accion FROM permisos ORDER BY rol, accion`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]entities.Permiso, 0, len(rows))
	for _, m := range rows {
		out = append(out, mapper.PermisoToEntity(&m))
	}
	return out, nil
}

func (r *permisoRepository) Sembrar(ctx context.Context, permisos []entities.Permiso) error {
	for _, p := range permisos {
		m := mapper.PermisoToModel(p)
		err := r.dbWithCtx(ctx).Exec(`
			INSERT INTO permisos (rol, accion)
			VALUES ($1, $2)
			ON CONFLICT (rol, accion) DO NOTHING`,
			m.Rol, m.Accion,
		).Error
		if err != nil {
			return err
		}
	}
	return nil
}
