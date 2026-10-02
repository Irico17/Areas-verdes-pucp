package etl

import (
	"encoding/json"
	"strings"
)

type sectorImport struct {
	Codigo string `json:"codigo"`
	Nombre string `json:"nombre"`
	Color  string `json:"color"`
}

type viaImport struct {
	FeatureID string
	Nombre    string
	GeoJSON   string
}

func leerSectoresImport(csvBody, geoBody []byte) ([]sectorImport, []Rechazo, error) {
	if len(bytesTrim(csvBody)) == 0 {
		return sectoresDesdeGeo(geoBody)
	}
	filas, err := readCSVBytes(csvBody)
	if err != nil {
		return nil, nil, err
	}
	if len(filas) < 2 {
		return nil, nil, ErrSinValidas
	}
	header := indexHeader(filas[0])
	var out []sectorImport
	var rech []Rechazo
	vistos := map[string]int{}
	for i, rec := range filas[1:] {
		codigo := strings.ToLower(strings.TrimSpace(cell(rec, header, "codigo")))
		nombre := strings.TrimSpace(cell(rec, header, "nombre"))
		color := strings.ToLower(strings.TrimSpace(cell(rec, header, "color")))
		if codigo == "" && nombre == "" {
			continue
		}
		if codigo == "" || nombre == "" || len(color) != 7 || color[0] != '#' {
			rech = append(rech, Rechazo{Fuente: "sectores_capataz", Fila: i + 2, Campo: "codigo", Motivo: "código, nombre y color #rrggbb"})
			continue
		}
		if prev, ok := vistos[codigo]; ok {
			out[prev].Nombre = nombre
			out[prev].Color = color
			continue
		}
		vistos[codigo] = len(out)
		out = append(out, sectorImport{Codigo: codigo, Nombre: nombre, Color: color})
	}
	return out, rech, nil
}

func sectoresDesdeGeo(body []byte) ([]sectorImport, []Rechazo, error) {
	var fc struct {
		Features []struct {
			Properties map[string]any `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(body, &fc); err != nil {
		return nil, nil, err
	}
	var out []sectorImport
	var rech []Rechazo
	for i, f := range fc.Features {
		codigo, _ := f.Properties["codigo"].(string)
		nombre, _ := f.Properties["nombre"].(string)
		color, _ := f.Properties["color"].(string)
		codigo = strings.ToLower(strings.TrimSpace(codigo))
		nombre = strings.TrimSpace(nombre)
		color = strings.ToLower(strings.TrimSpace(color))
		if codigo == "" || nombre == "" || len(color) != 7 {
			rech = append(rech, Rechazo{Fuente: "sectores_capataz", Fila: i + 1, Campo: "codigo", Motivo: "código, nombre y color"})
			continue
		}
		out = append(out, sectorImport{Codigo: codigo, Nombre: nombre, Color: color})
	}
	return out, rech, nil
}

func leerViasImport(geoBody []byte) ([]viaImport, []Rechazo, error) {
	var fc struct {
		Features []struct {
			ID         string          `json:"id"`
			Geometry   json.RawMessage `json:"geometry"`
			Properties map[string]any  `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(geoBody, &fc); err != nil {
		return nil, nil, err
	}
	var out []viaImport
	var rech []Rechazo
	for i, f := range fc.Features {
		id := strings.TrimSpace(f.ID)
		if id == "" {
			if s, ok := f.Properties["feature_id"].(string); ok {
				id = strings.TrimSpace(s)
			}
		}
		if id == "" || len(f.Geometry) == 0 || string(f.Geometry) == "null" {
			rech = append(rech, Rechazo{Fuente: "vias", Fila: i + 1, Campo: "geometry", Motivo: "la vía necesita id y línea"})
			continue
		}
		nombre, _ := f.Properties["nombre"].(string)
		out = append(out, viaImport{FeatureID: id, Nombre: strings.TrimSpace(nombre), GeoJSON: string(f.Geometry)})
	}
	return out, rech, nil
}

func bytesTrim(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}
