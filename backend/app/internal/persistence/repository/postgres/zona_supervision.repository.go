package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type zonaSupervisionRepository struct {
	db *gorm.DB
}

// NewZonaSupervisionRepository creates a new IZonaSupervisionRepository instance.
func NewZonaSupervisionRepository(db *gorm.DB) contracts.IZonaSupervisionRepository {
	return &zonaSupervisionRepository{db: db}
}

func (r *zonaSupervisionRepository) Listar(ctx context.Context) ([]entities.ZonaSupervision, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id, codigo, nombre, area_m2, geom IS NOT NULL, activo
		FROM zonas_supervision
		WHERE activo
		ORDER BY codigo`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.ZonaSupervision{}
	for rows.Next() {
		var z entities.ZonaSupervision
		var area *float64
		if err := rows.Scan(&z.ID, &z.Codigo, &z.Nombre, &area, &z.ConGeom, &z.Activo); err != nil {
			return nil, err
		}
		z.AreaM2 = area
		out = append(out, z)
	}
	return out, rows.Err()
}

func (r *zonaSupervisionRepository) Crear(ctx context.Context, codigo, nombre, geojson string, area *float64) (entities.ZonaSupervision, error) {
	var id int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO zonas_supervision (codigo, nombre, area_m2, geom)
		VALUES ($1, $2, $3, catastro_geom_4326($4))
		RETURNING id`, codigo, nombre, area, geojson).Scan(&id).Error
	if err != nil {
		return entities.ZonaSupervision{}, domainErrors.ErrEntrada
	}
	list, err := r.Listar(ctx)
	if err != nil {
		return entities.ZonaSupervision{}, err
	}
	for _, z := range list {
		if z.ID == id {
			return z, nil
		}
	}
	return entities.ZonaSupervision{}, domainErrors.ErrNoEncontrado
}
