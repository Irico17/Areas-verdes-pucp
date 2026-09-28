package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type cuadrillaRepository struct {
	db *gorm.DB
}

// NewCuadrillaRepository creates a new ICuadrillaRepository instance.
func NewCuadrillaRepository(db *gorm.DB) contracts.ICuadrillaRepository {
	return &cuadrillaRepository{db: db}
}

func (r *cuadrillaRepository) Listar(ctx context.Context) ([]entities.Cuadrilla, error) {
	var out []entities.Cuadrilla
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, nombre_ficticio, turno, activo
		FROM cuadrillas
		WHERE activo
		ORDER BY id`).Scan(&out).Error
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []entities.Cuadrilla{}
	}
	return out, nil
}

func (r *cuadrillaRepository) Crear(ctx context.Context, id, nombre, turno string) (entities.Cuadrilla, error) {
	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO cuadrillas (id, nombre_ficticio, turno) VALUES ($1, $2, $3)`,
		id, nombre, turno).Error
	if err != nil {
		return entities.Cuadrilla{}, domainErrors.ErrEntrada
	}
	return entities.Cuadrilla{ID: id, NombreFicticio: nombre, Turno: turno, Activo: true}, nil
}
