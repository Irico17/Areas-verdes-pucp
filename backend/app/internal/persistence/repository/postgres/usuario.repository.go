package postgres

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type usuarioRepository struct {
	db *gorm.DB
}

// NewUsuarioRepository creates a new postgres repository for users.
func NewUsuarioRepository(db *gorm.DB) contracts.IUsuarioRepository {
	return &usuarioRepository{db: db}
}

func (r *usuarioRepository) ObtenerPorUsuario(ctx context.Context, usuario string) (*entities.Usuario, error) {
	var row struct {
		ID        int64
		Usuario   string
		Nombre    string
		Rol       string
		RolNombre *string
		CapatazID *string
		Hash      string
		Activo    bool
	}

	err := r.db.WithContext(ctx).Raw(`
		SELECT u.id, u.usuario, u.nombre, u.rol, r.nombre, u.capataz_id, u.password_hash, u.activo
		FROM usuarios u
		LEFT JOIN roles r ON r.codigo = u.rol
		WHERE u.usuario = $1`, strings.TrimSpace(usuario)).Row().Scan(
		&row.ID, &row.Usuario, &row.Nombre, &row.Rol, &row.RolNombre, &row.CapatazID, &row.Hash, &row.Activo,
	)
	if err != nil {
		return nil, err
	}

	u := &entities.Usuario{
		ID:           row.ID,
		Usuario:      row.Usuario,
		Nombre:       row.Nombre,
		Rol:          row.Rol,
		PasswordHash: row.Hash,
		Activo:       row.Activo,
	}
	if row.RolNombre != nil {
		u.RolNombre = *row.RolNombre
	}
	if row.CapatazID != nil {
		u.CapatazID = *row.CapatazID
	}
	return u, nil
}

func (r *usuarioRepository) Listar(ctx context.Context) ([]entities.Usuario, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT u.id, u.usuario, u.nombre, u.rol, COALESCE(r.nombre, u.rol), COALESCE(u.capataz_id, '')
		FROM usuarios u
		LEFT JOIN roles r ON r.codigo = u.rol
		ORDER BY u.rol, u.usuario`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.Usuario{}
	for rows.Next() {
		var u entities.Usuario
		if err := rows.Scan(&u.ID, &u.Usuario, &u.Nombre, &u.Rol, &u.RolNombre, &u.CapatazID); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *usuarioRepository) ExistePorUsuario(ctx context.Context, usuario string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Raw(`SELECT count(*) FROM usuarios WHERE usuario = $1`, strings.TrimSpace(usuario)).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *usuarioRepository) Crear(ctx context.Context, usuario *entities.Usuario) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO usuarios (usuario, nombre, rol, capataz_id, password_hash)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5)`,
		usuario.Usuario, usuario.Nombre, usuario.Rol, usuario.CapatazID, usuario.PasswordHash,
	).Error
}
