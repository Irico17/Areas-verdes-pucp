package etl

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

const (
	latMin = -12.20
	latMax = -11.90
	lonMin = -77.30
	lonMax = -76.90
)

// Rechazo es una fila que no entra a la base.
type Rechazo struct {
	Fuente string `json:"fuente"`
	Fila   int    `json:"fila"`
	Campo  string `json:"campo"`
	Motivo string `json:"motivo"`
}

type zonaSup struct {
	Codigo string
	Nombre string
	AreaM2 *float64
	Geom   json.RawMessage
	Origen string
}

type poligonoCarga struct {
	FeatureID   string
	SourceIndex int
	Codigo      *string
	Nombre      *string
	Uso         *string
	ProyRiego   *string
	RiegoAct    *string
	Referencia  *string
	PerimetroM  *float64
	AreaM2      *float64
	Geom        json.RawMessage
	JefesRaw    string
	Origen      string
}

type lugarCarga struct {
	Nombre string
	Norm   string
	Lat    float64
	Lon    float64
	Origen string
	Fila   int
}

type ejemplarCarga struct {
	NumeroOrigen *int
	Codigo       string
	Cientifico   string
	Comun        string
	Tipo         string
	Cantidad     int
	Lugar        string
	Referencia   string
	Lat          float64
	Lon          float64
	Observacion  string
	Origen       string
	Fila         int
}

type medidaCarga struct {
	OrigenEjemplar string
	ClaveUnion     string
	UTMNorte       *float64
	UTMEste        *float64
	Altura         *float64
	AlturaFuste    *float64
	DAP            *float64
	Radio          *float64
	Zunchado       *bool
	Comun          string
	Cientifico     string
	Lugar          string
	Lat            float64
	Lon            float64
	Origen         string
}

type tipoActividad struct {
	Clase       string
	Tipo        string
	Descripcion string
	Origen      string
}

type loteLeido struct {
	Zonas      []zonaSup
	Poligonos  []poligonoCarga
	Lugares    []lugarCarga
	Ejemplares []ejemplarCarga
	Medidas    []medidaCarga
	Cafetos    []ejemplarCarga
	Tipos      []tipoActividad
	Rechazos   []Rechazo
}

func leerZonasSupervision(body []byte) ([]zonaSup, []Rechazo, error) {
	fc, err := parseFC(body)
	if err != nil {
		return nil, nil, fmt.Errorf("zonas de supervisión: %w", err)
	}
	var out []zonaSup
	var rechazos []Rechazo
	vistos := map[string]bool{}
	for i, ft := range fc.Features {
		fila := i + 1
		props := ft.Properties
		codigo, ok := codigoZona(propString(props, "ZONA", "Zona", "zona"))
		if !ok {
			rechazos = append(rechazos, Rechazo{Fuente: "supervisoress", Fila: fila, Campo: "ZONA", Motivo: "código fuera de Z1–Z4"})
			continue
		}
		if vistos[codigo] {
			rechazos = append(rechazos, Rechazo{Fuente: "supervisoress", Fila: fila, Campo: "ZONA", Motivo: "código repetido"})
			continue
		}
		geom, err := promoteMultiPolygon(ft.Geometry)
		if err != nil || bytes.Equal(bytes.TrimSpace(geom), []byte("null")) {
			rechazos = append(rechazos, Rechazo{Fuente: "supervisoress", Fila: fila, Campo: "geom", Motivo: "geometría inválida"})
			continue
		}
		vistos[codigo] = true
		area := propFloat(props, "Area", "Área", "area")
		out = append(out, zonaSup{
			Codigo: codigo,
			Nombre: "Zona " + strings.TrimPrefix(codigo, "Z"),
			AreaM2: area,
			Geom:   geom,
			Origen: "zona:" + codigo,
		})
	}
	return out, rechazos, nil
}

func codigoZona(raw string) (string, bool) {
	s := strings.ToUpper(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimPrefix(s, "ZONA")
	s = strings.TrimPrefix(s, "Z")
	switch s {
	case "1", "2", "3", "4":
		return "Z" + s, true
	default:
		return "", false
	}
}

func leerPoligonos(body []byte) ([]poligonoCarga, []Rechazo, error) {
	fc, err := parseFC(body)
	if err != nil {
		return nil, nil, fmt.Errorf("polígonos de cuadrilla: %w", err)
	}
	var out []poligonoCarga
	var rechazos []Rechazo
	for i, ft := range fc.Features {
		fila := i + 1
		geom, err := promoteMultiPolygon(ft.Geometry)
		if err != nil || bytes.Equal(bytes.TrimSpace(geom), []byte("null")) {
			rechazos = append(rechazos, Rechazo{Fuente: "jefe_de_grupo", Fila: fila, Campo: "geom", Motivo: "geometría inválida"})
			continue
		}
		props := ft.Properties
		_, per := soloNum(props["Perimetro"])
		_, area := soloNum(props["Área"])
		out = append(out, poligonoCarga{
			FeatureID:   fmt.Sprintf("PC-%04d", i+1),
			SourceIndex: i,
			Codigo:      strPtr(props["código"]),
			Nombre:      strPtr(props["Nombre"]),
			Uso:         strPtr(props["Uso"]),
			ProyRiego:   strPtr(props["Proy riego"]),
			RiegoAct:    strPtr(props["Riego act"]),
			Referencia:  cortar500(strPtr(props["Referenc_1"])),
			PerimetroM:  per,
			AreaM2:      area,
			Geom:        geom,
			JefesRaw:    propString(props, "jefes"),
			Origen:      fmt.Sprintf("jefe:%d", i),
		})
	}
	return out, rechazos, nil
}

func soloNum(v any) (json.RawMessage, *float64) {
	raw, f, ok := numRaw(v)
	if !ok {
		return nil, nil
	}
	return raw, f
}

func cortar500(s *string) *string {
	if s == nil {
		return nil
	}
	r := []rune(*s)
	if len(r) <= 500 {
		return s
	}
	cut := string(r[:500])
	return &cut
}

func leerLugares(body []byte) ([]lugarCarga, []Rechazo, error) {
	rows, err := readCSVBytes(body)
	if err != nil {
		return nil, nil, fmt.Errorf("lugares: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("lugares: csv vacío")
	}
	header := indexHeader(rows[0])
	var out []lugarCarga
	var rechazos []Rechazo
	vistos := map[string]int{}
	for i, rec := range rows[1:] {
		fila := i + 2
		nombre := strings.TrimSpace(cell(rec, header, "lugar"))
		if nombre == "" {
			continue
		}
		lat, lok := commaFloat(cell(rec, header, "latitud"))
		lon, ook := commaFloat(cell(rec, header, "longitud"))
		if !lok || !ook || !enCampus(lat, lon) {
			rechazos = append(rechazos, Rechazo{Fuente: "lugares", Fila: fila, Campo: "latitud", Motivo: "coordenada fuera del campus"})
			continue
		}
		norm := normalizarPersona(nombre)
		if prev, ok := vistos[norm]; ok {
			rechazos = append(rechazos, Rechazo{Fuente: "lugares", Fila: fila, Campo: "lugar", Motivo: fmt.Sprintf("nombre repetido respecto de la fila %d", prev)})
			continue
		}
		vistos[norm] = fila
		out = append(out, lugarCarga{
			Nombre: nombre,
			Norm:   norm,
			Lat:    lat,
			Lon:    lon,
			Origen: "lugar:" + norm,
			Fila:   fila,
		})
	}
	return out, rechazos, nil
}

func enCampus(lat, lon float64) bool {
	return lat >= latMin && lat <= latMax && lon >= lonMin && lon <= lonMax
}

func leerGvizFlora(body []byte) ([]ejemplarCarga, []Rechazo, error) {
	raw := bytes.TrimSpace(body)
	if i := bytes.IndexByte(raw, '{'); i >= 0 {
		if j := bytes.LastIndexByte(raw, '}'); j > i {
			raw = raw[i : j+1]
		}
	}
	var doc struct {
		Table struct {
			Cols []struct {
				Label string `json:"label"`
			} `json:"cols"`
			Rows []struct {
				C []struct {
					V any `json:"v"`
				} `json:"c"`
			} `json:"rows"`
		} `json:"table"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, nil, fmt.Errorf("gviz flora: %w", err)
	}
	cols := map[string]int{}
	for i, c := range doc.Table.Cols {
		cols[normalizarPersona(c.Label)] = i
	}
	idx := func(nombres ...string) int {
		for _, n := range nombres {
			key := normalizarPersona(n)
			if i, ok := cols[key]; ok {
				return i
			}
			for label, i := range cols {
				if strings.Contains(label, key) {
					return i
				}
			}
		}
		return -1
	}
	iNro := idx("N°", "Nº", "No")
	iUbi := idx("Ubicación", "Ubicacion")
	iRef := idx("Referencia")
	iLat := idx("Latitud")
	iLon := idx("Longitud")
	iCom := idx("Nombre común", "Nombre comun")
	iCie := idx("Nombre científico", "Nombre cientifico")
	iTipo := idx("Tipo de vegetación", "Tipo de vegetacion")
	iCant := idx("campus pucp Cantidad", "Cantidad")
	iCod := idx("Código", "Codigo")
	iObs := idx("OBSERVACIÓN FEN 2026", "Observación FEN 2026")
	var out []ejemplarCarga
	var rechazos []Rechazo
	vistos := map[int]int{}
	for n, row := range doc.Table.Rows {
		fila := n + 1
		val := func(i int) string {
			if i < 0 || i >= len(row.C) || row.C[i].V == nil {
				return ""
			}
			switch v := row.C[i].V.(type) {
			case string:
				return strings.TrimSpace(v)
			case float64:
				if v == float64(int64(v)) {
					return strconv.FormatInt(int64(v), 10)
				}
				return strconv.FormatFloat(v, 'f', -1, 64)
			default:
				return strings.TrimSpace(fmt.Sprint(v))
			}
		}
		numRaw := val(iNro)
		if numRaw == "" && val(iLat) == "" && val(iCie) == "" && val(iCom) == "" {
			continue
		}
		nro, err := strconv.Atoi(strings.TrimSpace(strings.Split(numRaw, ".")[0]))
		if err != nil || nro <= 0 {
			rechazos = append(rechazos, Rechazo{Fuente: "flora_gviz", Fila: fila, Campo: "N°", Motivo: "número de origen ausente"})
			continue
		}
		if prev, ok := vistos[nro]; ok {
			rechazos = append(rechazos, Rechazo{Fuente: "flora_gviz", Fila: fila, Campo: "N°", Motivo: fmt.Sprintf("duplicado del número %d (fila %d)", nro, prev)})
			continue
		}
		lat, lok := commaFloat(val(iLat))
		lon, ook := commaFloat(val(iLon))
		if !lok || !ook || !enCampus(lat, lon) {
			rechazos = append(rechazos, Rechazo{Fuente: "flora_gviz", Fila: fila, Campo: "Latitud", Motivo: "coordenada fuera del campus"})
			continue
		}
		tipo := canonTipo(val(iTipo))
		if val(iTipo) != "" && tipo == "" {
			rechazos = append(rechazos, Rechazo{Fuente: "flora_gviz", Fila: fila, Campo: "Tipo de vegetación", Motivo: "tipo fuera de catálogo"})
			continue
		}
		cant := 1
		if s := val(iCant); s != "" {
			if c, err := strconv.Atoi(strings.Split(s, ".")[0]); err == nil && c >= 1 {
				cant = c
			}
		}
		nroCopy := nro
		vistos[nro] = fila
		out = append(out, ejemplarCarga{
			NumeroOrigen: &nroCopy,
			Codigo:       limpiarTexto(val(iCod)),
			Cientifico:   limpiarTexto(val(iCie)),
			Comun:        limpiarTexto(val(iCom)),
			Tipo:         tipo,
			Cantidad:     cant,
			Lugar:        limpiarTexto(val(iUbi)),
			Referencia:   limpiarTexto(val(iRef)),
			Lat:          lat,
			Lon:          lon,
			Observacion:  limpiarTexto(val(iObs)),
			Origen:       fmt.Sprintf("gviz:%d", nro),
			Fila:         fila,
		})
	}
	return out, rechazos, nil
}

func canonTipo(s string) string {
	switch normalizarPersona(s) {
	case "arbol":
		return "Árbol"
	case "palmera":
		return "Palmera"
	case "arbusto":
		return "Arbusto"
	case "herbacea", "planta herbacea":
		return "Herbácea"
	case "trepadora":
		return "Trepadora"
	case "suculenta", "planta suculenta":
		return "Suculenta"
	case "cafeto":
		return "cafeto"
	default:
		return ""
	}
}

func limpiarTexto(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || strings.ContainsAny(s, "<") || strings.Contains(strings.ToLower(s), "http") {
		return ""
	}
	return s
}

// leerMedidasPalmera lee flora.csv. No corrige coordenadas mal tipeadas.
func leerMedidasPalmera(body []byte) ([]medidaCarga, []Rechazo, error) {
	return leerHojaPuntos(body, "flora", true)
}

func leerCafetos(body []byte) ([]ejemplarCarga, []medidaCarga, []Rechazo, error) {
	medidas, rechazos, err := leerHojaPuntos(body, "cafetos", false)
	if err != nil {
		return nil, nil, nil, err
	}
	var out []ejemplarCarga
	for _, m := range medidas {
		out = append(out, ejemplarCarga{
			Comun:      m.Comun,
			Cientifico: m.Cientifico,
			Tipo:       "cafeto",
			Cantidad:   1,
			Lugar:      m.Lugar,
			Lat:        m.Lat,
			Lon:        m.Lon,
			Origen:     m.Origen,
		})
	}
	return out, nil, rechazos, nil
}

func leerHojaPuntos(body []byte, fuente string, conMedidas bool) ([]medidaCarga, []Rechazo, error) {
	rows, err := readCSVBytes(body)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", fuente, err)
	}
	var out []medidaCarga
	var rechazos []Rechazo
	for i, rec := range rows {
		fila := i + 1
		if !filaNumerada(rec) {
			continue
		}
		lat, lon, ok := latLonDesdeColumnas(rec)
		if !ok {
			campo, motivo := motivoCoordenada(rec)
			rechazos = append(rechazos, Rechazo{Fuente: fuente, Fila: fila, Campo: campo, Motivo: motivo})
			continue
		}
		nro := strings.TrimSpace(rec[0])
		lugar, comun, cientifico := "", "", ""
		if len(rec) > 1 {
			lugar = limpiarTexto(rec[1])
		}
		if len(rec) > 6 {
			comun = limpiarTexto(rec[6])
		}
		if len(rec) > 7 {
			cientifico = limpiarTexto(rec[7])
		}
		item := medidaCarga{
			Comun:      comun,
			Cientifico: cientifico,
			Lugar:      lugar,
			Lat:        lat,
			Lon:        lon,
			Origen:     fmt.Sprintf("%s:%s", fuente, nro),
			ClaveUnion: claveUnion(lat, lon, cientifico, comun),
		}
		if conMedidas {
			item.UTMNorte = csvFloat(rec, 4)
			item.UTMEste = csvFloat(rec, 5)
			item.Altura = csvFloat(rec, 11)
			item.AlturaFuste = csvFloat(rec, 12)
			item.DAP = csvFloat(rec, 13)
			item.Radio = csvFloat(rec, 14)
			item.Zunchado = csvBool(rec, 15)
			if neg := medidaNegativa(item); neg != "" {
				rechazos = append(rechazos, Rechazo{Fuente: fuente, Fila: fila, Campo: neg, Motivo: "medida negativa"})
				continue
			}
		}
		out = append(out, item)
	}
	return out, rechazos, nil
}

func filaNumerada(rec []string) bool {
	if len(rec) == 0 {
		return false
	}
	s := strings.TrimSpace(rec[0])
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func latLonDesdeColumnas(rec []string) (float64, float64, bool) {
	if len(rec) < 4 {
		return 0, 0, false
	}
	lat, lok := commaFloat(rec[2])
	lon, ook := commaFloat(rec[3])
	if !lok || !ook || !enCampus(lat, lon) {
		return 0, 0, false
	}
	return lat, lon, true
}

func motivoCoordenada(rec []string) (string, string) {
	var lat, lon float64
	lok, ook := false, false
	if len(rec) > 2 {
		lat, lok = commaFloat(rec[2])
	}
	if len(rec) > 3 {
		lon, ook = commaFloat(rec[3])
	}
	latOK := lok && lat >= latMin && lat <= latMax
	lonOK := ook && lon >= lonMin && lon <= lonMax
	if !latOK {
		return "latitud", "latitud fuera del campus"
	}
	if !lonOK {
		return "longitud", "longitud fuera del campus"
	}
	return "latitud", "coordenada fuera del campus"
}

func csvFloat(rec []string, i int) *float64 {
	if i >= len(rec) {
		return nil
	}
	v, ok := commaFloat(rec[i])
	if !ok {
		return nil
	}
	return &v
}

func csvBool(rec []string, i int) *bool {
	if i >= len(rec) {
		return nil
	}
	switch normalizarPersona(rec[i]) {
	case "si", "sí", "s", "1", "true":
		v := true
		return &v
	case "no", "n", "0", "false":
		v := false
		return &v
	default:
		return nil
	}
}

func medidaNegativa(m medidaCarga) string {
	check := []struct {
		n *float64
		c string
	}{
		{m.Altura, "altura"},
		{m.AlturaFuste, "altura_fuste"},
		{m.DAP, "dap"},
		{m.Radio, "radio"},
	}
	for _, item := range check {
		if item.n != nil && *item.n < 0 {
			return item.c
		}
	}
	return ""
}

func claveUnion(lat, lon float64, cientifico, comun string) string {
	return fmt.Sprintf("%.5f|%.5f|%s|%s", lat, lon, normalizarPersona(cientifico), normalizarPersona(comun))
}

func leerActividades(body []byte) ([]tipoActividad, []Rechazo, error) {
	rows, err := readCSVBytes(body)
	if err != nil {
		return nil, nil, fmt.Errorf("actividades: %w", err)
	}
	var out []tipoActividad
	var rechazos []Rechazo
	clase := ""
	vistos := map[string]int{}
	for i, rec := range rows {
		fila := i + 1
		if i == 0 {
			continue
		}
		if len(rec) < 2 {
			continue
		}
		if c := strings.TrimSpace(rec[0]); c != "" {
			clase = c
		}
		tipo := ""
		if len(rec) > 1 {
			tipo = strings.TrimSpace(rec[1])
		}
		if tipo == "" {
			continue
		}
		if clase == "" {
			rechazos = append(rechazos, Rechazo{Fuente: "actividades", Fila: fila, Campo: "Clase", Motivo: "tipo sin clase"})
			continue
		}
		desc := ""
		if len(rec) > 2 {
			desc = limpiarTexto(rec[2])
		}
		clave := normalizarPersona(clase) + "|" + normalizarPersona(tipo)
		if prev, ok := vistos[clave]; ok {
			rechazos = append(rechazos, Rechazo{Fuente: "actividades", Fila: fila, Campo: "Tipo de actividad", Motivo: fmt.Sprintf("duplicado de la fila %d", prev)})
			continue
		}
		vistos[clave] = fila
		out = append(out, tipoActividad{
			Clase:       clase,
			Tipo:        tipo,
			Descripcion: desc,
			Origen:      "act:" + slug(clase) + ":" + slug(tipo),
		})
	}
	return out, rechazos, nil
}

func slug(s string) string {
	s = normalizarPersona(s)
	var b strings.Builder
	prev := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prev = false
			continue
		}
		if !prev {
			b.WriteByte('_')
			prev = true
		}
	}
	out := strings.Trim(b.String(), "_")
	if len(out) > 48 {
		out = out[:48]
	}
	if out == "" {
		out = "item"
	}
	return out
}

func readCSVBytes(body []byte) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(body))
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

func propString(props map[string]any, keys ...string) string {
	if props == nil {
		return ""
	}
	for _, key := range keys {
		if s, ok := props[key].(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func propFloat(props map[string]any, keys ...string) *float64 {
	if props == nil {
		return nil
	}
	for _, key := range keys {
		_, f, ok := numRaw(props[key])
		if ok {
			return f
		}
		if s, ok := props[key].(string); ok {
			if v, good := commaFloat(s); good {
				return &v
			}
		}
	}
	return nil
}
