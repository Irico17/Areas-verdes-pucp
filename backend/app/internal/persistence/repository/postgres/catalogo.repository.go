package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

type catalogoRepository struct {
	db *gorm.DB
}

// NewCatalogoRepository creates a new postgres repository for catalog items.
func NewCatalogoRepository(db *gorm.DB) contracts.ICatalogoRepository {
	return &catalogoRepository{db: db}
}

func (r *catalogoRepository) List(ctx context.Context, clase string, soloActivos bool) ([]entities.CatalogoItem, error) {
	q := `SELECT id, clase, codigo, nombre, activo, orden FROM catalogos WHERE 1=1`
	args := []any{}
	if clase != "" {
		q += ` AND clase = $1`
		args = append(args, clase)
	}
	if soloActivos {
		q += ` AND activo`
	}
	q += ` ORDER BY clase, orden, codigo`

	var rows []models.CatalogoModel
	err := r.db.WithContext(ctx).Raw(q, args...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]entities.CatalogoItem, 0, len(rows))
	for _, m := range rows {
		out = append(out, entities.CatalogoItem{
			ID:     m.ID,
			Clase:  m.Clase,
			Codigo: m.Codigo,
			Nombre: m.Nombre,
			Activo: m.Activo,
			Orden:  m.Orden,
		})
	}
	return out, nil
}

func (r *catalogoRepository) Activo(ctx context.Context, clase, codigo string) (bool, error) {
	var n int
	err := r.db.WithContext(ctx).Raw(`
		SELECT count(*) FROM catalogos WHERE clase = $1 AND codigo = $2 AND activo`,
		clase, codigo).Scan(&n).Error
	return n == 1, err
}

func (r *catalogoRepository) Create(ctx context.Context, clase, codigo, nombre string) (*entities.CatalogoItem, error) {
	var m models.CatalogoModel
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO catalogos (clase, codigo, nombre)
		VALUES ($1, $2, $3)
		ON CONFLICT (clase, codigo) DO UPDATE SET nombre = EXCLUDED.nombre, activo = TRUE
		RETURNING id, clase, codigo, nombre, activo, orden`,
		clase, codigo, nombre).Row().Scan(&m.ID, &m.Clase, &m.Codigo, &m.Nombre, &m.Activo, &m.Orden)
	if err != nil {
		return nil, err
	}
	return &entities.CatalogoItem{
		ID:     m.ID,
		Clase:  m.Clase,
		Codigo: m.Codigo,
		Nombre: m.Nombre,
		Activo: m.Activo,
		Orden:  m.Orden,
	}, nil
}

func (r *catalogoRepository) Deactivate(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Exec(`UPDATE catalogos SET activo = FALSE WHERE id = $1`, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrItemNoExiste
	}
	return nil
}
