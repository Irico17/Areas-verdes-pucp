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

type lugarRepository struct {
	db *gorm.DB
}

// NewLugarRepository creates a new ILugarRepository instance.
func NewLugarRepository(db *gorm.DB) contracts.ILugarRepository {
	return &lugarRepository{db: db}
}

func (r *lugarRepository) Listar(ctx context.Context) ([]entities.Lugar, error) {
	var list []models.LugarModel
	err := r.db.WithContext(ctx).Where("activo = ?", true).Order("nombre_norm").Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]entities.Lugar, len(list))
	for i, m := range list {
		out[i] = *mapper.LugarModelToEntity(&m)
	}
	return out, nil
}

func (r *lugarRepository) Crear(ctx context.Context, nombre string, lat, lon float64, zonaID *int64) (entities.Lugar, error) {
	norm := entities.NormalizarNombre(nombre)
	m := models.LugarModel{
		Nombre:            nombre,
		NombreNorm:        norm,
		Lat:               lat,
		Lon:               lon,
		ZonaSupervisionID: zonaID,
		Activo:            true,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return entities.Lugar{}, domainErrors.ErrEntrada
	}
	return *mapper.LugarModelToEntity(&m), nil
}
