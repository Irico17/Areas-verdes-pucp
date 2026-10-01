// Package postgres implements repository interfaces using GORM and PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type viveroRepository struct {
	db *gorm.DB
}

// NewViveroRepository creates a new IViveroRepository instance.
func NewViveroRepository(db *gorm.DB) contracts.IViveroRepository {
	return &viveroRepository{db: db}
}

func (r *viveroRepository) Listar(ctx context.Context, mes string) ([]*entities.Vivero, error) {
	q := `
		SELECT id::text, COALESCE(to_char(fecha, 'YYYY-MM-DD'), ''), area, subproceso, etapa,
		       descripcion, observaciones, responsables, COALESCE(lugar_id, ''), lugar_libre
		FROM vivero_registros
		WHERE archivada_en IS NULL`
	args := []any{}
	if mes != "" {
		q += ` AND to_char(fecha, 'YYYY-MM') = $1`
		args = append(args, mes)
	}
	q += ` ORDER BY fecha DESC NULLS LAST LIMIT 300`

	rows, err := r.db.WithContext(ctx).Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*entities.Vivero{}
	for rows.Next() {
		var v entities.Vivero
		var fecha, lugarID sql.NullString
		if err := rows.Scan(
			&v.ID, &fecha, &v.Area, &v.Subproceso, &v.Etapa,
			&v.Descripcion, &v.Observaciones, &v.Responsables, &lugarID, &v.LugarLibre,
		); err != nil {
			return nil, err
		}
		v.Fecha = nullStringPtr(fecha)
		v.LugarID = nullStringPtr(lugarID)
		out = append(out, &v)
	}
	return out, rows.Err()
}

func (r *viveroRepository) Guardar(ctx context.Context, in entities.GuardarVivero) (*entities.Vivero, error) {
	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO vivero_registros (
		  id, fecha, area, subproceso, etapa, descripcion, observaciones, responsables, lugar_id, lugar_libre
		) VALUES (
		  $1, NULLIF($2, '')::date, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), $10
		)
		ON CONFLICT (id) DO UPDATE SET
		  fecha = EXCLUDED.fecha,
		  area = EXCLUDED.area,
		  subproceso = EXCLUDED.subproceso,
		  etapa = EXCLUDED.etapa,
		  descripcion = EXCLUDED.descripcion,
		  observaciones = EXCLUDED.observaciones,
		  responsables = EXCLUDED.responsables,
		  lugar_id = EXCLUDED.lugar_id,
		  lugar_libre = EXCLUDED.lugar_libre,
		  updated_at = now()`,
		in.ID, strings.TrimSpace(in.Fecha), in.Area, strings.TrimSpace(in.Subproceso), strings.TrimSpace(in.Etapa),
		strings.TrimSpace(in.Descripcion), strings.TrimSpace(in.Observaciones), strings.TrimSpace(in.Responsables),
		strings.TrimSpace(in.LugarID), strings.TrimSpace(in.LugarLibre),
	).Error
	if err != nil {
		return nil, err
	}

	var fPtr, lugPtr *string
	if strings.TrimSpace(in.Fecha) != "" {
		f := strings.TrimSpace(in.Fecha)
		fPtr = &f
	}
	if strings.TrimSpace(in.LugarID) != "" {
		l := strings.TrimSpace(in.LugarID)
		lugPtr = &l
	}
	return &entities.Vivero{
		ID:           in.ID,
		Fecha:        fPtr,
		Area:         in.Area,
		Subproceso:   in.Subproceso,
		Etapa:        in.Etapa,
		Descripcion:  in.Descripcion,
		Responsables: in.Responsables,
		LugarID:      lugPtr,
		LugarLibre:   in.LugarLibre,
	}, nil
}

func (r *viveroRepository) Archivar(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Exec(`UPDATE vivero_registros SET archivada_en = now() WHERE id = $1`, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainErrors.ErrLaborNoEncontrada
	}
	return nil
}
