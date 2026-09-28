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
		SELECT id, numero_origen, COALESCE(codigo, ''), especie_id, COALESCE(nombre_comun, ''),
		       COALESCE(tipo_vegetacion, ''), cantidad, ubicacion_lugar_id, COALESCE(referencia, ''),
		       lat, lon, COALESCE(observacion_fen_2026, ''), salud, activo
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

	rows, err := r.db.WithContext(ctx).Raw(q, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []entities.Ejemplar{}
	for rows.Next() {
		var e entities.Ejemplar
		if err := rows.Scan(
			&e.ID, &e.NumeroOrigen, &e.Codigo, &e.EspecieID, &e.NombreComun,
			&e.TipoVegetacion, &e.Cantidad, &e.UbicacionLugarID, &e.Referencia,
			&e.Lat, &e.Lon, &e.ObservacionFen2026, &e.Salud, &e.Activo,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

func (r *ejemplarRepository) Crear(ctx context.Context, e entities.Ejemplar) (entities.Ejemplar, error) {
	var id int64
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
		RETURNING id`,
		e.NumeroOrigen, strings.TrimSpace(e.Codigo), e.EspecieID, strings.TrimSpace(e.NombreComun),
		strings.TrimSpace(e.TipoVegetacion), e.Cantidad, e.UbicacionLugarID, strings.TrimSpace(e.Referencia),
		e.Lat, e.Lon, strings.TrimSpace(e.ObservacionFen2026),
	).Scan(&id).Error
	if err != nil {
		return entities.Ejemplar{}, domainErrors.ErrEntrada
	}
	e.ID = id
	e.Salud = nil
	e.Activo = true
	return e, nil
}

func (r *ejemplarRepository) Recodificar(ctx context.Context, ejemplarID int64, codigoNuevo string) (entities.CodigoHistorico, error) {
	codigoNuevo = strings.TrimSpace(codigoNuevo)
	if ejemplarID < 1 || codigoNuevo == "" {
		return entities.CodigoHistorico{}, domainErrors.ErrEntrada
	}

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
		if err := tx.Raw(`
			INSERT INTO codigos_historicos (ejemplar_id, codigo_anterior, codigo_nuevo)
			VALUES ($1, $2, $3)
			RETURNING id, ejemplar_id, codigo_anterior, codigo_nuevo`,
			ejemplarID, previo, codigoNuevo).Scan(&out).Error; err != nil {
			return err
		}
		res := tx.Exec(`UPDATE ejemplares SET codigo = $2, updated_at = now() WHERE id = $1 AND activo`, ejemplarID, codigoNuevo)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return domainErrors.ErrNoEncontrado
		}
		return nil
	})
	if err != nil {
		return entities.CodigoHistorico{}, err
	}
	return out, nil
}

func (r *ejemplarRepository) ListarCodigos(ctx context.Context, ejemplarID int64) ([]entities.CodigoHistorico, error) {
	var out []entities.CodigoHistorico
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, ejemplar_id, codigo_anterior, COALESCE(codigo_nuevo, '')
		FROM codigos_historicos
		WHERE ejemplar_id = $1
		ORDER BY id`, ejemplarID).Scan(&out).Error
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []entities.CodigoHistorico{}
	}
	return out, nil
}
