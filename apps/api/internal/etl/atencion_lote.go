package etl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

// cargarAtencion hace upsert de labores, podas y vivero. Si la fuente no viene, no toca esa tabla.
func cargarAtencion(tx *gorm.DB, fuentes fuentesLote, loteID int64, rep *ReporteLote) error {
	if len(fuentes.Monitoreo) == 0 && len(fuentes.Poda) == 0 && len(fuentes.Vivero) == 0 {
		return nil
	}
	dir, err := os.MkdirTemp("", "lote-atencion")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	lugares, err := mapaLugares(tx)
	if err != nil {
		return err
	}
	tabla := NuevaTabla()
	if len(fuentes.Monitoreo) > 0 {
		path := filepath.Join(dir, "monitoreo.csv")
		if err := os.WriteFile(path, fuentes.Monitoreo, 0o644); err != nil {
			return err
		}
		rows, informe, err := ImportarMonitoreo(path, tabla)
		if err != nil {
			return err
		}
		if err := upsertLabores(tx, rows, loteID, rep); err != nil {
			return err
		}
		for _, aviso := range informe.Avisos {
			rep.Avisos = append(rep.Avisos, fmt.Sprintf("monitoreo fila %d: %s", aviso.Fila, aviso.Motivo))
		}
		rep.Cargados["labores_sin_estado"] = informe.SinEstado
	}
	if len(fuentes.Poda) > 0 {
		path := filepath.Join(dir, "poda.csv")
		if err := os.WriteFile(path, fuentes.Poda, 0o644); err != nil {
			return err
		}
		rows, errores, err := ImportarPoda(path, tabla, lugares)
		if err != nil {
			return err
		}
		if err := upsertPodas(tx, rows, loteID, rep); err != nil {
			return err
		}
		for _, aviso := range errores {
			rep.Avisos = append(rep.Avisos, fmt.Sprintf("poda fila %d: %s", aviso.Fila, aviso.Motivo))
		}
	}
	if len(fuentes.Vivero) > 0 {
		path := filepath.Join(dir, "vivero.csv")
		if err := os.WriteFile(path, fuentes.Vivero, 0o644); err != nil {
			return err
		}
		rows, informe, err := ImportarVivero(path, tabla, lugares)
		if err != nil {
			return err
		}
		if err := upsertVivero(tx, rows, loteID, rep); err != nil {
			return err
		}
		if informe.ViveroDelta != "" {
			rep.Avisos = append(rep.Avisos, informe.ViveroDelta)
		}
		rep.Cargados["vivero_lugar_libre"] = informe.LugarLibre
	}
	return nil
}

func mapaLugares(tx *gorm.DB) (map[string]string, error) {
	type fila struct {
		ID   string
		Norm string
	}
	var rows []fila
	err := tx.Raw(`SELECT id::text AS id, nombre_norm AS norm FROM lugares WHERE activo`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for _, row := range rows {
		out[row.Norm] = row.ID
	}
	return out, nil
}

func tipoOperativo(clase, tipo string) string {
	n := normalizarPersona(clase + " " + tipo)
	switch {
	case strings.Contains(n, "riego"):
		return "riego"
	case strings.Contains(n, "poda"):
		return "poda"
	case strings.Contains(n, "inspec"):
		return "inspeccion"
	case strings.Contains(n, "incid"):
		return "incidencia"
	default:
		return "limpieza"
	}
}

func upsertLabores(tx *gorm.DB, rows []LaborImportada, loteID int64, rep *ReporteLote) error {
	for _, item := range rows {
		var lat, lon any
		if item.Lat != nil && item.Lon != nil {
			lat, lon = *item.Lat, *item.Lon
		}
		var fs, fa any
		if item.FechaSolicitud != nil {
			fs = item.FechaSolicitud.Format("2006-01-02")
		}
		if item.FechaAtencion != nil {
			fa = item.FechaAtencion.Format("2006-01-02")
		}
		cuadID, err := asegurarCuadrilla(tx, item.Cuadrilla)
		if err != nil {
			return err
		}
		id := IDOrigen("labor", item.OrigenRef)
		var rowID string
		err = tx.Raw(`
			INSERT INTO actividades (
			  id, tipo, estado, titulo, detalle, geom, ejecutor,
			  fecha_solicitud, fecha_atencion, lugar_id, cuadrilla_id,
			  clase_codigo, tipo_codigo, comentario, lugar_libre, origen_ref, origen
			) VALUES (
			  $1, $2, $3, $4, $5,
			  CASE WHEN $6::float8 IS NULL THEN NULL ELSE ST_SetSRID(ST_MakePoint($7, $6), 4326) END,
			  'propia',
			  $8::date, $9::date,
			  (SELECT id FROM lugares WHERE nombre_norm = $10 LIMIT 1),
			  NULLIF($11, ''),
			  NULLIF($12, ''), NULLIF($13, ''), $14, $15, $16, 'monitoreo'
			)
			ON CONFLICT (origen_ref) WHERE origen_ref IS NOT NULL AND origen_ref <> ''
			DO UPDATE SET
			  tipo = EXCLUDED.tipo,
			  estado = EXCLUDED.estado,
			  titulo = EXCLUDED.titulo,
			  detalle = EXCLUDED.detalle,
			  geom = EXCLUDED.geom,
			  fecha_solicitud = EXCLUDED.fecha_solicitud,
			  fecha_atencion = EXCLUDED.fecha_atencion,
			  lugar_id = EXCLUDED.lugar_id,
			  cuadrilla_id = EXCLUDED.cuadrilla_id,
			  clase_codigo = EXCLUDED.clase_codigo,
			  tipo_codigo = EXCLUDED.tipo_codigo,
			  comentario = EXCLUDED.comentario,
			  lugar_libre = EXCLUDED.lugar_libre,
			  updated_at = now()
			RETURNING id::text`,
			id, tipoOperativo(item.Clase, item.Tipo), item.Estado, item.Titulo, item.Detalle,
			lat, lon, fs, fa, normalizarPersona(item.Lugar), cuadID,
			slug(item.Clase), slug(item.Tipo), item.Comentario, item.Lugar, item.OrigenRef,
		).Scan(&rowID).Error
		if err != nil {
			return fmt.Errorf("labor %s: %w", item.OrigenRef, err)
		}
		despues, _ := json.Marshal(map[string]any{"origen_ref": item.OrigenRef, "estado": item.Estado})
		if err := auditar(tx, "actividades", rowID, nil, despues, loteID); err != nil {
			return err
		}
	}
	rep.Cargados["labores"] = len(rows)
	return nil
}

func asegurarCuadrilla(tx *gorm.DB, nombre string) (string, error) {
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return "", nil
	}
	id := "cf-" + slug(nombre)
	if id == "cf-" {
		return "", nil
	}
	if err := tx.Exec(`
		INSERT INTO cuadrillas (id, nombre_ficticio, turno, activo)
		VALUES ($1, $2, 'manana', TRUE)
		ON CONFLICT (id) DO UPDATE SET nombre_ficticio = EXCLUDED.nombre_ficticio`,
		id, nombre).Error; err != nil {
		return "", err
	}
	return id, nil
}

func upsertPodas(tx *gorm.DB, rows []PodaImportada, loteID int64, rep *ReporteLote) error {
	for _, item := range rows {
		var fr, fe any
		if item.FechaReporte != nil {
			fr = item.FechaReporte.Format("2006-01-02")
		}
		if item.FechaEjecucion != nil {
			fe = item.FechaEjecucion.Format("2006-01-02")
		}
		id := IDOrigen("poda", item.OrigenRef)
		var rowID string
		err := tx.Raw(`
			INSERT INTO podas (
			  id, codigo, codigo_externo, tipo, tipo_actividad, fecha_reporte, fecha_ejecucion,
			  personal_ficticio, ubicacion, unidad, cantidad_pedida, cantidad_ejecutada,
			  tipo_vegetacion, nombre_comun, nombre_cientifico, prioridad, comentario, origen_ref
			) VALUES (
			  $1, $2, NULLIF($3, ''), $4, $5, $6::date, $7::date,
			  $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
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
			  tipo_vegetacion = EXCLUDED.tipo_vegetacion,
			  nombre_comun = EXCLUDED.nombre_comun,
			  nombre_cientifico = EXCLUDED.nombre_cientifico,
			  prioridad = EXCLUDED.prioridad,
			  comentario = EXCLUDED.comentario,
			  origen_ref = EXCLUDED.origen_ref,
			  updated_at = now()
			RETURNING id::text`,
			id, item.Codigo, item.CodigoExterno, item.Tipo, item.TipoActividad, fr, fe,
			item.Personal, item.Ubicacion, item.Unidad, item.CantidadPedida, item.CantidadEjecutada,
			item.TipoVegetacion, item.NombreComun, item.NombreCientifico, item.Prioridad, item.Comentario, item.OrigenRef,
		).Scan(&rowID).Error
		if err != nil {
			return fmt.Errorf("poda %s: %w", item.Codigo, err)
		}
		despues, _ := json.Marshal(map[string]any{"origen_ref": item.OrigenRef, "codigo": item.Codigo})
		if err := auditar(tx, "podas", rowID, nil, despues, loteID); err != nil {
			return err
		}
	}
	rep.Cargados["podas"] = len(rows)
	return nil
}

func upsertVivero(tx *gorm.DB, rows []ViveroImportado, loteID int64, rep *ReporteLote) error {
	for _, item := range rows {
		if err := altaCatalogoVivero(tx, "subproceso", item.Subproceso); err != nil {
			return err
		}
		if err := altaCatalogoVivero(tx, "etapa", item.Etapa); err != nil {
			return err
		}
		var fecha any
		if item.Fecha != nil {
			fecha = item.Fecha.Format("2006-01-02")
		}
		id := IDOrigen("vivero", item.OrigenRef)
		var rowID string
		err := tx.Raw(`
			INSERT INTO vivero_registros (
			  id, fecha, area, subproceso, etapa, descripcion, observaciones,
			  responsables, lugar_id, lugar_libre, origen_ref
			) VALUES (
			  $1, $2::date, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), $10, $11
			)
			ON CONFLICT (origen_ref) WHERE origen_ref IS NOT NULL AND origen_ref <> ''
			DO UPDATE SET
			  fecha = EXCLUDED.fecha,
			  area = EXCLUDED.area,
			  subproceso = EXCLUDED.subproceso,
			  etapa = EXCLUDED.etapa,
			  descripcion = EXCLUDED.descripcion,
			  observaciones = EXCLUDED.observaciones,
			  responsables = EXCLUDED.responsables,
			  lugar_id = EXCLUDED.lugar_id,
			  lugar_libre = EXCLUDED.lugar_libre,
			  updated_at = now()
			RETURNING id::text`,
			id, fecha, item.Area, item.Subproceso, item.Etapa, item.Descripcion, item.Observaciones,
			item.Responsables, item.LugarID, item.LugarLibre, item.OrigenRef,
		).Scan(&rowID).Error
		if err != nil {
			return fmt.Errorf("vivero %s: %w", item.OrigenRef, err)
		}
		despues, _ := json.Marshal(map[string]any{"origen_ref": item.OrigenRef, "area": item.Area})
		if err := auditar(tx, "vivero_registros", rowID, nil, despues, loteID); err != nil {
			return err
		}
	}
	rep.Cargados["vivero"] = len(rows)
	return nil
}

func altaCatalogoVivero(tx *gorm.DB, clase, nombre string) error {
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return nil
	}
	return tx.Exec(`
		INSERT INTO vivero_catalogo (clase, nombre)
		VALUES ($1, $2)
		ON CONFLICT (clase, nombre) DO NOTHING`, clase, nombre).Error
}
