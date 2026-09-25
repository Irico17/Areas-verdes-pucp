package etl

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ReporteLote son los conteos y los rechazos de una corrida.
type ReporteLote struct {
	LoteID     int64             `json:"lote_id"`
	Origen     map[string]string `json:"origen"`
	Cargados   map[string]int    `json:"cargados"`
	Rechazados []Rechazo         `json:"rechazados"`
	Avisos     []string          `json:"avisos,omitempty"`
}

// CargarLote hace upsert por origen_ref y registra cada fila en cambios.
func CargarLote(db *gorm.DB, fuentes fuentesLote) (ReporteLote, error) {
	leido, err := interpretar(fuentes)
	if err != nil {
		return ReporteLote{}, err
	}
	rep := ReporteLote{
		Origen:     fuentes.Origen,
		Cargados:   map[string]int{},
		Rechazados: leido.Rechazos,
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		var loteID int64
		if err := tx.Raw(`SELECT COALESCE(MAX(lote_id), 0) + 1 FROM cambios`).Scan(&loteID).Error; err != nil {
			return err
		}
		rep.LoteID = loteID
		if err := cargarZonas(tx, leido.Zonas, loteID, &rep); err != nil {
			return err
		}
		if err := cargarCuadrillasYPoligonos(tx, leido.Poligonos, loteID, &rep); err != nil {
			return err
		}
		if err := cargarLugares(tx, leido.Lugares, loteID, &rep); err != nil {
			return err
		}
		if err := cargarEjemplares(tx, leido.Ejemplares, loteID, &rep); err != nil {
			return err
		}
		if err := cargarMedidas(tx, leido.Medidas, leido.Ejemplares, loteID, &rep); err != nil {
			return err
		}
		if err := cargarCafetos(tx, leido.Cafetos, loteID, &rep); err != nil {
			return err
		}
		if err := cargarTipos(tx, leido.Tipos, loteID, &rep); err != nil {
			return err
		}
		if err := cargarAtencion(tx, fuentes, loteID, &rep); err != nil {
			return err
		}
		return nil
	})
	return rep, err
}

func interpretar(f fuentesLote) (loteLeido, error) {
	var leido loteLeido
	zonas, rz, err := leerZonasSupervision(f.Zonas)
	if err != nil {
		return leido, err
	}
	leido.Zonas = zonas
	leido.Rechazos = append(leido.Rechazos, rz...)
	pols, rp, err := leerPoligonos(f.Jefes)
	if err != nil {
		return leido, err
	}
	leido.Poligonos = pols
	leido.Rechazos = append(leido.Rechazos, rp...)
	lug, rl, err := leerLugares(f.Lugares)
	if err != nil {
		return leido, err
	}
	leido.Lugares = lug
	leido.Rechazos = append(leido.Rechazos, rl...)
	ej, re, err := leerGvizFlora(f.Gviz)
	if err != nil {
		return leido, err
	}
	leido.Ejemplares = ej
	leido.Rechazos = append(leido.Rechazos, re...)
	med, rm, err := leerMedidasPalmera(f.Palmeras)
	if err != nil {
		return leido, err
	}
	leido.Medidas = med
	leido.Rechazos = append(leido.Rechazos, rm...)
	caf, _, rc, err := leerCafetos(f.Cafetos)
	if err != nil {
		return leido, err
	}
	leido.Cafetos = caf
	leido.Rechazos = append(leido.Rechazos, rc...)
	tipos, rt, err := leerActividades(f.Tipos)
	if err != nil {
		return leido, err
	}
	leido.Tipos = tipos
	leido.Rechazos = append(leido.Rechazos, rt...)
	return leido, nil
}

func cargarZonas(tx *gorm.DB, zonas []zonaSup, loteID int64, rep *ReporteLote) error {
	for _, z := range zonas {
		var antes *string
		var id int64
		err := tx.Raw(`
			INSERT INTO zonas_supervision (codigo, nombre, area_m2, geom, origen_ref)
			VALUES ($1, $2, $3, ST_Multi(ST_SetSRID(ST_GeomFromGeoJSON($4), 4326)), $5)
			ON CONFLICT (origen_ref) DO UPDATE SET
			  codigo = EXCLUDED.codigo,
			  nombre = EXCLUDED.nombre,
			  area_m2 = EXCLUDED.area_m2,
			  geom = EXCLUDED.geom,
			  updated_at = now()
			RETURNING id`,
			z.Codigo, z.Nombre, z.AreaM2, string(z.Geom), z.Origen).Scan(&id).Error
		if err != nil {
			return fmt.Errorf("zona %s: %w", z.Codigo, err)
		}
		var areaCalc float64
		if err := tx.Raw(`SELECT ST_Area(geom::geography) FROM zonas_supervision WHERE id = $1`, id).Scan(&areaCalc).Error; err != nil {
			return err
		}
		if z.AreaM2 != nil && *z.AreaM2 > 0 {
			diff := (areaCalc - *z.AreaM2) / *z.AreaM2
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.05 {
				rep.Avisos = append(rep.Avisos, fmt.Sprintf("%s: el área de la fuente difiere más del 5 %% de ST_Area", z.Codigo))
			}
		}
		despues, _ := json.Marshal(map[string]any{"codigo": z.Codigo, "nombre": z.Nombre, "origen_ref": z.Origen, "area_m2": z.AreaM2})
		if err := auditar(tx, "zonas_supervision", fmt.Sprint(id), antes, despues, loteID); err != nil {
			return err
		}
	}
	rep.Cargados["zonas_supervision"] = len(zonas)
	return nil
}

func cargarCuadrillasYPoligonos(tx *gorm.DB, pols []poligonoCarga, loteID int64, rep *ReporteLote) error {
	valores := make([]string, 0, len(pols))
	for _, p := range pols {
		if p.JefesRaw != "" {
			valores = append(valores, p.JefesRaw)
		}
	}
	tabla := asignarCuadrillas(valores)
	nCuad := 0
	for _, c := range tabla {
		res := tx.Exec(`
			INSERT INTO cuadrillas (id, nombre_ficticio, turno, activo)
			VALUES ($1, $2, 'manana', TRUE)
			ON CONFLICT (id) DO UPDATE SET nombre_ficticio = EXCLUDED.nombre_ficticio`,
			c.ID, c.Nombre)
		if res.Error != nil {
			return res.Error
		}
		nCuad++
		despues, _ := json.Marshal(map[string]any{"id": c.ID, "nombre_ficticio": c.Nombre})
		if err := auditar(tx, "cuadrillas", c.ID, nil, despues, loteID); err != nil {
			return err
		}
	}
	rep.Cargados["cuadrillas"] = nCuad
	for _, p := range pols {
		cuad, ok := cuadrillaDe(tabla, p.JefesRaw)
		var cuadID *string
		if ok {
			cuadID = &cuad.ID
		}
		var id int64
		err := tx.Raw(`
			INSERT INTO poligonos_cuadrilla (
			  feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act, referencia,
			  perimetro_m, area_m2, geom, cuadrilla_id, origen_ref
			) VALUES (
			  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			  ST_Multi(ST_SetSRID(ST_GeomFromGeoJSON($11), 4326)), $12, $13
			)
			ON CONFLICT (origen_ref) DO UPDATE SET
			  feature_id = EXCLUDED.feature_id,
			  source_index = EXCLUDED.source_index,
			  codigo = EXCLUDED.codigo,
			  nombre = EXCLUDED.nombre,
			  uso = EXCLUDED.uso,
			  proy_riego = EXCLUDED.proy_riego,
			  riego_act = EXCLUDED.riego_act,
			  referencia = EXCLUDED.referencia,
			  perimetro_m = EXCLUDED.perimetro_m,
			  area_m2 = EXCLUDED.area_m2,
			  geom = EXCLUDED.geom,
			  cuadrilla_id = EXCLUDED.cuadrilla_id,
			  updated_at = now()
			RETURNING id`,
			p.FeatureID, p.SourceIndex, p.Codigo, p.Nombre, p.Uso, p.ProyRiego, p.RiegoAct, p.Referencia,
			p.PerimetroM, p.AreaM2, string(p.Geom), cuadID, p.Origen).Scan(&id).Error
		if err != nil {
			return fmt.Errorf("polígono %s: %w", p.FeatureID, err)
		}
		if ok {
			if err := tx.Exec(`
				INSERT INTO asignaciones_poligono (poligono_id, cuadrilla_id, vigente)
				SELECT $1, $2, TRUE
				WHERE NOT EXISTS (
				  SELECT 1 FROM asignaciones_poligono
				  WHERE poligono_id = $1 AND cuadrilla_id = $2 AND vigente
				)`, id, cuad.ID).Error; err != nil {
				return err
			}
		}
		despues, _ := json.Marshal(map[string]any{
			"feature_id": p.FeatureID, "origen_ref": p.Origen, "cuadrilla_id": cuadID,
		})
		if err := auditar(tx, "poligonos_cuadrilla", fmt.Sprint(id), nil, despues, loteID); err != nil {
			return err
		}
	}
	rep.Cargados["poligonos_cuadrilla"] = len(pols)
	if err := tx.Exec(`
		UPDATE poligonos_cuadrilla p
		SET zona_supervision_id = z.id
		FROM zonas_supervision z
		WHERE p.origen_ref LIKE 'jefe:%'
		  AND z.activo
		  AND ST_Intersects(p.geom, z.geom)
		  AND p.zona_supervision_id IS NULL`).Error; err != nil {
		return err
	}
	return nil
}

func cargarLugares(tx *gorm.DB, lugares []lugarCarga, loteID int64, rep *ReporteLote) error {
	for _, l := range lugares {
		var id int64
		err := tx.Raw(`
			INSERT INTO lugares (nombre, nombre_norm, lat, lon, origen_ref)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (origen_ref) DO UPDATE SET
			  nombre = EXCLUDED.nombre,
			  nombre_norm = EXCLUDED.nombre_norm,
			  lat = EXCLUDED.lat,
			  lon = EXCLUDED.lon,
			  updated_at = now()
			RETURNING id`,
			l.Nombre, l.Norm, l.Lat, l.Lon, l.Origen).Scan(&id).Error
		if err != nil {
			return fmt.Errorf("lugar %s: %w", l.Nombre, err)
		}
		despues, _ := json.Marshal(map[string]any{"nombre": l.Nombre, "origen_ref": l.Origen})
		if err := auditar(tx, "lugares", fmt.Sprint(id), nil, despues, loteID); err != nil {
			return err
		}
	}
	if err := tx.Exec(`
		UPDATE lugares l
		SET zona_supervision_id = z.id
		FROM zonas_supervision z
		WHERE l.origen_ref LIKE 'lugar:%'
		  AND ST_Covers(z.geom, ST_SetSRID(ST_MakePoint(l.lon, l.lat), 4326))`).Error; err != nil {
		return err
	}
	rep.Cargados["lugares"] = len(lugares)
	return nil
}

func cargarEjemplares(tx *gorm.DB, rows []ejemplarCarga, loteID int64, rep *ReporteLote) error {
	especies := 0
	sinLugar := 0
	for _, e := range rows {
		var especieID *int64
		if e.Cientifico != "" {
			id, nueva, err := upsertEspecie(tx, e.Cientifico, e.Comun, "especie:"+normalizarPersona(e.Cientifico), loteID)
			if err != nil {
				return err
			}
			especieID = &id
			if nueva {
				especies++
			}
		}
		var lugarID *int64
		if e.Lugar != "" {
			var id int64
			err := tx.Raw(`SELECT COALESCE((SELECT id FROM lugares WHERE nombre_norm = $1), 0)`, normalizarPersona(e.Lugar)).Scan(&id).Error
			if err != nil {
				return err
			}
			if id == 0 {
				sinLugar++
				nro := 0
				if e.NumeroOrigen != nil {
					nro = *e.NumeroOrigen
				}
				rep.Avisos = append(rep.Avisos, fmt.Sprintf("ejemplar gviz:%d fila %d: lugar sin match; se carga sin FK", nro, e.Fila))
			} else {
				lugarID = &id
			}
		}
		var id int64
		err := tx.Raw(`
			INSERT INTO ejemplares (
			  numero_origen, codigo, especie_id, nombre_comun, tipo_vegetacion, cantidad,
			  ubicacion_lugar_id, referencia, lat, lon, observacion_fen_2026, geom, origen_ref
			) VALUES (
			  $1, NULLIF($2, ''), $3, NULLIF($4, ''), NULLIF($5, ''), $6,
			  $7, NULLIF($8, ''), $9, $10, NULLIF($11, ''),
			  ST_SetSRID(ST_MakePoint($12, $9), 4326), $13
			)
			ON CONFLICT (origen_ref) DO UPDATE SET
			  numero_origen = EXCLUDED.numero_origen,
			  codigo = EXCLUDED.codigo,
			  especie_id = EXCLUDED.especie_id,
			  nombre_comun = EXCLUDED.nombre_comun,
			  tipo_vegetacion = EXCLUDED.tipo_vegetacion,
			  cantidad = EXCLUDED.cantidad,
			  ubicacion_lugar_id = EXCLUDED.ubicacion_lugar_id,
			  referencia = EXCLUDED.referencia,
			  lat = EXCLUDED.lat,
			  lon = EXCLUDED.lon,
			  observacion_fen_2026 = EXCLUDED.observacion_fen_2026,
			  geom = EXCLUDED.geom,
			  updated_at = now()
			RETURNING id`,
			e.NumeroOrigen, e.Codigo, especieID, e.Comun, e.Tipo, e.Cantidad,
			lugarID, e.Referencia, e.Lat, e.Lon, e.Observacion, e.Lon, e.Origen).Scan(&id).Error
		if err != nil {
			return fmt.Errorf("ejemplar %s: %w", e.Origen, err)
		}
		despues, _ := json.Marshal(map[string]any{"origen_ref": e.Origen, "numero_origen": e.NumeroOrigen, "tipo": e.Tipo})
		if err := auditar(tx, "ejemplares", fmt.Sprint(id), nil, despues, loteID); err != nil {
			return err
		}
	}
	rep.Cargados["ejemplares"] = len(rows)
	rep.Cargados["especies_nuevas"] = especies
	if sinLugar > 0 {
		rep.Avisos = append(rep.Avisos, fmt.Sprintf("%d ejemplares sin lugar en el diccionario", sinLugar))
	}
	return nil
}

func upsertEspecie(tx *gorm.DB, cientifico, comun, origen string, loteID int64) (int64, bool, error) {
	var id int64
	var inserted bool
	err := tx.Raw(`
		INSERT INTO especies (nombre_cientifico, nombre_comun)
		VALUES ($1, NULLIF($2, ''))
		ON CONFLICT (nombre_cientifico) DO UPDATE SET
		  nombre_comun = COALESCE(especies.nombre_comun, EXCLUDED.nombre_comun)
		RETURNING id, (xmax = 0) AS inserted`,
		cientifico, comun).Row().Scan(&id, &inserted)
	if err != nil {
		return 0, false, err
	}
	if inserted {
		despues, _ := json.Marshal(map[string]any{"nombre_cientifico": cientifico, "origen_ref": origen})
		if err := auditar(tx, "especies", fmt.Sprint(id), nil, despues, loteID); err != nil {
			return 0, false, err
		}
	}
	return id, inserted, nil
}

func cargarMedidas(tx *gorm.DB, medidas []medidaCarga, ejemplares []ejemplarCarga, loteID int64, rep *ReporteLote) error {
	indice := map[string]string{}
	for _, e := range ejemplares {
		indice[claveUnion(e.Lat, e.Lon, e.Cientifico, e.Comun)] = e.Origen
	}
	usados := map[int64]bool{}
	n := 0
	for _, m := range medidas {
		origenEj := indice[m.ClaveUnion]
		var ejemplarID int64
		if origenEj != "" {
			if err := tx.Raw(`SELECT COALESCE((SELECT id FROM ejemplares WHERE origen_ref = $1), 0)`, origenEj).Scan(&ejemplarID).Error; err != nil {
				return err
			}
			if usados[ejemplarID] {
				ejemplarID = 0
			}
		}
		if ejemplarID == 0 {
			var especieID *int64
			if m.Cientifico != "" {
				id, _, err := upsertEspecie(tx, m.Cientifico, m.Comun, "especie:"+normalizarPersona(m.Cientifico), loteID)
				if err != nil {
					return err
				}
				especieID = &id
			}
			err := tx.Raw(`
				INSERT INTO ejemplares (
				  especie_id, nombre_comun, tipo_vegetacion, cantidad, lat, lon, geom, origen_ref
				) VALUES (
				  $1, NULLIF($2, ''), 'Palmera', 1, $3, $4,
				  ST_SetSRID(ST_MakePoint($5, $3), 4326), $6
				)
				ON CONFLICT (origen_ref) DO UPDATE SET
				  lat = EXCLUDED.lat, lon = EXCLUDED.lon, geom = EXCLUDED.geom, updated_at = now()
				RETURNING id`,
				especieID, m.Comun, m.Lat, m.Lon, m.Lon, m.Origen).Scan(&ejemplarID).Error
			if err != nil {
				return fmt.Errorf("palmera %s: %w", m.Origen, err)
			}
		}
		if err := tx.Exec(`
			INSERT INTO medidas_palmera (ejemplar_id, utm_norte, utm_este, altura, altura_fuste, dap, radio, zunchado)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (ejemplar_id) DO UPDATE SET
			  utm_norte = EXCLUDED.utm_norte,
			  utm_este = EXCLUDED.utm_este,
			  altura = EXCLUDED.altura,
			  altura_fuste = EXCLUDED.altura_fuste,
			  dap = EXCLUDED.dap,
			  radio = EXCLUDED.radio,
			  zunchado = EXCLUDED.zunchado`,
			ejemplarID, m.UTMNorte, m.UTMEste, m.Altura, m.AlturaFuste, m.DAP, m.Radio, m.Zunchado).Error; err != nil {
			return err
		}
		usados[ejemplarID] = true
		despues, _ := json.Marshal(map[string]any{"ejemplar_id": ejemplarID, "origen_ref": m.Origen})
		if err := auditar(tx, "medidas_palmera", fmt.Sprint(ejemplarID), nil, despues, loteID); err != nil {
			return err
		}
		n++
	}
	rep.Cargados["medidas_palmera"] = n
	return nil
}

func cargarCafetos(tx *gorm.DB, rows []ejemplarCarga, loteID int64, rep *ReporteLote) error {
	nNuevos := 0
	nVinculos := 0
	for _, c := range rows {
		clave := claveUnion(c.Lat, c.Lon, c.Cientifico, c.Comun)
		var existente int64
		if err := tx.Raw(`
			SELECT COALESCE((
			  SELECT e.id FROM ejemplares e
			  LEFT JOIN especies s ON s.id = e.especie_id
			  WHERE round(e.lat::numeric, 5) = round($1::numeric, 5)
			    AND round(e.lon::numeric, 5) = round($2::numeric, 5)
			    AND (
			      lower(coalesce(s.nombre_cientifico, '')) = lower($3)
			      OR lower(coalesce(e.nombre_comun, '')) = lower($4)
			    )
			  LIMIT 1
			), 0)`, c.Lat, c.Lon, c.Cientifico, c.Comun).Scan(&existente).Error; err != nil {
			return err
		}
		if existente != 0 {
			nVinculos++
			despues, _ := json.Marshal(map[string]any{"vinculo": clave, "ejemplar_id": existente, "origen_ref": c.Origen})
			if err := auditar(tx, "ejemplares", fmt.Sprint(existente), nil, despues, loteID); err != nil {
				return err
			}
			continue
		}
		var especieID *int64
		if c.Cientifico != "" {
			id, _, err := upsertEspecie(tx, c.Cientifico, c.Comun, "especie:"+normalizarPersona(c.Cientifico), loteID)
			if err != nil {
				return err
			}
			especieID = &id
		}
		var lugarID *int64
		if c.Lugar != "" {
			var id int64
			if err := tx.Raw(`SELECT COALESCE((SELECT id FROM lugares WHERE nombre_norm = $1), 0)`, normalizarPersona(c.Lugar)).Scan(&id).Error; err != nil {
				return err
			}
			if id != 0 {
				lugarID = &id
			}
		}
		var id int64
		err := tx.Raw(`
			INSERT INTO ejemplares (
			  especie_id, nombre_comun, tipo_vegetacion, cantidad, ubicacion_lugar_id,
			  lat, lon, geom, origen_ref
			) VALUES (
			  $1, NULLIF($2, ''), 'cafeto', 1, $3, $4, $5,
			  ST_SetSRID(ST_MakePoint($6, $4), 4326), $7
			)
			ON CONFLICT (origen_ref) DO UPDATE SET
			  lat = EXCLUDED.lat, lon = EXCLUDED.lon, geom = EXCLUDED.geom,
			  tipo_vegetacion = EXCLUDED.tipo_vegetacion, updated_at = now()
			RETURNING id`,
			especieID, c.Comun, lugarID, c.Lat, c.Lon, c.Lon, c.Origen).Scan(&id).Error
		if err != nil {
			return fmt.Errorf("cafeto %s: %w", c.Origen, err)
		}
		nNuevos++
		despues, _ := json.Marshal(map[string]any{"origen_ref": c.Origen, "tipo": "cafeto"})
		if err := auditar(tx, "ejemplares", fmt.Sprint(id), nil, despues, loteID); err != nil {
			return err
		}
	}
	rep.Cargados["cafetos"] = nNuevos
	rep.Cargados["cafetos_vinculados"] = nVinculos
	return nil
}

func cargarTipos(tx *gorm.DB, tipos []tipoActividad, loteID int64, rep *ReporteLote) error {
	clases := map[string]string{}
	ordenClase := 0
	for _, t := range tipos {
		if _, ok := clases[t.Clase]; ok {
			continue
		}
		ordenClase++
		codigo := slug(t.Clase)
		clases[t.Clase] = codigo
		var id int64
		err := tx.Raw(`
			INSERT INTO catalogos (clase, codigo, nombre, orden)
			VALUES ('clase_actividad', $1, $2, $3)
			ON CONFLICT (clase, codigo) DO UPDATE SET nombre = EXCLUDED.nombre, activo = TRUE
			RETURNING id`, codigo, t.Clase, ordenClase).Scan(&id).Error
		if err != nil {
			return err
		}
		despues, _ := json.Marshal(map[string]any{"clase": t.Clase, "origen_ref": "clase:" + codigo})
		if err := auditar(tx, "catalogos", fmt.Sprint(id), nil, despues, loteID); err != nil {
			return err
		}
	}
	for i, t := range tipos {
		codigo := slug(t.Clase) + "__" + slug(t.Tipo)
		var id int64
		err := tx.Raw(`
			INSERT INTO catalogos (clase, codigo, nombre, orden)
			VALUES ('tipo_actividad_hoja', $1, $2, $3)
			ON CONFLICT (clase, codigo) DO UPDATE SET nombre = EXCLUDED.nombre, activo = TRUE, orden = EXCLUDED.orden
			RETURNING id`, codigo, t.Tipo, i+1).Scan(&id).Error
		if err != nil {
			return err
		}
		despues, _ := json.Marshal(map[string]any{
			"clase": t.Clase, "tipo": t.Tipo, "descripcion": t.Descripcion, "origen_ref": t.Origen,
		})
		if err := auditar(tx, "catalogos", fmt.Sprint(id), nil, despues, loteID); err != nil {
			return err
		}
	}
	mapeo := [][2]string{
		{"riego", "Riego"},
		{"poda", "Poda"},
		{"limpieza", "Mantenimiento de jardines"},
		{"incidencia", "Inspección y monitoreo"},
		{"inspeccion", "Inspección y monitoreo"},
	}
	for _, par := range mapeo {
		despues, _ := json.Marshal(map[string]any{"codigo_semilla": par[0], "clase_hoja": par[1]})
		if err := auditar(tx, "catalogos", "semilla:"+par[0], nil, despues, loteID); err != nil {
			return err
		}
	}
	rep.Cargados["clases_actividad"] = len(clases)
	rep.Cargados["tipos_actividad"] = len(tipos)
	return nil
}

func auditar(tx *gorm.DB, entidad, entidadID string, antes *string, despues []byte, loteID int64) error {
	return tx.Exec(`
		INSERT INTO cambios (entidad, entidad_id, accion, antes, despues, lote_id)
		VALUES ($1, $2, 'importacion', $3::jsonb, $4::jsonb, $5)`,
		entidad, entidadID, antes, string(despues), loteID).Error
}

// TextoCargado reúne columnas de texto del lote para la prueba de anonimato.
func TextoCargado(db *gorm.DB, loteID int64) (string, error) {
	var partes []string
	consultas := []string{
		`SELECT coalesce(string_agg(nombre_ficticio, E'\n'), '') FROM cuadrillas WHERE id LIKE 'cf-%'`,
		`SELECT coalesce(string_agg(coalesce(nombre,'') || ' ' || coalesce(referencia,'') || ' ' || coalesce(codigo,''), E'\n'), '') FROM poligonos_cuadrilla WHERE origen_ref LIKE 'jefe:%'`,
		`SELECT coalesce(string_agg(coalesce(antes::text,'') || coalesce(despues::text,''), E'\n'), '') FROM cambios WHERE lote_id = $1`,
		`SELECT coalesce(string_agg(coalesce(nombre_comun,'') || ' ' || coalesce(observacion_fen_2026,'') || ' ' || coalesce(codigo,''), E'\n'), '') FROM ejemplares WHERE origen_ref IS NOT NULL`,
	}
	for i, q := range consultas {
		var s string
		var err error
		if strings.Contains(q, "$1") {
			err = db.Raw(q, loteID).Scan(&s).Error
		} else {
			err = db.Raw(q).Scan(&s).Error
		}
		if err != nil {
			return "", fmt.Errorf("texto %d: %w", i, err)
		}
		partes = append(partes, s)
	}
	return strings.Join(partes, "\n"), nil
}
