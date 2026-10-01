// Package postgres implements repository interfaces using GORM and PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type servicioTercerizadoRepository struct {
	db *gorm.DB
}

// NewServicioTercerizadoRepository creates a new IServicioTercerizadoRepository.
func NewServicioTercerizadoRepository(db *gorm.DB) contracts.IServicioTercerizadoRepository {
	return &servicioTercerizadoRepository{db: db}
}

func (r *servicioTercerizadoRepository) Listar(ctx context.Context) ([]*entities.ServicioTercerizado, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id::text, actividad_id::text, empresa, referencia, frecuencia, estado, conformidad,
		       created_at, periodo_inicio, periodo_fin, reporte_proveedor
		FROM ordenes_servicio ORDER BY created_at DESC LIMIT 200`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*entities.ServicioTercerizado{}
	for rows.Next() {
		var o entities.ServicioTercerizado
		if err := rows.Scan(
			&o.ID, &o.ActividadID, &o.Empresa, &o.Referencia, &o.Frecuencia, &o.Estado, &o.Conformidad,
			&o.CreatedAt, &o.PeriodoInicio, &o.PeriodoFin, &o.ReporteProveedor,
		); err != nil {
			return nil, err
		}
		out = append(out, &o)
	}
	return out, rows.Err()
}

func (r *servicioTercerizadoRepository) ObtenerPorID(ctx context.Context, id string) (*entities.ServicioTercerizado, error) {
	var o entities.ServicioTercerizado
	err := r.db.WithContext(ctx).Raw(`
		SELECT id::text, actividad_id::text, empresa, referencia, frecuencia, estado, conformidad,
		       created_at, periodo_inicio, periodo_fin, reporte_proveedor
		FROM ordenes_servicio WHERE id = $1`, id).Row().Scan(
		&o.ID, &o.ActividadID, &o.Empresa, &o.Referencia, &o.Frecuencia, &o.Estado, &o.Conformidad,
		&o.CreatedAt, &o.PeriodoInicio, &o.PeriodoFin, &o.ReporteProveedor,
	)
	if err == sql.ErrNoRows {
		return nil, domainErrors.ErrLaborNoEncontrada
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *servicioTercerizadoRepository) Crear(ctx context.Context, in entities.NuevaOrdenServicio, actorRol, capatazID string) (*entities.ServicioTercerizado, error) {
	var o entities.ServicioTercerizado
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row, err := lockActividadDB(tx, in.ActividadID)
		if err != nil {
			return err
		}
		if row.Archivada {
			return domainErrors.InputError{Reason: "la labor está archivada"}
		}
		if actorRol == string(enums.RolCapataz) {
			if row.Capataz != strings.TrimSpace(capatazID) {
				return domainErrors.ForbiddenError{Reason: "el capataz solo puede registrar órdenes en sus propias labores"}
			}
		}
		if row.Ejecutor != string(enums.EjecutorTercerizada) {
			return domainErrors.InputError{Reason: "la orden solo se vincula a una labor tercerizada"}
		}
		if err := tx.Exec(`
			INSERT INTO ordenes_servicio (id, actividad_id, empresa, referencia, frecuencia)
			VALUES ($1, $2, $3, $4, $5)`,
			in.ID, in.ActividadID, in.Empresa, in.Referencia, strings.TrimSpace(in.Frecuencia),
		).Error; err != nil {
			return err
		}
		return tx.Raw(`
			SELECT id::text, actividad_id::text, empresa, referencia, frecuencia, estado, conformidad,
			       created_at, periodo_inicio, periodo_fin, reporte_proveedor
			FROM ordenes_servicio WHERE id = $1`, in.ID).Row().Scan(
			&o.ID, &o.ActividadID, &o.Empresa, &o.Referencia, &o.Frecuencia, &o.Estado, &o.Conformidad,
			&o.CreatedAt, &o.PeriodoInicio, &o.PeriodoFin, &o.ReporteProveedor,
		)
	})
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *servicioTercerizadoRepository) Editar(ctx context.Context, in entities.EditarOrdenServicio, actorRol, capatazID string) (*entities.ServicioTercerizado, error) {
	var o entities.ServicioTercerizado
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var actividadID string
		err := tx.Raw(`SELECT actividad_id::text FROM ordenes_servicio WHERE id = $1 FOR UPDATE`, in.ID).Row().Scan(&actividadID)
		if err == sql.ErrNoRows {
			return domainErrors.ErrLaborNoEncontrada
		}
		if err != nil {
			return err
		}
		row, err := lockActividadDB(tx, actividadID)
		if err != nil {
			return err
		}
		if row.Archivada {
			return domainErrors.InputError{Reason: "la labor está archivada"}
		}
		if actorRol == string(enums.RolCapataz) {
			if row.Capataz != strings.TrimSpace(capatazID) {
				return domainErrors.ForbiddenError{Reason: "el capataz solo puede editar órdenes de sus propias labores"}
			}
		}
		res := tx.Exec(`
			UPDATE ordenes_servicio SET
			  conformidad = $2,
			  periodo_inicio = NULLIF($3, '')::date,
			  periodo_fin = NULLIF($4, '')::date,
			  reporte_proveedor = $5,
			  estado = $6
			WHERE id = $1`,
			in.ID, strings.TrimSpace(in.Conformidad), strings.TrimSpace(in.PeriodoInicio),
			strings.TrimSpace(in.PeriodoFin), strings.TrimSpace(in.ReporteProveedor), in.Estado,
		)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domainErrors.ErrLaborNoEncontrada
		}
		err = tx.Raw(`
			SELECT id::text, actividad_id::text, empresa, referencia, frecuencia, estado, conformidad,
			       created_at, periodo_inicio, periodo_fin, reporte_proveedor
			FROM ordenes_servicio WHERE id = $1`, in.ID).Row().Scan(
			&o.ID, &o.ActividadID, &o.Empresa, &o.Referencia, &o.Frecuencia, &o.Estado, &o.Conformidad,
			&o.CreatedAt, &o.PeriodoInicio, &o.PeriodoFin, &o.ReporteProveedor,
		)
		if err == sql.ErrNoRows {
			return domainErrors.ErrLaborNoEncontrada
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return &o, nil
}
