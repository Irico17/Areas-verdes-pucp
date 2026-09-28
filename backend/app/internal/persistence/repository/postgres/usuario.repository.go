package postgres

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

type usuarioRepository struct {
	db *gorm.DB
}

// NewUsuarioRepository creates a new postgres repository for users.
func NewUsuarioRepository(db *gorm.DB) contracts.IUsuarioRepository {
	return &usuarioRepository{db: db}
}

func (r *usuarioRepository) dbWithCtx(ctx context.Context) *gorm.DB {
	return database.DBFromContext(ctx, r.db).WithContext(ctx)
}

func (r *usuarioRepository) ObtenerPorUsuario(ctx context.Context, usuario string) (*entities.Usuario, error) {
	var m models.UsuarioModel
	err := r.dbWithCtx(ctx).Raw(`
		SELECT u.id, u.usuario, u.nombre, u.rol, r.nombre AS rol_nombre, u.capataz_id, u.password_hash, u.activo
		FROM usuarios u
		LEFT JOIN roles r ON r.codigo = u.rol
		WHERE u.usuario = $1`, strings.TrimSpace(usuario)).Row().Scan(
		&m.ID, &m.Usuario, &m.Nombre, &m.Rol, &m.RolNombre, &m.CapatazID, &m.PasswordHash, &m.Activo,
	)
	if err != nil {
		return nil, err
	}
	return mapper.UsuarioToEntity(&m), nil
}

func (r *usuarioRepository) Listar(ctx context.Context) ([]entities.Usuario, error) {
	var rows []models.UsuarioModel
	err := r.dbWithCtx(ctx).Raw(`
		SELECT u.id, u.usuario, u.nombre, u.rol, COALESCE(r.nombre, u.rol) AS rol_nombre, COALESCE(u.capataz_id, '') AS capataz_id
		FROM usuarios u
		LEFT JOIN roles r ON r.codigo = u.rol
		ORDER BY u.rol, u.usuario`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]entities.Usuario, 0, len(rows))
	for i := range rows {
		out = append(out, *mapper.UsuarioToEntity(&rows[i]))
	}
	return out, nil
}

func (r *usuarioRepository) ExistePorUsuario(ctx context.Context, usuario string) (bool, error) {
	var count int64
	err := r.dbWithCtx(ctx).Raw(`SELECT count(*) FROM usuarios WHERE usuario = $1`, strings.TrimSpace(usuario)).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *usuarioRepository) Crear(ctx context.Context, usuario *entities.Usuario) error {
	m := mapper.UsuarioToModel(usuario)
	return r.dbWithCtx(ctx).Exec(`
		INSERT INTO usuarios (usuario, nombre, rol, capataz_id, password_hash)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5)`,
		m.Usuario, m.Nombre, m.Rol, m.CapatazID, m.PasswordHash,
	).Error
}
