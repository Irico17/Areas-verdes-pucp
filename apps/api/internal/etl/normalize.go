package etl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Record es una feature normalizada, lista para data/v1 y para PostGIS.
// El campo de personas "jefes" no se copia.
type Record struct {
	FeatureID    string
	SourceIndex  int
	Capa         string
	Codigo       *string
	Nombre       *string
	Uso          *string
	Clase        *string
	ProyRiego    *string
	RiegoAct     *string
	Referencia   *string
	Pertenecen   *string
	PerimetroM   *float64
	AreaM2       *float64
	PerimetroRaw json.RawMessage
	AreaRaw      json.RawMessage
	Geometry     json.RawMessage
}

type rawFC struct {
	Type     string       `json:"type"`
	Features []rawFeature `json:"features"`
}

type rawFeature struct {
	Type       string          `json:"type"`
	Geometry   json.RawMessage `json:"geometry"`
	Properties map[string]any  `json:"properties"`
}

type geomHead struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

// NormalizeAreas lee areas_verdes.geojson.
func NormalizeAreas(data []byte) ([]Record, error) {
	return normalize(data, "areas_verdes", "AV", false)
}

// NormalizeZonas lee jefe_de_grupo.json y descarta el campo jefes.
func NormalizeZonas(data []byte) ([]Record, error) {
	return normalize(data, "zonas", "Z", true)
}

// NormalizeCapa lee una capa opcional (jardines_reserva, xerofitica).
func NormalizeCapa(data []byte, capa, idPrefix string) ([]Record, error) {
	return normalize(data, capa, idPrefix, false)
}

func normalize(data []byte, capa, idPrefix string, dropJefes bool) ([]Record, error) {
	fc, err := parseFC(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", capa, err)
	}
	if fc.Type != "FeatureCollection" {
		return nil, fmt.Errorf("%s: se esperaba FeatureCollection, llegó %q", capa, fc.Type)
	}

	out := make([]Record, 0, len(fc.Features))
	for i, ft := range fc.Features {
		if ft.Type != "" && ft.Type != "Feature" {
			return nil, fmt.Errorf("%s: feature %d tipo %q", capa, i, ft.Type)
		}
		geom, err := promoteMultiPolygon(ft.Geometry)
		if err != nil {
			return nil, fmt.Errorf("%s: feature %d: %w", capa, i, err)
		}
		props := ft.Properties
		if props == nil {
			props = map[string]any{}
		}
		if dropJefes {
			delete(props, "jefes")
		}

		perRaw, per, _ := numRaw(props["Perimetro"])
		areaRaw, area, _ := numRaw(props["Área"])

		out = append(out, Record{
			FeatureID:    fmt.Sprintf("%s-%04d", idPrefix, i+1),
			SourceIndex:  i,
			Capa:         capa,
			Codigo:       strPtr(props["código"]),
			Nombre:       strPtr(props["Nombre"]),
			Uso:          strPtr(props["Uso"]),
			Clase:        strPtr(props["Class"]),
			ProyRiego:    strPtr(props["Proy riego"]),
			RiegoAct:     strPtr(props["Riego act"]),
			Referencia:   strPtr(props["Referenc_1"]),
			Pertenecen:   strPtr(props["Pertenecen"]),
			PerimetroM:   per,
			AreaM2:       area,
			PerimetroRaw: perRaw,
			AreaRaw:      areaRaw,
			Geometry:     geom,
		})
	}
	return out, nil
}

func parseFC(data []byte) (rawFC, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var fc rawFC
	if err := dec.Decode(&fc); err != nil {
		return fc, err
	}
	return fc, nil
}

// promoteMultiPolygon envuelve Polygon sin reescribir el texto de las coordenadas.
func promoteMultiPolygon(geom json.RawMessage) (json.RawMessage, error) {
	trimmed := bytes.TrimSpace(geom)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return json.RawMessage("null"), nil
	}
	var head geomHead
	if err := json.Unmarshal(trimmed, &head); err != nil {
		return nil, fmt.Errorf("geometría: %w", err)
	}
	coords := bytes.TrimSpace(head.Coordinates)
	if len(coords) == 0 {
		return nil, fmt.Errorf("geometría sin coordinates")
	}
	switch head.Type {
	case "MultiPolygon":
		return json.RawMessage(fmt.Sprintf(`{"type":"MultiPolygon","coordinates":%s}`, coords)), nil
	case "Polygon":
		return json.RawMessage(fmt.Sprintf(`{"type":"MultiPolygon","coordinates":[%s]}`, coords)), nil
	default:
		return nil, fmt.Errorf("tipo de geometría %q; se espera Polygon o MultiPolygon", head.Type)
	}
}

func strPtr(v any) *string {
	s, ok := v.(string)
	if !ok {
		return nil
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func numRaw(v any) (json.RawMessage, *float64, bool) {
	switch n := v.(type) {
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return nil, nil, false
		}
		return json.RawMessage(n.String()), &f, true
	case float64:
		raw := json.RawMessage(strconv.FormatFloat(n, 'f', 3, 64))
		f := n
		return raw, &f, true
	default:
		return nil, nil, false
	}
}
