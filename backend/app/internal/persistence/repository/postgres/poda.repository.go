// Package postgres implements repository interfaces using GORM and PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type podaRepository struct {
	db *gorm.DB
}

// NewPodaRepository creates a new IPodaRepository instance.
func NewPodaRepository(db *gorm.DB) contracts.IPodaRepository {
	return &podaRepository{db: db}
}

func (r *podaRepository) Listar(ctx context.Context) ([]*entities.Poda, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id::text, codigo, COALESCE(codigo_externo, ''), tipo, tipo_actividad,
		       COALESCE(to_char(fecha_reporte, 'YYYY-MM-DD'), ''),
		       COALESCE(to_char(fecha_ejecucion, 'YYYY-MM-DD'), ''),
		       personal_ficticio, ubicacion, unidad, cantidad_pedida, cantidad_ejecutada,
		       prioridad, comentario, nombre_comun, nombre_cientifico
		FROM podas WHERE archivada_en IS NULL ORDER BY codigo`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*entities.Poda{}
	for rows.Next() {
		var p entities.Poda
		var codExt, fRep, fEjec sql.NullString
		if err := rows.Scan(
			&p.ID, &p.Codigo, &codExt, &p.Tipo, &p.TipoActividad,
			&fRep, &fEjec,
			&p.Personal, &p.Ubicacion, &p.Unidad, &p.CantidadPedida, &p.CantidadEjecutada,
			&p.Prioridad, &p.Comentario, &p.NombreComun, &p.NombreCientifico,
		); err != nil {
			return nil, err
		}
		p.CodigoExterno = nullStringPtr(codExt)
		p.FechaReporte = nullStringPtr(fRep)
		p.FechaEjecucion = nullStringPtr(fEjec)
		out = append(out, &p)
	}
	return out, rows.Err()
}

func (r *podaRepository) Guardar(ctx context.Context, in dto.GuardarPodaDTO) (*entities.Poda, error) {
	prioridad := strings.TrimSpace(in.Prioridad)
	codigo := strings.TrimSpace(in.Codigo)
	codExt := strings.TrimSpace(in.CodigoExterno)
	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO podas (
		  id, codigo, codigo_externo, tipo, tipo_actividad, fecha_reporte, fecha_ejecucion,
		  personal_ficticio, ubicacion, unidad, cantidad_pedida, cantidad_ejecutada,
		  prioridad, comentario, nombre_comun, nombre_cientifico, origen_ref
		) VALUES (
		  $1, $2, NULLIF($3, ''), $4, $5, NULLIF($6, '')::date, NULLIF($7, '')::date,
		  $8, $9, $10, $11, $12, $13, $14, $15, $16, $2
		)
		ON CONFLICT (codigo) DO UPDATE SET
		  codigo_externo = EXCLUDED.codigo_externo,
		  tipo = EXCLUDED.tipo,
		  tipo_actividad = EXCLUDED.tipo_actividad,
		  fecha_reporte = EXCLUDED.fecha_reporte,
		  fecha_ejecucion = EXCLUDED.fecha_ejecucion,
		  personal_ficticio = EXCLUDED.personal_ficticio,
		  ubicacion = EXCLUDED.ubicacion,
		  unidad = EXCLUDED.unidad,
		  cantidad_pedida = EXCLUDED.cantidad_pedida,
		  cantidad_ejecutada = EXCLUDED.cantidad_ejecutada,
		  prioridad = EXCLUDED.prioridad,
		  comentario = EXCLUDED.comentario,
		  nombre_comun = EXCLUDED.nombre_comun,
		  nombre_cientifico = EXCLUDED.nombre_cientifico,
		  updated_at = now()`,
		in.ID, codigo, codExt, strings.TrimSpace(in.Tipo),
		strings.TrimSpace(in.TipoActividad), strings.TrimSpace(in.FechaReporte), strings.TrimSpace(in.FechaEjecucion),
		strings.TrimSpace(in.Personal), strings.TrimSpace(in.Ubicacion), strings.TrimSpace(in.Unidad),
		in.CantidadPedida, in.CantidadEjecutada, prioridad, strings.TrimSpace(in.Comentario),
		strings.TrimSpace(in.NombreComun), strings.TrimSpace(in.NombreCientifico),
	).Error
	if err != nil {
		return nil, err
	}

	var codExtPtr *string
	if codExt != "" {
		codExtPtr = &codExt
	}
	return &entities.Poda{
		ID:                in.ID,
		Codigo:            codigo,
		CodigoExterno:     codExtPtr,
		Tipo:              in.Tipo,
		TipoActividad:     in.TipoActividad,
		Personal:          in.Personal,
		Ubicacion:         in.Ubicacion,
		Unidad:            in.Unidad,
		CantidadPedida:    in.CantidadPedida,
		CantidadEjecutada: in.CantidadEjecutada,
		Prioridad:         prioridad,
		Comentario:        in.Comentario,
	}, nil
}

func (r *podaRepository) Archivar(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Exec(`UPDATE podas SET archivada_en = now(), updated_at = now() WHERE id = $1 AND archivada_en IS NULL`, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainErrors.ErrLaborNoEncontrada
	}
	return nil
}
