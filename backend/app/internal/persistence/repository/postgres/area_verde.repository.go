package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

type areaVerdeRepository struct {
	db *gorm.DB
}

// NewAreaVerdeRepository creates a new AreaVerdeRepository instance.
func NewAreaVerdeRepository(db *gorm.DB) contracts.IAreaVerdeRepository {
	return &areaVerdeRepository{db: db}
}

func (r *areaVerdeRepository) Fichas(ctx context.Context, q string) ([]entities.AreaVerdeFicha, error) {
	q = strings.TrimSpace(q)
	// Solo áreas activas en listados (como zonas_supervision).
	// Las áreas dadas de baja no salen en GET /catastro/areas.
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT feature_id, COALESCE(nombre, ''), COALESCE(uso, ''), COALESCE(riego_act, ''),
		       COALESCE(referencia, ''), area_m2, geom IS NOT NULL, activo
		FROM areas_verdes
		WHERE activo
		  AND ($1 = '' OR feature_id ILIKE '%' || $1 || '%'
		   OR COALESCE(nombre, '') ILIKE '%' || $1 || '%'
		   OR COALESCE(codigo, '') ILIKE '%' || $1 || '%'
		   OR COALESCE(uso, '') ILIKE '%' || $1 || '%')
		ORDER BY (NULLIF(btrim(COALESCE(nombre, '')), '') IS NULL), lower(COALESCE(nombre, '')), feature_id
		LIMIT 40`, q).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.AreaVerdeFicha{}
	for rows.Next() {
		var (
			fid, nombre, uso, riego, ref string
			area                         sql.NullFloat64
			conGeom                      bool
			activo                       bool
		)
		if err := rows.Scan(&fid, &nombre, &uso, &riego, &ref, &area, &conGeom, &activo); err != nil {
			return nil, err
		}
		m := models.AreaVerdeModel{
			FeatureID:  fid,
			Nombre:     &nombre,
			Uso:        &uso,
			RiegoAct:   &riego,
			Referencia: &ref,
			AreaM2:     mapper.NullFloat(area),
			Activo:     activo,
		}
		out = append(out, mapper.AreaVerdeModelToFicha(&m, conGeom))
	}
	return out, rows.Err()
}

func (r *areaVerdeRepository) ObtenerFichaPorFeatureID(ctx context.Context, featureID string) (*entities.AreaVerdeFicha, error) {
	featureID = strings.TrimSpace(featureID)
	var (
		fid, nombre, uso, riego, ref string
		area                         sql.NullFloat64
		conGeom                      bool
		activo                       bool
	)
	row := r.db.WithContext(ctx).Raw(`
		SELECT feature_id, COALESCE(nombre, ''), COALESCE(uso, ''), COALESCE(riego_act, ''),
		       COALESCE(referencia, ''), area_m2, geom IS NOT NULL, activo
		FROM areas_verdes
		WHERE feature_id = $1`, featureID).Row()

	err := row.Scan(&fid, &nombre, &uso, &riego, &ref, &area, &conGeom, &activo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domainErrors.ErrFichaNoEncontrada
	}
	if err != nil {
		return nil, err
	}
	m := models.AreaVerdeModel{
		FeatureID:  fid,
		Nombre:     &nombre,
		Uso:        &uso,
		RiegoAct:   &riego,
		Referencia: &ref,
		AreaM2:     mapper.NullFloat(area),
		Activo:     activo,
	}
	ficha := mapper.AreaVerdeModelToFicha(&m, conGeom)
	return &ficha, nil
}

func (r *areaVerdeRepository) ActualizarFicha(ctx context.Context, featureID, nombre, uso, riego, referencia string) (*entities.AreaVerdeFicha, error) {
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

func (r *areaVerdeRepository) CrearSinGeom(ctx context.Context, featureID, nombre, uso string) (*entities.AreaVerdeFicha, error) {
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

// Baja marca el área como inactiva. La fila permanece. Si ya estaba inactiva, no duplica el historial.
func (r *areaVerdeRepository) Baja(ctx context.Context, featureID string, usuarioID *int64) error {
	featureID = strings.TrimSpace(featureID)
	if featureID == "" {
		return domainErrors.ErrEntrada
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var id int64
		var activo bool
		err := tx.Raw(`SELECT id, activo FROM areas_verdes WHERE feature_id = $1 FOR UPDATE`, featureID).Row().Scan(&id, &activo)
		if errors.Is(err, sql.ErrNoRows) {
			return domainErrors.ErrFichaNoEncontrada
		}
		if err != nil {
			return err
		}
		if !activo {
			return nil
		}
		res := tx.Exec(`UPDATE areas_verdes SET activo = FALSE, updated_at = now() WHERE id = $1 AND activo`, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return domainErrors.ErrFichaNoEncontrada
		}
		return registrarCambioTx(tx, "areas_verdes", featureID, "baja",
			map[string]any{"feature_id": featureID, "activo": true},
			map[string]any{"feature_id": featureID, "activo": false},
			usuarioID,
		)
	})
}
