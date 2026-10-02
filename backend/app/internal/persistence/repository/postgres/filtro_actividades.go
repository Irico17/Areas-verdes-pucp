package postgres

import (
	"fmt"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// clausulasActividades arma el WHERE del listado.
// Si el rol es capataz, la cuadrilla pedida se ignora y queda la de la sesión.
func clausulasActividades(f entities.FiltroIntervenciones) (string, []any) {
	where := []string{"a.archivada_en IS NULL"}
	args := make([]any, 0, 8)
	n := 1
	add := func(clause string, value any) {
		where = append(where, strings.ReplaceAll(clause, "$?", fmt.Sprintf("$%d", n)))
		args = append(args, value)
		n++
	}

	if f.SoloAbiertas {
		where = append(where, "a.estado NOT IN ('cancelada', 'cerrada', 'archivada')")
	}
	if f.Rol == usecases.RolCapataz {
		id := strings.TrimSpace(f.CapatazID)
		where = append(where, fmt.Sprintf("(a.assigned_capataz_id = $%d AND (a.cuadrilla_id IS NULL OR a.cuadrilla_id = $%d))", n, n))
		args = append(args, id)
		n++
	} else if strings.TrimSpace(f.CuadrillaID) != "" {
		add("a.cuadrilla_id = $?", strings.TrimSpace(f.CuadrillaID))
	}
	if f.Estado != "" {
		add("a.estado = $?", f.Estado)
	}
	if f.Tipo != "" {
		add("a.tipo = $?", f.Tipo)
	}
	if f.ZonaSupervisionID != "" {
		add("a.zona_supervision_id = $?", f.ZonaSupervisionID)
	}
	if f.Origen != "" {
		add("a.origen = $?", f.Origen)
	}
	if f.Ejecutor != "" {
		add("a.ejecutor = $?", f.Ejecutor)
	}
	if f.NivelRiesgo != "" {
		add("a.nivel_riesgo = $?", f.NivelRiesgo)
	}
	if f.Desde != "" {
		add("a.fecha_programada >= $?::date", f.Desde)
	}
	if f.Hasta != "" {
		add("a.fecha_programada <= $?::date", f.Hasta)
	}
	if id := strings.TrimSpace(f.EjemplarID); id != "" {
		add(`EXISTS (
			SELECT 1 FROM actividad_avances av
			WHERE av.actividad_id = a.id
			  AND av.ejemplar_ref = $?
		)`, id)
	}
	if responsable := strings.TrimSpace(f.Responsable); responsable != "" {
		add(`(
			EXISTS (
				SELECT 1 FROM personal_labor pl
				WHERE pl.actividad_id = a.id AND pl.nombre_ficticio = $?
			)
			OR a.cuadrilla_id = $?
			OR a.assigned_capataz_id = $?
			OR EXISTS (
				SELECT 1 FROM cuadrillas q
				WHERE q.id = a.cuadrilla_id AND q.nombre_ficticio = $?
			)
			OR EXISTS (
				SELECT 1 FROM capataces cap
				WHERE cap.id = a.assigned_capataz_id AND cap.equipo = $?
			)
		)`, responsable)
	}
	if f.Sector != "" {
		add(`EXISTS (
			SELECT 1 FROM poligonos_cuadrilla p
			WHERE p.sector = $?
			  AND a.geom IS NOT NULL
			  AND p.geom IS NOT NULL
			  AND ST_Covers(p.geom, a.geom)
		)`, f.Sector)
	}
	return strings.Join(where, " AND "), args
}
