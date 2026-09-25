package etl

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// Conteos de tachos publicados en docs/MAPA-DATOS-Y-EDICION.md §4.11.
var ConteosTachosEsperados = map[string]int{
	"no_aprovechables":     294,
	"papel_carton":         142,
	"plastico":             260,
	"vidrio":               242,
	"pilas":                56,
	"peligrosos":           0,
	"raee":                 12,
	"metales":              2,
	"aniquem":              38,
	"intermedios_plastico": 26,
	"intermedios_metal":    4,
}

// ReporteCapas resume la carga 2B. ColumnasOmitidas nombra datos personales no guardados.
type ReporteCapas struct {
	Cargados         map[string]int `json:"cargados"`
	Rechazados       []Rechazo      `json:"rechazados"`
	ColumnasOmitidas []string       `json:"columnas_omitidas"`
	Avisos           []string       `json:"avisos,omitempty"`
	SumasTachos      map[string]int `json:"sumas_tachos,omitempty"`
}

type tachoCarga struct {
	Codigo          string
	Lat             float64
	Lon             float64
	Nota            string
	Lugar           string
	Espacios        string
	Accion          string
	TachoActual     string
	TachoNuevo      string
	Recomendaciones string
	Conteos         map[string]int
	Origen          string
	SinPunto        bool
}

type bebederoCarga struct {
	Codigo  string
	Subtipo string
	Estado  string
	Sede    string
	Lat     float64
	Lon     float64
	Foto    string
	Geom    json.RawMessage
	Origen  string
}

type puntoCarga struct {
	Titulo string
	Lat    float64
	Lon    float64
	URL    string
	Origen string
}

type reservaCarga struct {
	Origen    string
	JardinCod string
	JardinNom string
	Fecha     string
	Inicio    string
	Fin       string
	Estado    string
	Evento    string
	Unidad    string
}

type capaPunto struct {
	FeatureID string
	Nombre    string
	Codigo    string
	Nota      string
	Geom      json.RawMessage
}

// LeerTachos lee el CSV local. La fila sin coordenada queda en el reporte.
func LeerTachos(body []byte) ([]tachoCarga, []Rechazo, map[string]int, error) {
	records, err := csvRecords(body)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(records) < 2 {
		return nil, nil, nil, nil
	}
	header := indexHeader(records[0])
	cols := map[string]string{
		"no_aprovechables":     "no aprovechables",
		"papel_carton":         "papel y cartón",
		"plastico":             "plástico",
		"vidrio":               "vidrio",
		"pilas":                "pilas",
		"peligrosos":           "peligrosos",
		"raee":                 "raee",
		"metales":              "metales",
		"aniquem":              "aniquem",
		"intermedios_plastico": "intermedios plastico",
		"intermedios_metal":    "intermedios metal",
	}
	sumas := map[string]int{}
	var rows []tachoCarga
	var rech []Rechazo
	for i, rec := range records[1:] {
		fila := i + 2
		codigo := cell(rec, header, "id")
		conteos := map[string]int{}
		total := 0
		for campo, col := range cols {
			n := enteroCelda(cell(rec, header, col))
			if n < 0 {
				rech = append(rech, Rechazo{Fuente: "tachos", Fila: fila, Campo: campo, Motivo: "conteo negativo"})
				n = 0
			}
			conteos[campo] = n
			sumas[campo] += n
			total += n
		}
		if codigo == "" && total == 0 {
			continue
		}
		lat, lon, ok := latLonFrom(rec)
		sinPunto := !ok
		if codigo == "" {
			codigo = fmt.Sprintf("PT_sin_id_%d", fila)
			sinPunto = true
			rech = append(rech, Rechazo{Fuente: "tachos", Fila: fila, Campo: "codigo", Motivo: "sin identificador ni coordenada; los conteos se conservan para el total"})
		} else if !strings.HasPrefix(codigo, "PT") {
			rech = append(rech, Rechazo{Fuente: "tachos", Fila: fila, Campo: "codigo", Motivo: "código distinto de PT"})
			continue
		} else if sinPunto {
			rech = append(rech, Rechazo{Fuente: "tachos", Fila: fila, Campo: "latitud", Motivo: "sin coordenada en el campus"})
		}
		rows = append(rows, tachoCarga{
			Codigo:          codigo,
			Lat:             lat,
			Lon:             lon,
			Nota:            sinHTML(cell(rec, header, "note")),
			Lugar:           sinHTML(cell(rec, header, "lugar")),
			Espacios:        sinHTML(cell(rec, header, "espacios")),
			Accion:          sinHTML(cell(rec, header, "accion")),
			TachoActual:     sinHTML(cell(rec, header, "tacho actual")),
			TachoNuevo:      sinHTML(cell(rec, header, "tacho nuevo")),
			Recomendaciones: sinHTML(cell(rec, header, "recomendaciones")),
			Conteos:         conteos,
			Origen:          "tacho:" + codigo,
			SinPunto:        sinPunto,
		})
	}
	return rows, rech, sumas, nil
}

// LeerBebederos toma estado y sede de las columnas KML (_4 y _5), no del nombre del archivo.
func LeerBebederos(rawDir string) ([]bebederoCarga, []Rechazo, error) {
	fotos := fotoIndex(rawDir)
	var rows []bebederoCarga
	var rech []Rechazo
	seen := map[string]int{}
	for _, spec := range bebederoFiles {
		path := filepath.Join(rawDir, spec.file)
		body, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, err
		}
		fc, err := readFC(path)
		if err != nil {
			return nil, nil, err
		}
		_ = body
		for i, f := range fc.Features {
			codigo := firstString(f.Properties, "Name", "name")
			estado := colSufijo(f.Properties, "_4")
			sede := colSufijo(f.Properties, "_5")
			if codigo == "" || !strings.HasPrefix(codigo, "PT") {
				rech = append(rech, Rechazo{Fuente: spec.file, Fila: i + 1, Campo: "codigo", Motivo: "sin código PT"})
				continue
			}
			if estado == "" || sede == "" {
				rech = append(rech, Rechazo{Fuente: spec.file, Fila: i + 1, Campo: "estado", Motivo: "el archivo no trae columna de estado o sede; se marca sin_dato"})
				if estado == "" {
					estado = "sin_dato"
				}
				if sede == "" {
					sede = "sin_dato"
				}
			}
			if strings.Contains(estado, "<") || strings.Contains(sede, "<") || strings.Contains(estado, "http") {
				rech = append(rech, Rechazo{Fuente: spec.file, Fila: i + 1, Campo: "estado", Motivo: "HTML no se guarda"})
				continue
			}
			lat, lon := puntoDe(f.Geometry)
			if !puntoEnCampus(lat, lon) {
				rech = append(rech, Rechazo{Fuente: spec.file, Fila: i + 1, Campo: "latitud", Motivo: "fuera del campus"})
				continue
			}
			id := uniqueID(seen, codigo)
			rows = append(rows, bebederoCarga{
				Codigo:  codigo,
				Subtipo: spec.subtipo,
				Estado:  estado,
				Sede:    sede,
				Lat:     lat,
				Lon:     lon,
				Foto:    matchFoto(fotos, codigo),
				Geom:    f.Geometry,
				Origen:  "bebedero:" + id,
			})
		}
	}
	return rows, rech, nil
}

// LeerPuntosPUCP guarda título, latitud, longitud y una URL sin placeId.
// Teléfono, placeId, website e imagen se listan como columnas omitidas.
func LeerPuntosPUCP(body []byte) ([]puntoCarga, []Rechazo, []string, error) {
	records, err := csvRecords(body)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(records) < 2 {
		return nil, nil, nil, nil
	}
	omitidas := []string{"phone", "phoneUnformatted", "placeId", "website", "image", "searchPageUrl"}
	header := indexHeader(records[0])
	var rows []puntoCarga
	var rech []Rechazo
	for i, rec := range records[1:] {
		fila := i + 2
		titulo := cell(rec, header, "title")
		lat := parseCoord(cell(rec, header, "location/lat"))
		lon := parseCoord(cell(rec, header, "location/lng"))
		if titulo == "" || !puntoEnCampus(lat, lon) {
			rech = append(rech, Rechazo{Fuente: "puntos_pucp", Fila: fila, Campo: "latitud", Motivo: "sin título o fuera del campus"})
			continue
		}
		url := urlSinContacto(cell(rec, header, "url"))
		rows = append(rows, puntoCarga{
			Titulo: titulo,
			Lat:    lat,
			Lon:    lon,
			URL:    url,
			Origen: fmt.Sprintf("punto:%d", fila),
		})
	}
	return rows, rech, omitidas, nil
}

// LeerReservasFicticias lee el mock local. No abre la hoja (HTTP 401).
func LeerReservasFicticias(body []byte) ([]reservaCarga, []Rechazo, error) {
	var doc struct {
		Reservas []struct {
			ID                string `json:"id"`
			Jardin            string `json:"jardin"`
			JardinCodigo      string `json:"jardin_codigo"`
			Fecha             string `json:"fecha"`
			Hora              string `json:"hora"`
			Evento            string `json:"evento"`
			UnidadResponsable string `json:"unidad_responsable"`
			Estado            string `json:"estado"`
			Fake              bool   `json:"_fake"`
		} `json:"reservas"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, nil, err
	}
	var rows []reservaCarga
	var rech []Rechazo
	for i, r := range doc.Reservas {
		if !r.Fake || !strings.HasPrefix(r.ID, "FAKE-") {
			rech = append(rech, Rechazo{Fuente: "reservas", Fila: i + 1, Campo: "origen", Motivo: "la hoja sigue en 401: solo se acepta origen ficticio"})
			continue
		}
		inicio, fin, ok := partirHora(r.Hora)
		if !ok {
			rech = append(rech, Rechazo{Fuente: "reservas", Fila: i + 1, Campo: "hora", Motivo: "hora fin no es posterior al inicio"})
			continue
		}
		estado := estadoReserva(r.Estado)
		if estado == "" {
			rech = append(rech, Rechazo{Fuente: "reservas", Fila: i + 1, Campo: "estado", Motivo: "estado fuera de reservado, realizado, cancelado"})
			continue
		}
		rows = append(rows, reservaCarga{
			Origen:    r.ID,
			JardinCod: r.JardinCodigo,
			JardinNom: r.Jardin,
			Fecha:     r.Fecha,
			Inicio:    inicio,
			Fin:       fin,
			Estado:    estado,
			Evento:    r.Evento,
			Unidad:    r.UnidadResponsable,
		})
	}
	return rows, rech, nil
}

// CargarAreasVerdes hace upsert del catastro por feature_id, sin TRUNCATE.
// El comando cmd/etl (Load) sigue truncando; este es el camino de etl-lote.
// Si las 521 áreas ya están por ese otro comando, el upsert las reconcilia y el total no crece.
func CargarAreasVerdes(db *gorm.DB, rawDir string) (int, error) {
	body, err := os.ReadFile(filepath.Join(rawDir, "areas_verdes.geojson"))
	if err != nil {
		return 0, err
	}
	rows, err := NormalizeAreas(body)
	if err != nil {
		return 0, err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, r := range rows {
			if err := tx.Exec(`
				INSERT INTO areas_verdes (
				  feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act, referencia,
				  perimetro_m, area_m2, geom, origen_ref, activo
				) VALUES (
				  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, catastro_geom_4326($11), $1, TRUE
				)
				ON CONFLICT (feature_id) DO UPDATE SET
				  codigo = EXCLUDED.codigo,
				  nombre = EXCLUDED.nombre,
				  uso = EXCLUDED.uso,
				  proy_riego = EXCLUDED.proy_riego,
				  riego_act = EXCLUDED.riego_act,
				  referencia = EXCLUDED.referencia,
				  perimetro_m = EXCLUDED.perimetro_m,
				  area_m2 = EXCLUDED.area_m2,
				  geom = EXCLUDED.geom,
				  origen_ref = EXCLUDED.origen_ref,
				  updated_at = now()`,
				r.FeatureID, r.SourceIndex, r.Codigo, r.Nombre, r.Uso, r.ProyRiego, r.RiegoAct, r.Referencia,
				r.PerimetroM, r.AreaM2, geomArg(r),
			).Error; err != nil {
				return fmt.Errorf("área %s: %w", r.FeatureID, err)
			}
		}
		return nil
	})
	return len(rows), err
}

// CargarFrente2B inserta tachos, bebederos, puntos, capas 4.13 y reservas ficticias.
func CargarFrente2B(db *gorm.DB, rawDir string) (ReporteCapas, error) {
	rep := ReporteCapas{Cargados: map[string]int{}, ColumnasOmitidas: []string{}}
	tachosBody, err := os.ReadFile(filepath.Join(rawDir, "sheets", "tachos.csv"))
	if err != nil {
		return rep, err
	}
	tachos, rt, sumas, err := LeerTachos(tachosBody)
	if err != nil {
		return rep, err
	}
	rep.Rechazados = append(rep.Rechazados, rt...)
	rep.SumasTachos = sumas
	bebederos, rb, err := LeerBebederos(rawDir)
	if err != nil {
		return rep, err
	}
	rep.Rechazados = append(rep.Rechazados, rb...)
	puntosBody, err := os.ReadFile(filepath.Join(rawDir, "sheets", "puntos_pucp.csv"))
	if err != nil {
		return rep, err
	}
	puntos, rp, omitidas, err := LeerPuntosPUCP(puntosBody)
	if err != nil {
		return rep, err
	}
	rep.Rechazados = append(rep.Rechazados, rp...)
	rep.ColumnasOmitidas = append(rep.ColumnasOmitidas, omitidas...)
	rep.ColumnasOmitidas = append(rep.ColumnasOmitidas, "tachos.Foto (URL de Drive)")
	reservasBody, err := os.ReadFile(filepath.Join(rawDir, "..", "mocks", "reservas_agenda.mock.json"))
	if err != nil {
		return rep, err
	}
	reservas, rr, err := LeerReservasFicticias(reservasBody)
	if err != nil {
		return rep, err
	}
	rep.Rechazados = append(rep.Rechazados, rr...)

	err = db.Transaction(func(tx *gorm.DB) error {
		var loteID int64
		if err := tx.Raw(`SELECT COALESCE(MAX(lote_id), 0) + 1 FROM cambios`).Scan(&loteID).Error; err != nil {
			return err
		}
		if err := guardarTachos(tx, tachos, loteID); err != nil {
			return err
		}
		rep.Cargados["tachos"] = len(tachos)
		if err := guardarBebederos(tx, bebederos, loteID); err != nil {
			return err
		}
		rep.Cargados["bebederos"] = len(bebederos)
		if err := guardarPuntos(tx, puntos, loteID); err != nil {
			return err
		}
		rep.Cargados["puntos_pucp"] = len(puntos)
		n, err := guardarCapasEditables(tx, rawDir, loteID)
		if err != nil {
			return err
		}
		for k, v := range n {
			rep.Cargados[k] = v
		}
		if err := guardarReservas(tx, reservas, loteID, &rep); err != nil {
			return err
		}
		return nil
	})
	return rep, err
}

func guardarTachos(tx *gorm.DB, rows []tachoCarga, loteID int64) error {
	for _, t := range rows {
		var id int64
		var lat, lon any
		if !t.SinPunto {
			lat, lon = t.Lat, t.Lon
		}
		err := tx.Raw(`
			INSERT INTO tachos (
			  codigo, lat, lon, nota, lugar, espacios, accion, tacho_actual, tacho_nuevo, recomendaciones,
			  no_aprovechables, papel_carton, plastico, vidrio, pilas, peligrosos, raee, metales, aniquem,
			  intermedios_plastico, intermedios_metal, geom, origen_ref
			) VALUES (
			  $1,$2,$3, NULLIF($4,''), NULLIF($5,''), NULLIF($6,''), NULLIF($7,''), NULLIF($8,''), NULLIF($9,''), NULLIF($10,''),
			  $11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,
			  CASE WHEN $23::bool THEN NULL ELSE ST_SetSRID(ST_MakePoint($3,$2),4326) END, $22
			)
			ON CONFLICT (origen_ref) DO UPDATE SET
			  codigo = EXCLUDED.codigo, lat = EXCLUDED.lat, lon = EXCLUDED.lon,
			  nota = EXCLUDED.nota, lugar = EXCLUDED.lugar, espacios = EXCLUDED.espacios,
			  accion = EXCLUDED.accion, tacho_actual = EXCLUDED.tacho_actual, tacho_nuevo = EXCLUDED.tacho_nuevo,
			  recomendaciones = EXCLUDED.recomendaciones,
			  no_aprovechables = EXCLUDED.no_aprovechables, papel_carton = EXCLUDED.papel_carton,
			  plastico = EXCLUDED.plastico, vidrio = EXCLUDED.vidrio, pilas = EXCLUDED.pilas,
			  peligrosos = EXCLUDED.peligrosos, raee = EXCLUDED.raee, metales = EXCLUDED.metales,
			  aniquem = EXCLUDED.aniquem, intermedios_plastico = EXCLUDED.intermedios_plastico,
			  intermedios_metal = EXCLUDED.intermedios_metal, geom = EXCLUDED.geom, updated_at = now()
			RETURNING id`,
			t.Codigo, lat, lon, t.Nota, t.Lugar, t.Espacios, t.Accion, t.TachoActual, t.TachoNuevo, t.Recomendaciones,
			t.Conteos["no_aprovechables"], t.Conteos["papel_carton"], t.Conteos["plastico"], t.Conteos["vidrio"],
			t.Conteos["pilas"], t.Conteos["peligrosos"], t.Conteos["raee"], t.Conteos["metales"], t.Conteos["aniquem"],
			t.Conteos["intermedios_plastico"], t.Conteos["intermedios_metal"], t.Origen, t.SinPunto,
		).Scan(&id).Error
		if err != nil {
			return fmt.Errorf("tacho %s: %w", t.Codigo, err)
		}
		despues, _ := json.Marshal(map[string]any{"codigo": t.Codigo, "origen_ref": t.Origen})
		if err := auditar(tx, "tachos", fmt.Sprint(id), nil, despues, loteID); err != nil {
			return err
		}
	}
	if err := tx.Exec(`
		UPDATE tachos t SET zona_supervision_id = z.id
		FROM zonas_supervision z
		WHERE t.geom IS NOT NULL AND z.geom IS NOT NULL AND ST_Covers(z.geom, t.geom)
	`).Error; err != nil {
		return err
	}
	return nil
}

func guardarBebederos(tx *gorm.DB, rows []bebederoCarga, loteID int64) error {
	for _, b := range rows {
		var id int64
		err := tx.Raw(`
			INSERT INTO bebederos (codigo, subtipo, estado, sede, lat, lon, foto, geom, origen_ref)
			VALUES ($1,$2,$3, NULLIF($4,''), $5, $6, NULLIF($7,''), inventario_geom_4326($8), $9)
			ON CONFLICT (origen_ref) DO UPDATE SET
			  codigo = EXCLUDED.codigo, subtipo = EXCLUDED.subtipo, estado = EXCLUDED.estado,
			  sede = EXCLUDED.sede, lat = EXCLUDED.lat, lon = EXCLUDED.lon, foto = EXCLUDED.foto,
			  geom = EXCLUDED.geom, updated_at = now()
			RETURNING id`,
			b.Codigo, b.Subtipo, b.Estado, b.Sede, b.Lat, b.Lon, b.Foto, string(b.Geom), b.Origen,
		).Scan(&id).Error
		if err != nil {
			return fmt.Errorf("bebedero %s: %w", b.Codigo, err)
		}
		despues, _ := json.Marshal(map[string]any{"codigo": b.Codigo, "estado": b.Estado, "sede": b.Sede, "subtipo": b.Subtipo})
		if err := auditar(tx, "bebederos", fmt.Sprint(id), nil, despues, loteID); err != nil {
			return err
		}
	}
	return nil
}

func guardarPuntos(tx *gorm.DB, rows []puntoCarga, loteID int64) error {
	for _, p := range rows {
		var id int64
		err := tx.Raw(`
			INSERT INTO puntos_pucp (titulo, lat, lon, url, geom, origen_ref)
			VALUES ($1,$2,$3, NULLIF($4,''), ST_SetSRID(ST_MakePoint($3,$2),4326), $5)
			ON CONFLICT (origen_ref) DO UPDATE SET
			  titulo = EXCLUDED.titulo, lat = EXCLUDED.lat, lon = EXCLUDED.lon, url = EXCLUDED.url,
			  geom = EXCLUDED.geom, updated_at = now()
			RETURNING id`,
			p.Titulo, p.Lat, p.Lon, p.URL, p.Origen,
		).Scan(&id).Error
		if err != nil {
			return fmt.Errorf("punto %s: %w", p.Origen, err)
		}
		despues, _ := json.Marshal(map[string]any{"titulo": p.Titulo, "origen_ref": p.Origen})
		if strings.Contains(string(despues), "placeId") || strings.Contains(strings.ToLower(p.URL), "phone") {
			return fmt.Errorf("punto %s todavía arrastra contacto", p.Origen)
		}
		if err := auditar(tx, "puntos_pucp", fmt.Sprint(id), nil, despues, loteID); err != nil {
			return err
		}
	}
	return nil
}

func guardarCapasEditables(tx *gorm.DB, rawDir string, loteID int64) (map[string]int, error) {
	out := map[string]int{}
	puntos := []struct {
		file, capa, prefix, tabla, nameKey string
	}{
		{"fauna.geojson", "fauna", "FA", "fauna", "Animal"},
		{"puertas_entradas.geojson", "puertas", "PU", "puertas", "Name"},
		{"playas_de_estacionamiento.geojson", "playas", "PL", "playas_estacionamiento", "Name"},
		{"area_vereda_peligro.geojson", "vereda", "VD", "veredas_riesgo", "Name"},
	}
	for _, spec := range puntos {
		rows, err := leerCapaPuntos(rawDir, spec.file, spec.prefix, spec.nameKey)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			if err := upsertCapaPunto(tx, spec.tabla, row, loteID); err != nil {
				return nil, err
			}
		}
		out[spec.tabla] = len(rows)
	}
	jrBody, err := os.ReadFile(filepath.Join(rawDir, "jardines_reserva.geojson"))
	if err != nil {
		return nil, err
	}
	jr, err := NormalizeCapa(jrBody, "jardines_reserva", "JR")
	if err != nil {
		return nil, err
	}
	for _, r := range jr {
		if err := tx.Exec(`
			INSERT INTO jardines_reserva (
			  feature_id, codigo, nombre, uso, proy_riego, riego_act, referencia, pertenecen,
			  perimetro_m, area_m2, geom, origen_ref
			) VALUES (
			  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10, catastro_geom_4326($11), $1
			)
			ON CONFLICT (feature_id) DO UPDATE SET
			  codigo = EXCLUDED.codigo, nombre = EXCLUDED.nombre, uso = EXCLUDED.uso,
			  proy_riego = EXCLUDED.proy_riego, riego_act = EXCLUDED.riego_act,
			  referencia = EXCLUDED.referencia, pertenecen = EXCLUDED.pertenecen,
			  perimetro_m = EXCLUDED.perimetro_m, area_m2 = EXCLUDED.area_m2,
			  geom = EXCLUDED.geom, origen_ref = EXCLUDED.origen_ref, updated_at = now()`,
			r.FeatureID, r.Codigo, r.Nombre, r.Uso, r.ProyRiego, r.RiegoAct, r.Referencia, r.Pertenecen,
			r.PerimetroM, r.AreaM2, geomArg(r),
		).Error; err != nil {
			return nil, fmt.Errorf("jardín %s: %w", r.FeatureID, err)
		}
	}
	out["jardines_reserva"] = len(jr)
	xeBody, err := os.ReadFile(filepath.Join(rawDir, "xerofitica.geojson"))
	if err != nil {
		return nil, err
	}
	xe, err := NormalizeCapa(xeBody, "xerofitica", "XE")
	if err != nil {
		return nil, err
	}
	for _, r := range xe {
		if err := tx.Exec(`
			INSERT INTO xerofiticas (feature_id, clase, riego, area_m2, perimetro_m, geom, origen_ref)
			VALUES ($1,$2,$3,$4,$5, catastro_geom_4326($6), $1)
			ON CONFLICT (feature_id) DO UPDATE SET
			  clase = EXCLUDED.clase, riego = EXCLUDED.riego, area_m2 = EXCLUDED.area_m2,
			  perimetro_m = EXCLUDED.perimetro_m, geom = EXCLUDED.geom, origen_ref = EXCLUDED.origen_ref,
			  updated_at = now()`,
			r.FeatureID, r.Clase, r.RiegoAct, r.AreaM2, r.PerimetroM, geomArg(r),
		).Error; err != nil {
			return nil, fmt.Errorf("xerofítica %s: %w", r.FeatureID, err)
		}
	}
	out["xerofiticas"] = len(xe)
	return out, nil
}

func guardarReservas(tx *gorm.DB, rows []reservaCarga, loteID int64, rep *ReporteCapas) error {
	for _, r := range rows {
		var jardinID *int64
		var idEncontrado int64
		err := tx.Raw(`
			SELECT id FROM jardines_reserva
			WHERE activo AND codigo = $1 AND nombre = $2
			ORDER BY id LIMIT 1`, r.JardinCod, r.JardinNom).Scan(&idEncontrado).Error
		if err != nil {
			return err
		}
		if idEncontrado > 0 {
			jardinID = &idEncontrado
		} else {
			rep.Avisos = append(rep.Avisos, fmt.Sprintf("reserva %s sin jardín %s", r.Origen, r.JardinCod))
		}
		var id int64
		err = tx.Raw(`
			INSERT INTO reservas_jardin (jardin_id, fecha, hora_inicio, hora_fin, estado, evento, unidad, origen, origen_ref)
			VALUES ($1,$2,$3,$4,$5,$6, NULLIF($7,''), 'ficticio', $8)
			ON CONFLICT (origen_ref) DO UPDATE SET
			  jardin_id = EXCLUDED.jardin_id, fecha = EXCLUDED.fecha,
			  hora_inicio = EXCLUDED.hora_inicio, hora_fin = EXCLUDED.hora_fin,
			  estado = EXCLUDED.estado, evento = EXCLUDED.evento, unidad = EXCLUDED.unidad,
			  origen = 'ficticio', updated_at = now()
			RETURNING id`,
			jardinID, r.Fecha, r.Inicio, r.Fin, r.Estado, r.Evento, r.Unidad, r.Origen,
		).Scan(&id).Error
		if err != nil {
			return fmt.Errorf("reserva %s: %w", r.Origen, err)
		}
		despues, _ := json.Marshal(map[string]any{"origen": "ficticio", "origen_ref": r.Origen, "evento": r.Evento})
		if err := auditar(tx, "reservas_jardin", fmt.Sprint(id), nil, despues, loteID); err != nil {
			return err
		}
	}
	rep.Cargados["reservas_jardin"] = len(rows)
	return nil
}

func leerCapaPuntos(rawDir, file, prefix, nameKey string) ([]capaPunto, error) {
	path := filepath.Join(rawDir, file)
	fc, err := readFC(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	rows := make([]capaPunto, 0, len(fc.Features))
	for i, f := range fc.Features {
		nombre := firstString(f.Properties, nameKey, "Name", "nombre", "Animal")
		codigo := firstString(f.Properties, "Name", "código", "codigo")
		if nameKey == "Animal" {
			codigo = ""
		}
		nota := firstString(f.Properties, "nota", "Note", "detalle")
		rows = append(rows, capaPunto{
			FeatureID: fmt.Sprintf("%s-%04d", prefix, i+1),
			Nombre:    nombre,
			Codigo:    codigo,
			Nota:      nota,
			Geom:      f.Geometry,
		})
	}
	return rows, nil
}

func upsertCapaPunto(tx *gorm.DB, tabla string, row capaPunto, loteID int64) error {
	geom := string(row.Geom)
	var q string
	var args []any
	switch tabla {
	case "fauna":
		q = `INSERT INTO fauna (feature_id, nombre, geom, origen_ref) VALUES ($1, NULLIF($2,''), inventario_geom_4326($3), $1)
			ON CONFLICT (feature_id) DO UPDATE SET nombre = EXCLUDED.nombre, geom = EXCLUDED.geom, origen_ref = EXCLUDED.origen_ref, updated_at = now()`
		args = []any{row.FeatureID, row.Nombre, geom}
	case "puertas":
		q = `INSERT INTO puertas (feature_id, codigo, nombre, geom, origen_ref) VALUES ($1, NULLIF($2,''), NULL, inventario_geom_4326($3), $1)
			ON CONFLICT (feature_id) DO UPDATE SET codigo = EXCLUDED.codigo, geom = EXCLUDED.geom, origen_ref = EXCLUDED.origen_ref, updated_at = now()`
		args = []any{row.FeatureID, row.Codigo, geom}
	case "playas_estacionamiento":
		q = `INSERT INTO playas_estacionamiento (feature_id, codigo, geom, origen_ref) VALUES ($1, NULLIF($2,''), inventario_geom_4326($3), $1)
			ON CONFLICT (feature_id) DO UPDATE SET codigo = EXCLUDED.codigo, geom = EXCLUDED.geom, origen_ref = EXCLUDED.origen_ref, updated_at = now()`
		args = []any{row.FeatureID, row.Codigo, geom}
	case "veredas_riesgo":
		q = `INSERT INTO veredas_riesgo (feature_id, nota, geom, origen_ref) VALUES ($1, NULLIF($2,''), inventario_geom_4326($3), $1)
			ON CONFLICT (feature_id) DO UPDATE SET nota = EXCLUDED.nota, geom = EXCLUDED.geom, origen_ref = EXCLUDED.origen_ref, updated_at = now()`
		args = []any{row.FeatureID, row.Nota, geom}
	default:
		return fmt.Errorf("tabla %s", tabla)
	}
	if err := tx.Exec(q, args...).Error; err != nil {
		return fmt.Errorf("%s %s: %w", tabla, row.FeatureID, err)
	}
	despues, _ := json.Marshal(map[string]any{"feature_id": row.FeatureID})
	return auditar(tx, tabla, row.FeatureID, nil, despues, loteID)
}

func csvRecords(body []byte) ([][]string, error) {
	r := csv.NewReader(strings.NewReader(string(body)))
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	var rows [][]string
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, rec)
	}
	return rows, nil
}

func enteroCelda(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "null") {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return n
}

func sinHTML(s string) string {
	if strings.Contains(s, "<") || strings.Contains(s, "http") || strings.Contains(s, "drive.google") {
		return ""
	}
	return s
}

func colSufijo(props map[string]any, sufijo string) string {
	for key, value := range props {
		if !strings.HasSuffix(key, sufijo) {
			continue
		}
		s, ok := value.(string)
		if !ok {
			continue
		}
		s = strings.TrimSpace(s)
		if s == "" || strings.Contains(s, "<") || strings.Contains(s, "http") {
			continue
		}
		if _, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64); err == nil && strings.Contains(s, ".") {
			continue
		}
		return s
	}
	return ""
}

func puntoDe(raw json.RawMessage) (float64, float64) {
	var g struct {
		Coordinates []float64 `json:"coordinates"`
	}
	if err := json.Unmarshal(raw, &g); err != nil || len(g.Coordinates) < 2 {
		return 0, 0
	}
	return g.Coordinates[1], g.Coordinates[0]
}

func puntoEnCampus(lat, lon float64) bool {
	return lat <= -11.90 && lat >= -12.20 && lon <= -76.90 && lon >= -77.30
}

func parseCoord(raw string) float64 {
	v, ok := commaFloat(raw)
	if !ok {
		return 0
	}
	return v
}

func urlSinContacto(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.Contains(strings.ToLower(raw), "phone") {
		return ""
	}
	cortes := []string{"&query_place_id=", "?query_place_id=", "&placeId=", "?placeId=", "&place_id=", "?place_id="}
	for _, corte := range cortes {
		if i := strings.Index(raw, corte); i >= 0 {
			raw = raw[:i]
		}
	}
	if strings.Contains(raw, "placeId") || strings.Contains(raw, "place_id") {
		return ""
	}
	return raw
}

func partirHora(raw string) (string, string, bool) {
	partes := strings.Split(raw, "-")
	if len(partes) != 2 {
		return "", "", false
	}
	inicio := strings.TrimSpace(partes[0])
	fin := strings.TrimSpace(partes[1])
	if inicio == "" || fin == "" || fin <= inicio {
		return "", "", false
	}
	return inicio, fin, true
}

func estadoReserva(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "reservado", "confirmado":
		return "reservado"
	case "realizado", "realizada":
		return "realizado"
	case "cancelado", "cancelada":
		return "cancelado"
	default:
		return ""
	}
}
