package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type especieRepository struct {
	db *gorm.DB
}

// NewEspecieRepository creates a new IEspecieRepository instance.
func NewEspecieRepository(db *gorm.DB) contracts.IEspecieRepository {
	return &especieRepository{db: db}
}

func (r *especieRepository) Listar(ctx context.Context) ([]entities.Especie, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id, nombre_cientifico, COALESCE(nombre_comun, ''), activo
		FROM especies WHERE activo ORDER BY nombre_cientifico`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.Especie{}
	for rows.Next() {
		var e entities.Especie
		if err := rows.Scan(&e.ID, &e.NombreCientifico, &e.NombreComun, &e.Activo); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *especieRepository) Crear(ctx context.Context, cientifico, comun string) (entities.Especie, error) {
	var id int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO especies (nombre_cientifico, nombre_comun)
		VALUES ($1, NULLIF($2, ''))
		RETURNING id`, cientifico, comun).Scan(&id).Error
	if err != nil {
		return entities.Especie{}, domainErrors.ErrEntrada
	}
	return entities.Especie{
		ID:               id,
		NombreCientifico: cientifico,
		NombreComun:      comun,
		Activo:           true,
	}, nil
}
