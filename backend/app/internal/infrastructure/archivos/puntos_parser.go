// Package archivos provides file reading and parsing adapters.
package archivos

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type puntosParser struct{}

// NewPuntosParser creates a new IPuntosParser instance.
func NewPuntosParser() contracts.IPuntosParser {
	return &puntosParser{}
}

// LeerPuntosPUCP parses puntos PUCP CSV content and returns parsed points, rejections and omitted columns.
func (p *puntosParser) LeerPuntosPUCP(body []byte) ([]entities.PuntoCarga, []entities.RechazoFormato, []string, error) {
	records, err := csvRecords(body)
	if err != nil {
		return nil, nil, nil, err
	}
	if len(records) < 2 {
		return nil, nil, nil, nil
	}
	omitidas := []string{"phone", "phoneUnformatted", "placeId", "website", "image", "searchPageUrl"}
	header := indexHeader(records[0])
	var rows []entities.PuntoCarga
	var rech []entities.RechazoFormato
	for i, rec := range records[1:] {
		fila := i + 2
		titulo := cell(rec, header, "title")
		lat := coordDecimal(cell(rec, header, "location/lat"))
		lon := coordDecimal(cell(rec, header, "location/lng"))
		if titulo == "" || !puntoEnCampus(lat, lon) {
			rech = append(rech, entities.RechazoFormato{Fuente: "puntos_pucp", Fila: fila, Campo: "latitud", Motivo: "sin título o fuera del campus"})
			continue
		}
		url := UrlSinContacto(cell(rec, header, "url"))
		rows = append(rows, entities.PuntoCarga{
			Titulo: titulo,
			Lat:    lat,
			Lon:    lon,
			URL:    url,
			Origen: fmt.Sprintf("punto:%d", fila),
		})
	}
	return rows, rech, omitidas, nil
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

func indexHeader(header []string) map[string]int {
	out := map[string]int{}
	for i, h := range header {
		key := strings.ToLower(strings.TrimSpace(h))
		if key != "" {
			out[key] = i
		}
	}
	return out
}

func cell(rec []string, header map[string]int, name string) string {
	i, ok := header[name]
	if !ok || i >= len(rec) {
		return ""
	}
	v := strings.TrimSpace(rec[i])
	if strings.EqualFold(v, "null") {
		return ""
	}
	return v
}

func commaFloat(raw string) (float64, bool) {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, `"`)
	if s == "" || strings.EqualFold(s, "null") || strings.Contains(s, "/") {
		return 0, false
	}
	if strings.Count(s, ",") == 1 && !strings.Contains(s, ".") {
		s = strings.ReplaceAll(s, ",", ".")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func coordDecimal(raw string) float64 {
	v, ok := commaFloat(raw)
	if !ok {
		return 0
	}
	return v
}

func puntoEnCampus(lat, lon float64) bool {
	return lat <= -11.90 && lat >= -12.20 && lon <= -76.90 && lon >= -77.30
}

// UrlSinContacto removes placeId and contact parameters from URLs.
func UrlSinContacto(raw string) string {
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
