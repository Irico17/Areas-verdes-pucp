package postgres

import (
	"context"
	"database/sql"
	"errors"

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

func (r *permisoRepository) Concedido(ctx context.Context, rol, accion string) (bool, error) {
	var n int
	err := r.dbWithCtx(ctx).Raw(
		`SELECT count(*) FROM permisos WHERE rol = $1 AND accion = $2`, rol, accion,
	).Scan(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *permisoRepository) Establecer(ctx context.Context, rol, accion string, concedido bool) error {
	if concedido {
		return r.dbWithCtx(ctx).Exec(`
			INSERT INTO permisos (rol, accion)
			VALUES ($1, $2)
			ON CONFLICT (rol, accion) DO NOTHING`, rol, accion).Error
	}
	return r.dbWithCtx(ctx).Exec(
		`DELETE FROM permisos WHERE rol = $1 AND accion = $2`, rol, accion,
	).Error
}

func (r *permisoRepository) RolActivo(ctx context.Context, rol string) (bool, error) {
	var activo bool
	err := r.dbWithCtx(ctx).Raw(
		`SELECT activo FROM roles WHERE codigo = $1`, rol,
	).Row().Scan(&activo)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return activo, nil
}

func (r *permisoRepository) ListarRoles(ctx context.Context) ([]entities.RolCatalogo, error) {
	var rows []struct {
		Codigo string `gorm:"column:codigo"`
		Nombre string `gorm:"column:nombre"`
		Activo bool   `gorm:"column:activo"`
		Orden  int    `gorm:"column:orden"`
	}
	err := r.dbWithCtx(ctx).Raw(
		`SELECT codigo, nombre, activo, orden FROM roles ORDER BY orden, codigo`,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]entities.RolCatalogo, 0, len(rows))
	for _, row := range rows {
		out = append(out, entities.RolCatalogo{
			Codigo: row.Codigo,
			Nombre: row.Nombre,
			Activo: row.Activo,
			Orden:  row.Orden,
		})
	}
	return out, nil
}

func (r *permisoRepository) CrearRol(ctx context.Context, rol entities.RolCatalogo) error {
	return r.dbWithCtx(ctx).Exec(`
		INSERT INTO roles (codigo, nombre, orden, activo)
		VALUES ($1, $2, (SELECT COALESCE(MAX(orden), 0) + 1 FROM roles), $3)`,
		rol.Codigo, rol.Nombre, rol.Activo,
	).Error
}

func (r *permisoRepository) ActualizarRol(ctx context.Context, codigo string, activo bool) error {
	return r.dbWithCtx(ctx).Exec(
		`UPDATE roles SET activo = $2 WHERE codigo = $1`, codigo, activo,
	).Error
}
