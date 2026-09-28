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

type especieRepository struct {
	db *gorm.DB
}

// NewEspecieRepository creates a new IEspecieRepository instance.
func NewEspecieRepository(db *gorm.DB) contracts.IEspecieRepository {
	return &especieRepository{db: db}
}

func (r *especieRepository) Listar(ctx context.Context) ([]entities.Especie, error) {
	var list []models.EspecieModel
	err := r.db.WithContext(ctx).Where("activo = ?", true).Order("nombre_cientifico").Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]entities.Especie, len(list))
	for i, m := range list {
		out[i] = *mapper.EspecieModelToEntity(&m)
	}
	return out, nil
}

func (r *especieRepository) Crear(ctx context.Context, cientifico, comun string) (entities.Especie, error) {
	var comunPtr *string
	if comun != "" {
		comunPtr = &comun
	}
	m := models.EspecieModel{
		NombreCientifico: cientifico,
		NombreComun:      comunPtr,
		Activo:           true,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return entities.Especie{}, domainErrors.ErrEntrada
	}
	return *mapper.EspecieModelToEntity(&m), nil
}
