package etl

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

// archivoListo es el mismo contenido en CSV y en GeoJSON para alimentar al lector.
type archivoListo struct {
	Formato string
	CSV     []byte
	Geo     []byte
	JSON    []byte
}

func prepararArchivo(entidad, nombre string, body []byte) (archivoListo, error) {
	body = bytes.TrimPrefix(body, []byte{0xEF, 0xBB, 0xBF})
	formato, err := detectarFormato(nombre, body)
	if err != nil {
		return archivoListo{}, err
	}
	switch formato {
	case "json":
		return archivoListo{Formato: "json", JSON: bytes.TrimSpace(body)}, nil
	case "xlsx":
		filas, err := leerXLSX(entidad, body)
		if err != nil {
			return archivoListo{}, err
		}
		csvBody, err := filasCSV(filas)
		if err != nil {
			return archivoListo{}, err
		}
		geo, err := filasAGeoJSON(filas)
		if err != nil {
			return archivoListo{}, err
		}
		return archivoListo{Formato: formato, CSV: csvBody, Geo: geo}, nil
	case "geojson":
		geo := bytes.TrimSpace(body)
		filas, err := geojsonAFilas(geo)
		if err != nil {
			return archivoListo{}, err
		}
		csvBody, err := filasCSV(filas)
		if err != nil {
			return archivoListo{}, err
		}
		return archivoListo{Formato: formato, CSV: csvBody, Geo: geo, JSON: geo}, nil
	default:
		filas, err := readCSVBytes(body)
		if err != nil {
			return archivoListo{}, fmt.Errorf("el CSV no se pudo leer: %w", err)
		}
		geo, err := filasAGeoJSON(filas)
		if err != nil {
			return archivoListo{}, err
		}
		return archivoListo{Formato: "csv", CSV: body, Geo: geo}, nil
	}
}

func detectarFormato(nombre string, body []byte) (string, error) {
	ext := strings.ToLower(path.Ext(nombre))
	trim := bytes.TrimSpace(body)
	switch {
	case ext == ".xlsx" || bytes.HasPrefix(trim, []byte("PK")):
		return "xlsx", nil
	case ext == ".geojson" || ext == ".json" || (len(trim) > 0 && trim[0] == '{'):
		if bytes.Contains(trim, []byte(`"FeatureCollection"`)) || bytes.Contains(trim, []byte(`"features"`)) {
			return "geojson", nil
		}
		return "json", nil
	case ext == ".csv" || ext == ".txt" || ext == "":
		return "csv", nil
	default:
		return "", fmt.Errorf("formato no admitido: use CSV, XLSX o GeoJSON")
	}
}

func filasCSV(filas [][]string) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	for _, fila := range filas {
		if err := w.Write(fila); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func filasAGeoJSON(filas [][]string) ([]byte, error) {
	if len(filas) == 0 {
		return []byte(`{"type":"FeatureCollection","features":[]}`), nil
	}
	header := filas[0]
	features := make([]map[string]any, 0, len(filas)-1)
	for _, rec := range filas[1:] {
		props := map[string]any{}
		for i, key := range header {
			if strings.TrimSpace(key) == "" {
				continue
			}
			props[key] = col(rec, i)
		}
		features = append(features, map[string]any{
			"type":       "Feature",
			"geometry":   geometriaDeProps(props),
			"properties": props,
		})
	}
	raw, err := json.Marshal(map[string]any{"type": "FeatureCollection", "features": features})
	return raw, err
}

func geometriaDeProps(props map[string]any) any {
	if g, ok := props["geom"].(string); ok && strings.TrimSpace(g) != "" {
		var parsed any
		if json.Unmarshal([]byte(g), &parsed) == nil {
			return parsed
		}
	}
	lat, lok := primeraCoord(props, "latitud", "lat", "Latitud", "location/lat")
	lon, ook := primeraCoord(props, "longitud", "lon", "lng", "Longitud", "location/lng")
	if lok && ook {
		return map[string]any{"type": "Point", "coordinates": []float64{lon, lat}}
	}
	return nil
}

func primeraCoord(props map[string]any, claves ...string) (float64, bool) {
	for _, clave := range claves {
		raw, ok := props[clave].(string)
		if !ok {
			continue
		}
		n, ok := commaFloat(raw)
		if ok {
			return n, true
		}
	}
	return 0, false
}

func geojsonAFilas(body []byte) ([][]string, error) {
	var fc struct {
		Features []struct {
			Geometry   json.RawMessage `json:"geometry"`
			Properties map[string]any  `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(body, &fc); err != nil {
		return nil, fmt.Errorf("GeoJSON ilegible: %w", err)
	}
	orden := []string{}
	visto := map[string]bool{}
	for _, ft := range fc.Features {
		for k := range ft.Properties {
			if !visto[k] {
				visto[k] = true
				orden = append(orden, k)
			}
		}
	}
	if !visto["latitud"] {
		orden = append(orden, "latitud", "longitud")
	}
	filas := [][]string{orden}
	for _, ft := range fc.Features {
		rec := make([]string, len(orden))
		lat, lon := puntoDe(ft.Geometry)
		for i, k := range orden {
			switch k {
			case "latitud":
				if lat != 0 {
					rec[i] = strconv.FormatFloat(lat, 'f', -1, 64)
					continue
				}
			case "longitud":
				if lon != 0 {
					rec[i] = strconv.FormatFloat(lon, 'f', -1, 64)
					continue
				}
			}
			rec[i] = propString(ft.Properties, k)
		}
		filas = append(filas, rec)
	}
	return filas, nil
}

func csvAGviz(body []byte) ([]byte, error) {
	filas, err := readCSVBytes(body)
	if err != nil {
		return nil, err
	}
	if len(filas) == 0 {
		return nil, fmt.Errorf("el CSV no trae cabecera")
	}
	cols := make([]map[string]string, len(filas[0]))
	for i, h := range filas[0] {
		cols[i] = map[string]string{"label": h}
	}
	rows := make([]map[string]any, 0, len(filas)-1)
	for _, rec := range filas[1:] {
		celdas := make([]map[string]any, len(filas[0]))
		for i := range filas[0] {
			celdas[i] = map[string]any{"v": col(rec, i)}
		}
		rows = append(rows, map[string]any{"c": celdas})
	}
	return json.Marshal(map[string]any{"table": map[string]any{"cols": cols, "rows": rows}})
}

func reservasDesdeTabular(body []byte) ([]byte, error) {
	filas, err := readCSVBytes(body)
	if err != nil {
		return nil, err
	}
	if len(filas) < 2 {
		return []byte(`{"reservas":[]}`), nil
	}
	header := indexHeader(filas[0])
	type item struct {
		ID                string `json:"id"`
		Jardin            string `json:"jardin"`
		JardinCodigo      string `json:"jardin_codigo"`
		Fecha             string `json:"fecha"`
		Hora              string `json:"hora"`
		Evento            string `json:"evento"`
		UnidadResponsable string `json:"unidad_responsable"`
		Estado            string `json:"estado"`
		Fake              bool   `json:"_fake"`
	}
	var reservas []item
	for i, rec := range filas[1:] {
		origen := strings.ToLower(strings.TrimSpace(cell(rec, header, "origen")))
		id := cell(rec, header, "id")
		falso := origen == "ficticio"
		if id == "" {
			if falso {
				id = fmt.Sprintf("FAKE-%d", i+1)
			} else {
				id = fmt.Sprintf("fila-%d", i+1)
			}
		}
		if falso && !strings.HasPrefix(id, "FAKE-") {
			id = "FAKE-" + id
		}
		inicio := cell(rec, header, "hora_inicio")
		fin := cell(rec, header, "hora_fin")
		hora := cell(rec, header, "hora")
		if hora == "" && inicio != "" {
			hora = inicio + "-" + fin
		}
		reservas = append(reservas, item{
			ID:                id,
			Jardin:            primero(cell(rec, header, "jardin"), cell(rec, header, "jardin_id")),
			JardinCodigo:      primero(cell(rec, header, "jardin_codigo"), cell(rec, header, "jardin_id")),
			Fecha:             cell(rec, header, "fecha"),
			Hora:              hora,
			Evento:            cell(rec, header, "evento"),
			UnidadResponsable: cell(rec, header, "unidad"),
			Estado:            cell(rec, header, "estado"),
			Fake:              falso,
		})
	}
	return json.Marshal(map[string]any{"reservas": reservas})
}

func primero(valores ...string) string {
	for _, v := range valores {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func leerXLSX(entidad string, body []byte) ([][]string, error) {
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, fmt.Errorf("xlsx ilegible: %w", err)
	}
	archivos := map[string]*zip.File{}
	for _, f := range zr.File {
		archivos[f.Name] = f
	}
	libro, err := leerZip(archivos, "xl/workbook.xml")
	if err != nil {
		return nil, err
	}
	rels, err := leerZip(archivos, "xl/_rels/workbook.xml.rels")
	if err != nil {
		return nil, err
	}
	nombres := hojasXLSX(libro)
	destinos := relsXLSX(rels)
	var elegida string
	for _, hoja := range nombres {
		if hoja.nombre == entidad {
			elegida = destinos[hoja.rid]
			break
		}
	}
	if elegida == "" {
		return nil, fmt.Errorf("el xlsx debe traer una hoja llamada %s", entidad)
	}
	if !strings.HasPrefix(elegida, "xl/") {
		elegida = path.Join("xl", elegida)
	}
	hoja, err := leerZip(archivos, elegida)
	if err != nil {
		return nil, err
	}
	textos, _ := leerZip(archivos, "xl/sharedStrings.xml")
	return filasHoja(hoja, textosCompartidos(textos)), nil
}

func leerZip(archivos map[string]*zip.File, nombre string) ([]byte, error) {
	f, ok := archivos[nombre]
	if !ok {
		return nil, fmt.Errorf("al xlsx le falta %s", nombre)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

type hojaRef struct {
	nombre string
	rid    string
}

func hojasXLSX(body []byte) []hojaRef {
	dec := xml.NewDecoder(bytes.NewReader(body))
	var out []hojaRef
	for {
		tok, err := dec.Token()
		if err != nil {
			return out
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "sheet" {
			continue
		}
		item := hojaRef{}
		for _, attr := range se.Attr {
			switch attr.Name.Local {
			case "name":
				item.nombre = attr.Value
			case "id":
				item.rid = attr.Value
			}
		}
		out = append(out, item)
	}
}

func relsXLSX(body []byte) map[string]string {
	dec := xml.NewDecoder(bytes.NewReader(body))
	out := map[string]string{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return out
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "Relationship" {
			continue
		}
		var id, target string
		for _, attr := range se.Attr {
			switch attr.Name.Local {
			case "Id":
				id = attr.Value
			case "Target":
				target = attr.Value
			}
		}
		out[id] = target
	}
}

func textosCompartidos(body []byte) []string {
	if len(body) == 0 {
		return nil
	}
	dec := xml.NewDecoder(bytes.NewReader(body))
	var out []string
	var actual strings.Builder
	dentro := false
	for {
		tok, err := dec.Token()
		if err != nil {
			return out
		}
		switch el := tok.(type) {
		case xml.StartElement:
			if el.Name.Local == "si" {
				actual.Reset()
				dentro = true
			}
		case xml.EndElement:
			if el.Name.Local == "si" {
				out = append(out, actual.String())
				dentro = false
			}
		case xml.CharData:
			if dentro {
				actual.Write(el)
			}
		}
	}
}

func filasHoja(body []byte, compartidos []string) [][]string {
	dec := xml.NewDecoder(bytes.NewReader(body))
	var filas [][]string
	var fila []string
	var tipo string
	var ref string
	enV := false
	enT := false
	var valor strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			if len(fila) > 0 {
				filas = append(filas, fila)
			}
			return filas
		}
		switch el := tok.(type) {
		case xml.StartElement:
			switch el.Name.Local {
			case "row":
				fila = nil
			case "c":
				tipo = ""
				ref = ""
				for _, attr := range el.Attr {
					switch attr.Name.Local {
					case "t":
						tipo = attr.Value
					case "r":
						ref = attr.Value
					}
				}
				fila = rellenarHuecos(fila, ref)
			case "v":
				enV = true
				valor.Reset()
			case "t":
				enT = true
				valor.Reset()
			}
		case xml.EndElement:
			switch el.Name.Local {
			case "row":
				filas = append(filas, fila)
			case "v":
				enV = false
				fila = append(fila, celdaXLSX(tipo, valor.String(), compartidos))
			case "t":
				if tipo == "inlineStr" {
					enT = false
					fila = append(fila, valor.String())
				}
			}
		case xml.CharData:
			if enV || enT {
				valor.Write(el)
			}
		}
	}
}

func rellenarHuecos(fila []string, ref string) []string {
	col := 0
	for _, r := range ref {
		if r < 'A' || r > 'Z' {
			break
		}
		col = col*26 + int(r-'A'+1)
	}
	if col == 0 {
		return fila
	}
	for len(fila) < col-1 {
		fila = append(fila, "")
	}
	return fila
}

func celdaXLSX(tipo, valor string, compartidos []string) string {
	if tipo == "s" {
		i, err := strconv.Atoi(valor)
		if err == nil && i >= 0 && i < len(compartidos) {
			return compartidos[i]
		}
	}
	return valor
}
