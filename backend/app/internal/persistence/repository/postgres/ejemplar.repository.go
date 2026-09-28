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

type ejemplarRepository struct {
	db *gorm.DB
}

// NewEjemplarRepository creates a new IEjemplarRepository instance.
func NewEjemplarRepository(db *gorm.DB) contracts.IEjemplarRepository {
	return &ejemplarRepository{db: db}
}

func (r *ejemplarRepository) Listar(ctx context.Context, limit, offset int) ([]entities.Ejemplar, int, error) {
	var total int
	if err := r.db.WithContext(ctx).Raw(`SELECT count(*) FROM ejemplares WHERE activo`).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	q := `
		SELECT id, numero_origen, codigo, especie_id, nombre_comun,
		       tipo_vegetacion, cantidad, ubicacion_lugar_id, referencia,
		       lat, lon, observacion_fen_2026, salud, activo, created_at, updated_at
		FROM ejemplares
		WHERE activo
		ORDER BY numero_origen NULLS LAST, id`
	args := []any{}
	if limit > 0 {
		if offset < 0 {
			offset = 0
		}
		q += ` LIMIT $1 OFFSET $2`
		args = []any{limit, offset}
	}

	var modelsList []models.EjemplarModel
	if err := r.db.WithContext(ctx).Raw(q, args...).Scan(&modelsList).Error; err != nil {
		return nil, 0, err
	}

	out := make([]entities.Ejemplar, len(modelsList))
	for i, m := range modelsList {
		out[i] = *mapper.EjemplarModelToEntity(&m)
	}
	return out, total, nil
}

func (r *ejemplarRepository) Crear(ctx context.Context, e entities.Ejemplar) (entities.Ejemplar, error) {
	var m models.EjemplarModel
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO ejemplares (
		  numero_origen, codigo, especie_id, nombre_comun, tipo_vegetacion, cantidad,
		  ubicacion_lugar_id, referencia, lat, lon, observacion_fen_2026, salud, geom
		) VALUES (
		  $1, NULLIF($2, ''), $3, NULLIF($4, ''), NULLIF($5, ''), $6,
		  $7, NULLIF($8, ''), $9, $10, NULLIF($11, ''), NULL,
		  CASE WHEN $9::float8 IS NULL THEN NULL
		       ELSE ST_SetSRID(ST_MakePoint($10, $9), 4326) END
		)
		RETURNING id, numero_origen, codigo, especie_id, nombre_comun, tipo_vegetacion, cantidad,
		          ubicacion_lugar_id, referencia, lat, lon, observacion_fen_2026, salud, activo, created_at, updated_at`,
		e.NumeroOrigen, strings.TrimSpace(e.Codigo), e.EspecieID, strings.TrimSpace(e.NombreComun),
		strings.TrimSpace(e.TipoVegetacion), e.Cantidad, e.UbicacionLugarID, strings.TrimSpace(e.Referencia),
		e.Lat, e.Lon, strings.TrimSpace(e.ObservacionFen2026),
	).Scan(&m).Error
	if err != nil {
		return entities.Ejemplar{}, domainErrors.ErrEntrada
	}
	return *mapper.EjemplarModelToEntity(&m), nil
}

func (r *ejemplarRepository) Recodificar(ctx context.Context, ejemplarID int64, codigoNuevo string) (entities.CodigoHistorico, error) {
	var out entities.CodigoHistorico
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var anterior sql.NullString
		row := tx.Raw(`SELECT codigo FROM ejemplares WHERE id = $1 AND activo`, ejemplarID).Row()
		if err := row.Scan(&anterior); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domainErrors.ErrNoEncontrado
			}
			return err
		}
		previo := ""
		if anterior.Valid {
			previo = anterior.String
		}
		if previo == codigoNuevo {
			out = entities.CodigoHistorico{EjemplarID: ejemplarID, CodigoAnterior: previo, CodigoNuevo: codigoNuevo}
			return nil
		}
		var m models.CodigoHistoricoModel
		if err := tx.Raw(`
			INSERT INTO codigos_historicos (ejemplar_id, codigo_anterior, codigo_nuevo)
			VALUES ($1, $2, $3)
			RETURNING id, ejemplar_id, codigo_anterior, codigo_nuevo, created_at`,
			ejemplarID, previo, codigoNuevo).Scan(&m).Error; err != nil {
			return err
		}
		res := tx.Exec(`UPDATE ejemplares SET codigo = $2, updated_at = now() WHERE id = $1 AND activo`, ejemplarID, codigoNuevo)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return domainErrors.ErrNoEncontrado
		}
		out = *mapper.CodigoHistoricoModelToEntity(&m)
		return nil
	})
	if err != nil {
		return entities.CodigoHistorico{}, err
	}
	return out, nil
}

func (r *ejemplarRepository) ListarCodigos(ctx context.Context, ejemplarID int64) ([]entities.CodigoHistorico, error) {
	var list []models.CodigoHistoricoModel
	err := r.db.WithContext(ctx).Where("ejemplar_id = ?", ejemplarID).Order("id").Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]entities.CodigoHistorico, len(list))
	for i, m := range list {
		out[i] = *mapper.CodigoHistoricoModelToEntity(&m)
	}
	return out, nil
}
