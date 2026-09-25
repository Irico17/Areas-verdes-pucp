package etl

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	ViveroVivo  = 689
	ViveroLocal = 683
	MonitoreoN  = 283
	PodaN       = 25
)

// ErrorFila es un rechazo o un aviso de una fila de importación.
type ErrorFila struct {
	Fila   int    `json:"fila"`
	Campo  string `json:"campo"`
	Motivo string `json:"motivo"`
}

// LaborImportada es una fila de monitoreo lista para upsert.
type LaborImportada struct {
	OrigenRef      string
	Clase          string
	Tipo           string
	Estado         string
	FechaSolicitud *time.Time
	FechaAtencion  *time.Time
	Lugar          string
	Lat            *float64
	Lon            *float64
	Titulo         string
	Comentario     string
	Detalle        string
	Cuadrilla      string
	Ejecutor       string
	FotoOmitida    bool
}

// PodaImportada es un registro PO-n.
type PodaImportada struct {
	Codigo            string
	CodigoExterno     string
	Tipo              string
	FechaReporte      *time.Time
	FechaEjecucion    *time.Time
	Personal          string
	Ubicacion         string
	Unidad            string
	CantidadPedida    float64
	CantidadEjecutada float64
	TipoActividad     string
	TipoVegetacion    string
	NombreComun       string
	NombreCientifico  string
	Prioridad         string
	Comentario        string
	OrigenRef         string
}

// ViveroImportado es una fila de vivero.
type ViveroImportado struct {
	OrigenRef     string
	Fecha         *time.Time
	Area          string
	Subproceso    string
	Etapa         string
	Descripcion   string
	Observaciones string
	Responsables  string
	LugarID       string
	LugarLibre    string
}

// ReporteImportacion resume la carga sin datos personales.
type ReporteImportacion struct {
	Labores         int         `json:"labores"`
	SinEstado       int         `json:"sin_estado"`
	ConCoordenada   int         `json:"con_coordenada"`
	SinResponsable  int         `json:"sin_responsable"`
	Podas           int         `json:"podas"`
	CodigosExternos int         `json:"codigos_externos"`
	Vivero          int         `json:"vivero"`
	ViveroVivo      int         `json:"vivero_vivo"`
	ViveroDelta     string      `json:"vivero_delta"`
	LugarLibre      int         `json:"lugar_libre"`
	Errores         []ErrorFila `json:"errores"`
	Avisos          []ErrorFila `json:"avisos"`
}

// CodigoExterno conserva OSG-… si viene. No inventa un código.
func CodigoExterno(valor string) string {
	v := strings.TrimSpace(valor)
	if v == "" {
		return ""
	}
	norma := strings.ToLower(normalizarPersona(v))
	switch norma {
	case "aun no tiene codigo", "no aplica", "sin codigo", "s/c":
		return ""
	}
	if strings.HasPrefix(strings.ToUpper(v), "OSG-") {
		return strings.ToUpper(v)
	}
	return ""
}

func uuidOrigen(kind, ref string) string {
	sum := sha256.Sum256([]byte(kind + ":" + ref))
	h := hex.EncodeToString(sum[:16])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

// IDOrigen es el UUID estable de una fila importada.
func IDOrigen(kind, ref string) string {
	return uuidOrigen(kind, ref)
}

func parseFecha(valor string) (*time.Time, error) {
	v := strings.TrimSpace(valor)
	if v == "" {
		return nil, nil
	}
	for _, layout := range []string{"02/01/2006", "2/1/2006", "2006-01-02"} {
		if t, err := time.Parse(layout, v); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("fecha no reconocida")
}

func parseNumero(valor string) (float64, error) {
	v := strings.TrimSpace(strings.ReplaceAll(valor, ",", "."))
	if v == "" {
		return 0, nil
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func parseCoord(valor string) (*float64, error) {
	v := strings.TrimSpace(strings.Trim(valor, `"`))
	v = strings.ReplaceAll(v, ",", ".")
	if v == "" {
		return nil, nil
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func enCampus(lat, lon float64) bool {
	return lat <= -11.90 && lat >= -12.20 && lon <= -76.90 && lon >= -77.30
}

func leerCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
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

func col(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

// ImportarMonitoreo lee la hoja de 2026. El estado vacío queda en sin_estado.
func ImportarMonitoreo(path string, tabla *Tabla) ([]LaborImportada, ReporteImportacion, error) {
	rows, err := leerCSV(path)
	if err != nil {
		return nil, ReporteImportacion{}, err
	}
	if len(rows) == 0 {
		return nil, ReporteImportacion{}, fmt.Errorf("monitoreo vacío")
	}
	var nombres []string
	for _, row := range rows[1:] {
		if !filaConDato(row) {
			continue
		}
		if v := col(row, 12); v != "" {
			nombres = append(nombres, v)
		}
	}
	tabla.Aplicar(nombres)

	var out []LaborImportada
	rep := ReporteImportacion{ViveroVivo: ViveroVivo}
	for i, row := range rows[1:] {
		if !filaConDato(row) {
			continue
		}
		fila := i + 2
		item := LaborImportada{
			OrigenRef:  fmt.Sprintf("M-%d", len(out)+1),
			Clase:      col(row, 0),
			Ejecutor:   "propia",
			Comentario: col(row, 9),
			Detalle:    col(row, 13),
			Lugar:      col(row, 5),
		}
		estado := col(row, 1)
		if estado == "" {
			item.Estado = "sin_estado"
			rep.SinEstado++
		} else if strings.EqualFold(estado, "cerrado") || strings.EqualFold(estado, "cerrada") {
			item.Estado = "cerrada"
		} else {
			item.Estado = "sin_estado"
			rep.Avisos = append(rep.Avisos, ErrorFila{Fila: fila, Campo: "estado", Motivo: "estado no catalogado; se guardó sin_estado"})
			rep.SinEstado++
		}
		fs, err := parseFecha(col(row, 2))
		if err != nil {
			rep.Errores = append(rep.Errores, ErrorFila{Fila: fila, Campo: "fecha_solicitud", Motivo: "fecha no reconocida"})
		}
		fa, err := parseFecha(col(row, 3))
		if err != nil {
			rep.Errores = append(rep.Errores, ErrorFila{Fila: fila, Campo: "fecha_atencion", Motivo: "fecha no reconocida"})
		}
		if fs != nil && fa != nil && fa.Before(*fs) {
			rep.Errores = append(rep.Errores, ErrorFila{Fila: fila, Campo: "fecha_atencion", Motivo: "la atención es anterior a la solicitud"})
			fa = nil
		}
		item.FechaSolicitud = fs
		item.FechaAtencion = fa
		lat, errLat := parseCoord(col(row, 6))
		lon, errLon := parseCoord(col(row, 7))
		if errLat != nil || errLon != nil {
			rep.Errores = append(rep.Errores, ErrorFila{Fila: fila, Campo: "coordenada", Motivo: "coordenada ilegible"})
		} else if lat != nil && lon != nil {
			if !enCampus(*lat, *lon) {
				rep.Errores = append(rep.Errores, ErrorFila{Fila: fila, Campo: "coordenada", Motivo: "fuera del campus"})
			} else {
				item.Lat = lat
				item.Lon = lon
				rep.ConCoordenada++
			}
		}
		item.Tipo = col(row, 8)
		if item.Tipo == "" {
			rep.Avisos = append(rep.Avisos, ErrorFila{Fila: fila, Campo: "tipo", Motivo: "sin actividad en la hoja; se conserva la fila"})
		}
		resp := col(row, 12)
		if resp == "" {
			rep.SinResponsable++
		} else {
			item.Cuadrilla = tabla.Ficticio(resp)
		}
		foto := col(row, 11)
		if foto != "" {
			item.FotoOmitida = true
			if strings.Contains(foto, "drive.google.com") || strings.Contains(foto, "http") {
				rep.Avisos = append(rep.Avisos, ErrorFila{Fila: fila, Campo: "foto", Motivo: "no se guarda la URL; hace falta copia al almacén propio"})
			}
		}
		item.Titulo = tituloLabor(item)
		if utf8.RuneCountInString(item.Titulo) > 160 {
			item.Titulo = string([]rune(item.Titulo)[:160])
		}
		out = append(out, item)
	}
	rep.Labores = len(out)
	return out, rep, nil
}

func tituloLabor(item LaborImportada) string {
	if t := strings.TrimSpace(item.Tipo); t != "" {
		return t
	}
	if t := strings.TrimSpace(item.Comentario); t != "" {
		return t
	}
	if t := strings.TrimSpace(item.Detalle); t != "" {
		return t
	}
	if item.Lugar != "" {
		return "Labor en " + item.Lugar
	}
	return "Labor " + item.OrigenRef
}

func filaConDato(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return true
		}
	}
	return false
}

// ImportarPoda lee poda.csv. La primera fila es un rótulo; la cabecera es la segunda.
func ImportarPoda(path string, tabla *Tabla, lugares map[string]string) ([]PodaImportada, []ErrorFila, error) {
	rows, err := leerCSV(path)
	if err != nil {
		return nil, nil, err
	}
	if len(rows) < 2 {
		return nil, nil, fmt.Errorf("poda sin cabecera")
	}
	var nombres []string
	for _, row := range rows[2:] {
		if !filaConDato(row) {
			continue
		}
		if v := col(row, 6); v != "" {
			nombres = append(nombres, v)
		}
	}
	tabla.Aplicar(nombres)
	var out []PodaImportada
	var errores []ErrorFila
	vistos := map[string]bool{}
	for i, row := range rows[2:] {
		if !filaConDato(row) {
			continue
		}
		fila := i + 3
		codigo := col(row, 0)
		if !strings.HasPrefix(codigo, "PO-") {
			errores = append(errores, ErrorFila{Fila: fila, Campo: "codigo", Motivo: "se esperaba PO-n"})
			continue
		}
		if vistos[codigo] {
			errores = append(errores, ErrorFila{Fila: fila, Campo: "codigo", Motivo: "código repetido"})
			continue
		}
		vistos[codigo] = true
		pedida, err1 := parseNumero(col(row, 9))
		ejec, err2 := parseNumero(col(row, 10))
		if err1 != nil || err2 != nil || pedida < 0 || ejec < 0 {
			errores = append(errores, ErrorFila{Fila: fila, Campo: "cantidad", Motivo: "cantidad inválida"})
			continue
		}
		fr, err := parseFecha(col(row, 4))
		if err != nil {
			errores = append(errores, ErrorFila{Fila: fila, Campo: "fecha_reporte", Motivo: "fecha no reconocida"})
		}
		fe, err := parseFecha(col(row, 5))
		if err != nil {
			errores = append(errores, ErrorFila{Fila: fila, Campo: "fecha_ejecucion", Motivo: "fecha no reconocida"})
		}
		if fr != nil && fe != nil && fe.Before(*fr) {
			errores = append(errores, ErrorFila{Fila: fila, Campo: "fecha_ejecucion", Motivo: "la ejecución es anterior al reporte"})
			fe = nil
		}
		prioridad := strings.ToLower(col(row, 17))
		switch prioridad {
		case "baja", "media", "alta":
		default:
			if prioridad == "" {
				prioridad = "media"
			} else {
				errores = append(errores, ErrorFila{Fila: fila, Campo: "prioridad", Motivo: "prioridad no catalogada"})
				prioridad = "media"
			}
		}
		ubic := col(row, 7)
		item := PodaImportada{
			Codigo:            codigo,
			CodigoExterno:     CodigoExterno(col(row, 1)),
			Tipo:              col(row, 2),
			FechaReporte:      fr,
			FechaEjecucion:    fe,
			Personal:          tabla.Ficticio(col(row, 6)),
			Ubicacion:         ubic,
			Unidad:            col(row, 8),
			CantidadPedida:    pedida,
			CantidadEjecutada: ejec,
			TipoActividad:     col(row, 11),
			TipoVegetacion:    col(row, 12),
			NombreComun:       col(row, 13),
			NombreCientifico:  col(row, 14),
			Prioridad:         prioridad,
			Comentario:        col(row, 16),
			OrigenRef:         codigo,
		}
		if ubic != "" && lugares != nil {
			if _, ok := lugares[normalizarPersona(ubic)]; !ok {
				errores = append(errores, ErrorFila{Fila: fila, Campo: "ubicacion", Motivo: "sin lugar conocido; se guarda el texto"})
			}
		}
		out = append(out, item)
	}
	return out, errores, nil
}

// ImportarVivero lee el CSV local. El delta contra el vivo publicado queda en el reporte.
func ImportarVivero(path string, tabla *Tabla, lugares map[string]string) ([]ViveroImportado, ReporteImportacion, error) {
	rows, err := leerCSV(path)
	if err != nil {
		return nil, ReporteImportacion{}, err
	}
	var nombres []string
	for _, row := range rows[1:] {
		if !filaConDato(row) {
			continue
		}
		if v := col(row, 5); v != "" {
			nombres = append(nombres, strings.Split(v, ",")...)
		}
	}
	tabla.Aplicar(nombres)
	var out []ViveroImportado
	rep := ReporteImportacion{ViveroVivo: ViveroVivo}
	areas := map[string]bool{"fauna": true, "flora": true, "ambiental": true, "otros": true}
	for i, row := range rows[1:] {
		if !filaConDato(row) {
			continue
		}
		fila := i + 2
		area := col(row, 1)
		normaArea := normalizarPersona(area)
		if area == "" {
			rep.Avisos = append(rep.Avisos, ErrorFila{Fila: fila, Campo: "area", Motivo: "área vacía"})
		} else if !areas[normaArea] {
			rep.Avisos = append(rep.Avisos, ErrorFila{Fila: fila, Campo: "area", Motivo: "área fuera de catálogo; se conserva el texto"})
		}
		if normaArea == "otros" {
			area = "Otros"
		}
		fecha, err := parseFecha(col(row, 0))
		if err != nil {
			rep.Errores = append(rep.Errores, ErrorFila{Fila: fila, Campo: "fecha", Motivo: "fecha no reconocida"})
		}
		lugar := col(row, 7)
		item := ViveroImportado{
			OrigenRef:     fmt.Sprintf("V-%d", len(out)+1),
			Fecha:         fecha,
			Area:          area,
			Subproceso:    col(row, 2),
			Etapa:         col(row, 3),
			Descripcion:   col(row, 4),
			Observaciones: col(row, 6),
			Responsables:  tabla.Lista(col(row, 5)),
		}
		if lugar != "" {
			if id, ok := lugares[normalizarPersona(lugar)]; ok {
				item.LugarID = id
			} else {
				item.LugarLibre = lugar
				rep.LugarLibre++
			}
		}
		out = append(out, item)
	}
	rep.Vivero = len(out)
	rep.ViveroDelta = fmt.Sprintf(
		"copia local %d filas no vacías; publicado el 2026-09-25: %d. Delta %d, la copia del repo no trae las filas que solo están en la hoja viva.",
		rep.Vivero, ViveroVivo, ViveroVivo-rep.Vivero,
	)
	return out, rep, nil
}
