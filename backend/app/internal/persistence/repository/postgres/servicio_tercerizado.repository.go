// Package postgres implements repository interfaces using GORM and PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

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

const ordenSelect = `
	SELECT o.id::text, o.actividad_id::text, o.empresa, o.referencia, o.frecuencia, o.estado, o.conformidad,
	       o.created_at, o.periodo_inicio, o.periodo_fin, o.reporte_proveedor,
	       o.empresa_id, ce.nombre, ce.codigo,
	       o.frecuencia_id, cf.nombre, cf.codigo
	FROM ordenes_servicio o
	LEFT JOIN catalogos ce ON ce.id = o.empresa_id AND ce.clase = 'empresa'
	LEFT JOIN catalogos cf ON cf.id = o.frecuencia_id AND cf.clase = 'frecuencia'`

func (r *servicioTercerizadoRepository) Listar(ctx context.Context) ([]*entities.ServicioTercerizado, error) {
	rows, err := r.db.WithContext(ctx).Raw(ordenSelect + ` ORDER BY o.created_at DESC LIMIT 200`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*entities.ServicioTercerizado{}
	for rows.Next() {
		o, err := scanOrden(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.anexarEvidencias(ctx, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *servicioTercerizadoRepository) ObtenerPorID(ctx context.Context, id string) (*entities.ServicioTercerizado, error) {
	rows, err := r.db.WithContext(ctx).Raw(ordenSelect+` WHERE o.id = $1`, id).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, domainErrors.ErrLaborNoEncontrada
	}
	o, err := scanOrden(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.anexarEvidencias(ctx, []*entities.ServicioTercerizado{o}); err != nil {
		return nil, err
	}
	return o, nil
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
			INSERT INTO ordenes_servicio (
			  id, actividad_id, empresa, empresa_id, referencia, frecuencia, frecuencia_id, conformidad
			) VALUES (
			  $1, $2, $3, CASE WHEN $4 THEN $5::bigint END, $6, $7, CASE WHEN $8 THEN $9::bigint END, $10
			)`,
			in.ID, in.ActividadID, in.Empresa, in.EmpresaID != nil, valorInt64(in.EmpresaID),
			in.Referencia, strings.TrimSpace(in.Frecuencia), in.FrecuenciaID != nil, valorInt64(in.FrecuenciaID),
			strings.TrimSpace(in.Conformidad),
		).Error; err != nil {
			return err
		}
		leida, err := scanOrden(tx.Raw(ordenSelect+` WHERE o.id = $1`, in.ID).Row())
		if err != nil {
			return err
		}
		o = *leida
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := r.anexarEvidencias(ctx, []*entities.ServicioTercerizado{&o}); err != nil {
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
		leida, err := scanOrden(tx.Raw(ordenSelect+` WHERE o.id = $1`, in.ID).Row())
		if err == sql.ErrNoRows {
			return domainErrors.ErrLaborNoEncontrada
		}
		if err != nil {
			return err
		}
		o = *leida
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := r.anexarEvidencias(ctx, []*entities.ServicioTercerizado{&o}); err != nil {
		return nil, err
	}
	return &o, nil
}

type filaOrden interface {
	Scan(dest ...any) error
}

func scanOrden(row filaOrden) (*entities.ServicioTercerizado, error) {
	var o entities.ServicioTercerizado
	var empresaID, frecuenciaID sql.NullInt64
	var empresaNombre, empresaCodigo, frecuenciaNombre, frecuenciaCodigo sql.NullString
	if err := row.Scan(
		&o.ID, &o.ActividadID, &o.Empresa, &o.Referencia, &o.Frecuencia, &o.Estado, &o.Conformidad,
		&o.CreatedAt, &o.PeriodoInicio, &o.PeriodoFin, &o.ReporteProveedor,
		&empresaID, &empresaNombre, &empresaCodigo,
		&frecuenciaID, &frecuenciaNombre, &frecuenciaCodigo,
	); err != nil {
		return nil, err
	}
	o.EmpresaID = nullInt64Ptr(empresaID)
	o.EmpresaCatalogo = strings.TrimSpace(empresaNombre.String)
	o.EmpresaCodigo = strings.TrimSpace(empresaCodigo.String)
	o.FrecuenciaID = nullInt64Ptr(frecuenciaID)
	o.FrecuenciaCatalogo = strings.TrimSpace(frecuenciaNombre.String)
	o.FrecuenciaCodigo = strings.TrimSpace(frecuenciaCodigo.String)
	o.Evidencias = []entities.EvidenciaOrden{}
	return &o, nil
}

func (r *servicioTercerizadoRepository) anexarEvidencias(ctx context.Context, items []*entities.ServicioTercerizado) error {
	if len(items) == 0 {
		return nil
	}
	porID := make(map[string]*entities.ServicioTercerizado, len(items))
	for _, it := range items {
		it.Evidencias = []entities.EvidenciaOrden{}
		porID[it.ID] = it
	}
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT orden_id::text, id::text, nombre, mime, bytes, COALESCE(nota, ''), created_at
		FROM evidencias
		WHERE orden_id IS NOT NULL
		ORDER BY created_at`).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var ordenID string
		var ev entities.EvidenciaOrden
		var when time.Time
		if err := rows.Scan(&ordenID, &ev.ID, &ev.Nombre, &ev.Mime, &ev.Bytes, &ev.Nota, &when); err != nil {
			return err
		}
		ev.CreatedAt = when
		if dest, ok := porID[ordenID]; ok {
			dest.Evidencias = append(dest.Evidencias, ev)
		}
	}
	return rows.Err()
}
