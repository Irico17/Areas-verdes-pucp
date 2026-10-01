// Package postgres implements repository contracts using PostgreSQL and GORM.
package postgres

import (
	"context"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

type reporteRepository struct {
	db *gorm.DB
}

// NewReporteRepository creates a new IReporteRepository instance.
func NewReporteRepository(db *gorm.DB) contracts.IReporteRepository {
	return &reporteRepository{db: db}
}

// ObtenerReporte generates the full report with counts, rows and pending indicators.
func (r *reporteRepository) ObtenerReporte(ctx context.Context, f entities.FiltroReporte) (*entities.Reporte, error) {
	out := &entities.Reporte{
		Aviso:      "Conteos de labores del filtro. Cobertura, rendimiento y métricas de proveedor: definición pendiente.",
		PorEstado:  []entities.ConteoReporte{},
		Filas:      []entities.FilaReporte{},
		Pendientes: entities.HuecosIndicador(),
	}

	where, args, err := ClausulasReporte(f, false)
	if err != nil {
		return out, err
	}

	conteoQuery := `
		SELECT a.estado, count(*)::int AS n
		FROM actividades a
		LEFT JOIN lugares l ON l.id = a.lugar_id
		LEFT JOIN zonas_supervision z ON z.id = COALESCE(a.zona_supervision_id, l.zona_supervision_id)
		LEFT JOIN cuadrillas q ON q.id = a.cuadrilla_id
		WHERE ` + strings.Join(where, " AND ") + `
		GROUP BY a.estado ORDER BY a.estado`

	var modelosConteo []models.ConteoReporteModel
	if err := r.db.WithContext(ctx).Raw(conteoQuery, args...).Scan(&modelosConteo).Error; err != nil {
		return out, err
	}
	for _, c := range modelosConteo {
		out.PorEstado = append(out.PorEstado, mapper.ToDomainConteoReporte(c))
	}

	err = r.RecorrerFilas(ctx, f, func(fila entities.FilaReporte) error {
		out.Filas = append(out.Filas, fila)
		return nil
	})
	return out, err
}

// ValidarFiltro validates the report filter criteria without running queries.
func (r *reporteRepository) ValidarFiltro(f entities.FiltroReporte) error {
	_, _, err := ClausulasReporte(f, true)
	return err
}

// RecorrerFilas streams report rows one by one through the provided callback.
func (r *reporteRepository) RecorrerFilas(ctx context.Context, f entities.FiltroReporte, fn func(entities.FilaReporte) error) error {
	where, args, err := ClausulasReporte(f, true)
	if err != nil {
		return err
	}

	q := `
		SELECT a.id::text, a.titulo, a.tipo, a.estado, a.ejecutor,
		       COALESCE(c.equipo, ''),
		       COALESCE(z.codigo, a.zona_feature_id, ''),
		       COALESCE(s.codigo_externo, ''), COALESCE(s.fuente, ''), a.created_at,
		       COALESCE(cl.nombre, a.clase_codigo, ''),
		       COALESCE(NULLIF(l.nombre, ''), a.lugar_libre, ''),
		       COALESCE(q.nombre_ficticio, ''),
		       a.fecha_solicitud, a.fecha_atencion
		FROM actividades a
		LEFT JOIN capataces c ON c.id = a.assigned_capataz_id
		LEFT JOIN solicitudes s ON s.actividad_id = a.id
		LEFT JOIN lugares l ON l.id = a.lugar_id
		LEFT JOIN zonas_supervision z ON z.id = COALESCE(a.zona_supervision_id, l.zona_supervision_id)
		LEFT JOIN cuadrillas q ON q.id = a.cuadrilla_id
		LEFT JOIN catalogos cl ON cl.clase = 'clase_actividad' AND cl.codigo = a.clase_codigo
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY a.created_at DESC`

	rows, err := r.db.WithContext(ctx).Raw(q, args...).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		m, err := scanFilaReporte(rows)
		if err != nil {
			return err
		}
		if err := fn(mapper.ToDomainFilaReporte(m)); err != nil {
			return err
		}
	}
	return rows.Err()
}

func scanFilaReporte(rows interface {
	Scan(dest ...any) error
}) (models.FilaReporteModel, error) {
	var m models.FilaReporteModel
	if err := rows.Scan(
		&m.ID, &m.Titulo, &m.Tipo, &m.Estado, &m.Ejecutor, &m.Equipo, &m.Zona,
		&m.CodigoExterno, &m.Fuente, &m.CreatedAt, &m.Clase, &m.Lugar, &m.Cuadrilla,
		&m.FechaSolicitud, &m.FechaAtencion,
	); err != nil {
		return models.FilaReporteModel{}, err
	}
	return m, nil
}

// ClausulasReporte builds the WHERE clauses and arguments for reporting queries.
// conEstado controls whether f.Estado is included in the condition.
func ClausulasReporte(f entities.FiltroReporte, conEstado bool) ([]string, []any, error) {
	where := []string{"a.archivada_en IS NULL"}
	args := []any{}
	n := 1
	add := func(clause string, val any) {
		where = append(where, strings.ReplaceAll(clause, "$?", "$"+strconv.Itoa(n)))
		args = append(args, val)
		n++
	}
	if conEstado && f.Estado != "" {
		add("$? = a.estado", f.Estado)
	}
	if f.Desde != "" {
		if _, err := time.Parse("2006-01-02", f.Desde); err != nil {
			return nil, nil, domainErrors.InputError{Reason: "desde usa AAAA-MM-DD"}
		}
		add(`COALESCE(a.fecha_atencion, a.fecha_solicitud, a.created_at::date) >= $?::date`, f.Desde)
	}
	if f.Hasta != "" {
		if _, err := time.Parse("2006-01-02", f.Hasta); err != nil {
			return nil, nil, domainErrors.InputError{Reason: "hasta usa AAAA-MM-DD"}
		}
		add(`COALESCE(a.fecha_atencion, a.fecha_solicitud, a.created_at::date) <= $?::date`, f.Hasta)
	}
	if zona := strings.TrimSpace(f.Zona); zona != "" {
		add(`(
			z.codigo = $?
			OR a.zona_supervision_id::text = $?
			OR l.zona_supervision_id::text = $?
			OR a.zona_feature_id = $?
		)`, zona)
	}
	if cuad := strings.TrimSpace(f.Cuadrilla); cuad != "" {
		add(`(a.cuadrilla_id = $? OR q.nombre_ficticio = $?)`, cuad)
	}
	if origen := strings.TrimSpace(f.Origen); origen != "" {
		add("a.origen = $?", origen)
	}
	return where, args, nil
}
