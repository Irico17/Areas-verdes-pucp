package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type lugarRepository struct {
	db *gorm.DB
}

// NewLugarRepository creates a new ILugarRepository instance.
func NewLugarRepository(db *gorm.DB) contracts.ILugarRepository {
	return &lugarRepository{db: db}
}

func (r *lugarRepository) Listar(ctx context.Context) ([]entities.Lugar, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id, nombre, nombre_norm, lat, lon, zona_supervision_id, activo
		FROM lugares
		WHERE activo
		ORDER BY nombre_norm`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.Lugar{}
	for rows.Next() {
		var l entities.Lugar
		var zona *int64
		if err := rows.Scan(&l.ID, &l.Nombre, &l.NombreNorm, &l.Lat, &l.Lon, &zona, &l.Activo); err != nil {
			return nil, err
		}
		l.ZonaSupervisionID = zona
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *lugarRepository) Crear(ctx context.Context, nombre string, lat, lon float64, zonaID *int64) (entities.Lugar, error) {
	norm := entities.NormalizarNombre(nombre)
	var id int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO lugares (nombre, nombre_norm, lat, lon, zona_supervision_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`, nombre, norm, lat, lon, zonaID).Scan(&id).Error
	if err != nil {
		return entities.Lugar{}, domainErrors.ErrEntrada
	}
	return entities.Lugar{
		ID:                id,
		Nombre:            nombre,
		NombreNorm:        norm,
		Lat:               lat,
		Lon:               lon,
		ZonaSupervisionID: zonaID,
		Activo:            true,
	}, nil
}
