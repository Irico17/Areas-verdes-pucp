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
		lat := coordDecimal(cell(rec, header, "location/lat"))
		lon := coordDecimal(cell(rec, header, "location/lng"))
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

func coordDecimal(raw string) float64 {
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
