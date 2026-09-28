package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

type cuadrillaRepository struct {
	db *gorm.DB
}

// NewCuadrillaRepository creates a new ICuadrillaRepository instance.
func NewCuadrillaRepository(db *gorm.DB) contracts.ICuadrillaRepository {
	return &cuadrillaRepository{db: db}
}

func (r *cuadrillaRepository) Listar(ctx context.Context) ([]entities.Cuadrilla, error) {
	var list []models.CuadrillaModel
	err := r.db.WithContext(ctx).Where("activo = ?", true).Order("id").Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]entities.Cuadrilla, len(list))
	for i, m := range list {
		out[i] = *mapper.CuadrillaModelToEntity(&m)
	}
	return out, nil
}

func (r *cuadrillaRepository) Crear(ctx context.Context, id, nombre, turno string) (entities.Cuadrilla, error) {
	m := models.CuadrillaModel{
		ID:             id,
		NombreFicticio: nombre,
		Turno:          turno,
		Activo:         true,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return entities.Cuadrilla{}, domainErrors.ErrEntrada
	}
	return *mapper.CuadrillaModelToEntity(&m), nil
}
