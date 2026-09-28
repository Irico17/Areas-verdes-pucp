package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
)

type areaVerdeRepository struct {
	db *gorm.DB
}

// NewAreaVerdeRepository creates a new AreaVerdeRepository instance.
func NewAreaVerdeRepository(db *gorm.DB) contracts.IAreaVerdeRepository {
	return &areaVerdeRepository{db: db}
}

func (r *areaVerdeRepository) Fichas(ctx context.Context, q string) ([]dto.FichaDTO, error) {
	q = strings.TrimSpace(q)
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT feature_id, COALESCE(nombre, ''), COALESCE(uso, ''), COALESCE(riego_act, ''),
		       COALESCE(referencia, ''), area_m2, geom IS NOT NULL
		FROM areas_verdes
		WHERE ($1 = '' OR feature_id ILIKE '%' || $1 || '%'
		   OR COALESCE(nombre, '') ILIKE '%' || $1 || '%'
		   OR COALESCE(codigo, '') ILIKE '%' || $1 || '%'
		   OR COALESCE(uso, '') ILIKE '%' || $1 || '%')
		ORDER BY (NULLIF(btrim(COALESCE(nombre, '')), '') IS NULL), lower(COALESCE(nombre, '')), feature_id
		LIMIT 40`, q).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []dto.FichaDTO{}
	for rows.Next() {
		var f dto.FichaDTO
		var area sql.NullFloat64
		if err := rows.Scan(&f.FeatureID, &f.Nombre, &f.Uso, &f.RiegoAct, &f.Referencia, &area, &f.ConGeom); err != nil {
			return nil, err
		}
		f.AreaM2 = mapper.NullFloat(area)
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *areaVerdeRepository) ObtenerFichaPorFeatureID(ctx context.Context, featureID string) (*dto.FichaDTO, error) {
	featureID = strings.TrimSpace(featureID)
	var f dto.FichaDTO
	var area sql.NullFloat64
	row := r.db.WithContext(ctx).Raw(`
		SELECT feature_id, COALESCE(nombre, ''), COALESCE(uso, ''), COALESCE(riego_act, ''),
		       COALESCE(referencia, ''), area_m2, geom IS NOT NULL
		FROM areas_verdes
		WHERE feature_id = $1`, featureID).Row()

	err := row.Scan(&f.FeatureID, &f.Nombre, &f.Uso, &f.RiegoAct, &f.Referencia, &area, &f.ConGeom)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domainErrors.ErrFichaNoEncontrada
	}
	if err != nil {
		return nil, err
	}
	f.AreaM2 = mapper.NullFloat(area)
	return &f, nil
}

func (r *areaVerdeRepository) ActualizarFicha(ctx context.Context, featureID, nombre, uso, riego, referencia string) (*dto.FichaDTO, error) {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE areas_verdes
		SET nombre = NULLIF($2, ''), uso = NULLIF($3, ''), riego_act = NULLIF($4, ''),
		    referencia = NULLIF($5, ''), updated_at = now()
		WHERE feature_id = $1`,
		featureID, nombre, uso, riego, referencia,
	)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, domainErrors.ErrFichaNoEncontrada
	}
	return r.ObtenerFichaPorFeatureID(ctx, featureID)
}

func (r *areaVerdeRepository) CrearSinGeom(ctx context.Context, featureID, nombre, uso string) (*dto.FichaDTO, error) {
	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO areas_verdes (feature_id, source_index, codigo, nombre, uso, geom)
		VALUES (
		  $1,
		  (SELECT COALESCE(MAX(source_index), 0) + 1 FROM areas_verdes),
		  $1, $2, NULLIF($3, ''), NULL
		)`, featureID, nombre, uso).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, domainErrors.ErrEntrada
		}
		return nil, err
	}
	return r.ObtenerFichaPorFeatureID(ctx, featureID)
}
