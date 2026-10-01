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

// NewSolicitudRepository creates a new ISolicitudRepository.
func NewSolicitudRepository(db *gorm.DB) contracts.ISolicitudRepository {
	return &solicitudRepository{db: db}
}

func (r *solicitudRepository) Listar(ctx context.Context) ([]*entities.Solicitud, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id::text, codigo_externo, fuente, titulo, detalle, prioridad, estado,
		       lugar, cantidad, actividad_id::text, created_at, updated_at, origen_ref, archivada_en
		FROM solicitudes ORDER BY created_at DESC LIMIT 200`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*entities.Solicitud{}
	for rows.Next() {
		var item entities.Solicitud
		var codigo, lugar, actividad, origen sql.NullString
		var cantidad sql.NullInt64
		var created, updated time.Time
		var archivada sql.NullTime

		if err := rows.Scan(
			&item.ID, &codigo, &item.Fuente, &item.Titulo, &item.Detalle, &item.Prioridad, &item.Estado,
			&lugar, &cantidad, &actividad, &created, &updated, &origen, &archivada,
		); err != nil {
			return nil, err
		}

		item.CodigoExterno = nullStringPtr(codigo)
		item.Lugar = nullStringPtr(lugar)
		item.ActividadID = nullStringPtr(actividad)
		item.OrigenRef = nullStringPtr(origen)
		if cantidad.Valid {
			n := int(cantidad.Int64)
			item.Cantidad = &n
		}
		item.CreatedAt = created
		item.UpdatedAt = updated
		if archivada.Valid {
			t := archivada.Time
			item.ArchivadaEn = &t
		}
		out = append(out, &item)
	}
	return out, rows.Err()
}

func (r *solicitudRepository) ObtenerPorID(ctx context.Context, id string) (*entities.Solicitud, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id::text, codigo_externo, fuente, titulo, detalle, prioridad, estado,
		       lugar, cantidad, actividad_id::text, created_at, updated_at, origen_ref, archivada_en
		FROM solicitudes WHERE id = $1`, id).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, domainErrors.ErrLaborNoEncontrada
	}

	var item entities.Solicitud
	var codigo, lugar, actividad, origen sql.NullString
	var cantidad sql.NullInt64
	var created, updated time.Time
	var archivada sql.NullTime

	if err := rows.Scan(
		&item.ID, &codigo, &item.Fuente, &item.Titulo, &item.Detalle, &item.Prioridad, &item.Estado,
		&lugar, &cantidad, &actividad, &created, &updated, &origen, &archivada,
	); err != nil {
		return nil, err
	}

	item.CodigoExterno = nullStringPtr(codigo)
	item.Lugar = nullStringPtr(lugar)
	item.ActividadID = nullStringPtr(actividad)
	item.OrigenRef = nullStringPtr(origen)
	if cantidad.Valid {
		n := int(cantidad.Int64)
		item.Cantidad = &n
	}
	item.CreatedAt = created
	item.UpdatedAt = updated
	if archivada.Valid {
		t := archivada.Time
		item.ArchivadaEn = &t
	}
	return &item, rows.Err()
}

func (r *solicitudRepository) Crear(ctx context.Context, in entities.NuevaSolicitud) (*entities.Solicitud, error) {
	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO solicitudes (
		  id, codigo_externo, fuente, titulo, detalle, prioridad, lugar, cantidad, actividad_id
		) VALUES (
		  $1, NULLIF($2, ''), $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, 0), NULLIF($9, '')::uuid
		)`,
		in.ID, in.CodigoExterno, in.Fuente, in.Titulo, strings.TrimSpace(in.Detalle),
		in.Prioridad, strings.TrimSpace(in.Lugar), in.Cantidad, strings.TrimSpace(in.ActividadID),
	).Error
	if err != nil {
		return nil, err
	}
	return r.ObtenerPorID(ctx, in.ID)
}

func (r *solicitudRepository) Editar(ctx context.Context, in entities.EditarSolicitud) (*entities.Solicitud, error) {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE solicitudes SET
		  codigo_externo = CASE WHEN btrim($2) = '' THEN codigo_externo ELSE $2 END,
		  titulo = $3,
		  detalle = $4,
		  prioridad = $5,
		  lugar = NULLIF($6, ''),
		  updated_at = now()
		WHERE id = $1 AND archivada_en IS NULL`,
		in.ID, strings.TrimSpace(in.CodigoExterno), in.Titulo, strings.TrimSpace(in.Detalle),
		strings.TrimSpace(in.Prioridad), strings.TrimSpace(in.Lugar),
	)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, domainErrors.ErrLaborNoEncontrada
	}
	return r.ObtenerPorID(ctx, in.ID)
}

func nullStringPtr(v sql.NullString) *string {
	if !v.Valid || strings.TrimSpace(v.String) == "" {
		return nil
	}
	s := v.String
	return &s
}
