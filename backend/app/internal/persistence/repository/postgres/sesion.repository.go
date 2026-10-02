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

type sesionRepository struct {
	db *gorm.DB
}

// NewSesionRepository creates a new postgres repository for sessions.
func NewSesionRepository(db *gorm.DB) contracts.ISesionRepository {
	return &sesionRepository{db: db}
}

func (r *sesionRepository) dbWithCtx(ctx context.Context) *gorm.DB {
	return database.DBFromContext(ctx, r.db).WithContext(ctx)
}

func (r *sesionRepository) Crear(ctx context.Context, sesion *entities.Sesion) error {
	m := models.SesionModel{
		TokenHash: sesion.TokenHash,
		UsuarioID: sesion.UsuarioID,
	}
	return r.dbWithCtx(ctx).Exec(`
		INSERT INTO sesiones (token_hash, usuario_id, expires_at)
		VALUES ($1, $2, now() + interval '12 hours')`,
		m.TokenHash, m.UsuarioID,
	).Error
}

func (r *sesionRepository) ObtenerPorTokenHash(ctx context.Context, tokenHash string) (*entities.Usuario, error) {
	var m models.UsuarioModel
	err := r.dbWithCtx(ctx).Raw(`
		SELECT u.id, u.usuario, u.nombre, u.rol, r.nombre AS rol_nombre, u.capataz_id, u.debe_cambiar_password
		FROM sesiones s
		JOIN usuarios u ON u.id = s.usuario_id
		LEFT JOIN roles r ON r.codigo = u.rol
		WHERE s.token_hash = $1 AND s.expires_at > now() AND u.activo AND COALESCE(r.activo, false)`,
		tokenHash,
	).Row().Scan(&m.ID, &m.Usuario, &m.Nombre, &m.Rol, &m.RolNombre, &m.CapatazID, &m.DebeCambiarPassword)
	if err != nil {
		return nil, err
	}
	m.Activo = true
	m.RolActivo = true
	return mapper.UsuarioToEntity(&m), nil
}

func (r *sesionRepository) EliminarPorTokenHash(ctx context.Context, tokenHash string) error {
	return r.dbWithCtx(ctx).Exec(`DELETE FROM sesiones WHERE token_hash = $1`, tokenHash).Error
}
