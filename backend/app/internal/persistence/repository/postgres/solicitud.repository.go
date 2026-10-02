// Package postgres implements repository interfaces using GORM and PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type solicitudRepository struct {
	db *gorm.DB
}

const solicitudSelect = `
	SELECT s.id::text, s.codigo_externo, s.fuente, s.titulo, s.detalle, s.prioridad, s.estado,
	       s.lugar, s.lugar_id, l.nombre, s.lat, s.lon,
	       s.cantidad, s.cantidad_solicitada, s.cantidad_ejecutada,
	       s.actividad_id::text, s.created_at, s.updated_at, s.origen_ref, s.archivada_en
	FROM solicitudes s
	LEFT JOIN lugares l ON l.id = s.lugar_id`

// NewSolicitudRepository creates a new ISolicitudRepository.
func NewSolicitudRepository(db *gorm.DB) contracts.ISolicitudRepository {
	return &solicitudRepository{db: db}
}

func (r *solicitudRepository) Listar(ctx context.Context) ([]*entities.Solicitud, error) {
	rows, err := r.db.WithContext(ctx).Raw(solicitudSelect + ` ORDER BY s.created_at DESC LIMIT 200`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*entities.Solicitud{}
	for rows.Next() {
		item, err := scanSolicitud(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *solicitudRepository) ObtenerPorID(ctx context.Context, id string) (*entities.Solicitud, error) {
	rows, err := r.db.WithContext(ctx).Raw(solicitudSelect+` WHERE s.id = $1`, id).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domainErrors.ErrLaborNoEncontrada
	}
	item, err := scanSolicitud(rows)
	if err != nil {
		return nil, err
	}
	return item, rows.Err()
}

func (r *solicitudRepository) Crear(ctx context.Context, in entities.NuevaSolicitud) (*entities.Solicitud, error) {
	if in.CantidadSolicitada == nil && in.Cantidad != 0 {
		n := in.Cantidad
		in.CantidadSolicitada = &n
	}
	if in.Cantidad == 0 && in.CantidadSolicitada != nil {
		in.Cantidad = *in.CantidadSolicitada
	}
	if err := lugarActivo(r.db.WithContext(ctx), in.LugarID); err != nil {
		return nil, err
	}
	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO solicitudes (
		  id, codigo_externo, fuente, titulo, detalle, prioridad, estado,
		  lugar, cantidad, cantidad_solicitada, cantidad_ejecutada,
		  lugar_id, lat, lon, actividad_id
		) VALUES (
		  $1, NULLIF($2, ''), $3, $4, $5, $6, COALESCE(NULLIF($7, ''), 'por_iniciar'),
		  NULLIF($8, ''), NULLIF($9, 0),
		  CASE WHEN $10 THEN $11::integer END,
		  CASE WHEN $12 THEN $13::integer END,
		  CASE WHEN $14 THEN $15::bigint END,
		  CASE WHEN $16 THEN $17::double precision END,
		  CASE WHEN $18 THEN $19::double precision END,
		  NULLIF($20, '')::uuid
		)`,
		in.ID, in.CodigoExterno, in.Fuente, in.Titulo, strings.TrimSpace(in.Detalle), in.Prioridad, in.Estado,
		strings.TrimSpace(in.Lugar), in.Cantidad,
		in.CantidadSolicitada != nil, valorEntero(in.CantidadSolicitada),
		in.CantidadEjecutada != nil, valorEntero(in.CantidadEjecutada),
		in.LugarID != nil, valorInt64(in.LugarID),
		in.Lat != nil, valorFloat(in.Lat),
		in.Lon != nil, valorFloat(in.Lon),
		strings.TrimSpace(in.ActividadID),
	).Error
	if err != nil {
		return nil, err
	}
	return r.ObtenerPorID(ctx, in.ID)
}

func (r *solicitudRepository) Editar(ctx context.Context, in entities.EditarSolicitud) (*entities.Solicitud, error) {
	if err := lugarActivo(r.db.WithContext(ctx), in.LugarID); err != nil {
		return nil, err
	}
	res := r.db.WithContext(ctx).Exec(`
		UPDATE solicitudes SET
		  codigo_externo = CASE WHEN btrim($2) = '' THEN codigo_externo ELSE $2 END,
		  titulo = $3,
		  detalle = $4,
		  prioridad = $5,
		  lugar = NULLIF($6, ''),
		  estado = CASE WHEN btrim($7) = '' THEN estado ELSE $7 END,
		  cantidad_solicitada = CASE WHEN $8 THEN $9::integer ELSE cantidad_solicitada END,
		  cantidad_ejecutada = CASE WHEN $10 THEN $11::integer ELSE cantidad_ejecutada END,
		  lugar_id = CASE WHEN $12 THEN $13::bigint ELSE lugar_id END,
		  lat = CASE WHEN $14 THEN $15::double precision ELSE lat END,
		  lon = CASE WHEN $16 THEN $17::double precision ELSE lon END,
		  updated_at = now()
		WHERE id = $1 AND archivada_en IS NULL`,
		in.ID, strings.TrimSpace(in.CodigoExterno), in.Titulo, strings.TrimSpace(in.Detalle),
		strings.TrimSpace(in.Prioridad), strings.TrimSpace(in.Lugar), strings.TrimSpace(in.Estado),
		in.CantidadSolicitada != nil, valorEntero(in.CantidadSolicitada),
		in.CantidadEjecutada != nil, valorEntero(in.CantidadEjecutada),
		in.LugarID != nil, valorInt64(in.LugarID),
		in.Lat != nil, valorFloat(in.Lat),
		in.Lon != nil, valorFloat(in.Lon),
	)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, domainErrors.ErrLaborNoEncontrada
	}
	return r.ObtenerPorID(ctx, in.ID)
}

type filaSolicitud interface {
	Scan(dest ...any) error
}

func scanSolicitud(row filaSolicitud) (*entities.Solicitud, error) {
	var item entities.Solicitud
	var codigo, lugar, lugarNombre, actividad, origen sql.NullString
	var lugarID, cantidad, solicitada, ejecutada sql.NullInt64
	var lat, lon sql.NullFloat64
	var created, updated time.Time
	var archivada sql.NullTime
	if err := row.Scan(
		&item.ID, &codigo, &item.Fuente, &item.Titulo, &item.Detalle, &item.Prioridad, &item.Estado,
		&lugar, &lugarID, &lugarNombre, &lat, &lon,
		&cantidad, &solicitada, &ejecutada,
		&actividad, &created, &updated, &origen, &archivada,
	); err != nil {
		return nil, err
	}
	item.CodigoExterno = nullStringPtr(codigo)
	item.Lugar = nullStringPtr(lugar)
	item.LugarID = nullInt64Ptr(lugarID)
	item.LugarNombre = nullStringPtr(lugarNombre)
	item.Lat = nullFloatPtr(lat)
	item.Lon = nullFloatPtr(lon)
	item.Cantidad = nullIntPtr(cantidad)
	item.CantidadSolicitada = nullIntPtr(solicitada)
	item.CantidadEjecutada = nullIntPtr(ejecutada)
	item.ActividadID = nullStringPtr(actividad)
	item.OrigenRef = nullStringPtr(origen)
	item.CreatedAt = created
	item.UpdatedAt = updated
	if archivada.Valid {
		t := archivada.Time
		item.ArchivadaEn = &t
	}
	return &item, nil
}

func lugarActivo(db *gorm.DB, id *int64) error {
	if id == nil {
		return nil
	}
	var n int
	if err := db.Raw(`SELECT count(*) FROM lugares WHERE id = $1 AND activo`, *id).Scan(&n).Error; err != nil {
		return err
	}
	if n != 1 {
		return domainErrors.InputError{Reason: "el lugar no está en el catálogo"}
	}
	return nil
}

func valorEntero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func valorInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}

func valorFloat(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func nullIntPtr(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	n := int(v.Int64)
	return &n
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	n := v.Int64
	return &n
}

func nullFloatPtr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	n := v.Float64
	return &n
}

func nullStringPtr(v sql.NullString) *string {
	if !v.Valid || strings.TrimSpace(v.String) == "" {
		return nil
	}
	s := v.String
	return &s
}
