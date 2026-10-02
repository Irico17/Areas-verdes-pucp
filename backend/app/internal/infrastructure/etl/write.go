package etl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type v1Props struct {
	FeatureID   string          `json:"feature_id"`
	SourceIndex int             `json:"source_index"`
	Codigo      *string         `json:"codigo,omitempty"`
	Nombre      *string         `json:"nombre,omitempty"`
	Uso         *string         `json:"uso,omitempty"`
	ProyRiego   *string         `json:"proy_riego,omitempty"`
	RiegoAct    *string         `json:"riego_act,omitempty"`
	Referencia  *string         `json:"referencia,omitempty"`
	Clase       *string         `json:"clase,omitempty"`
	Pertenecen  *string         `json:"pertenecen,omitempty"`
	Capa        string          `json:"capa,omitempty"`
	PerimetroM  json.RawMessage `json:"perimetro_m,omitempty"`
	AreaM2      json.RawMessage `json:"area_m2,omitempty"`
}

type v1Feature struct {
	Type       string          `json:"type"`
	ID         string          `json:"id"`
	Geometry   json.RawMessage `json:"geometry"`
	Properties v1Props         `json:"properties"`
}

type v1Collection struct {
	Type     string      `json:"type"`
	Name     string      `json:"name"`
	CRS      string      `json:"crs_nota"`
	Features []v1Feature `json:"features"`
}

// WriteGeoJSON escribe un FeatureCollection normalizado.
// crs_nota documenta EPSG:4326; las coordenadas siguen RFC 7946 (lon/lat, sin miembro crs).
func WriteGeoJSON(dir, name string, records []Record) error {
	fc := v1Collection{
		Type:     "FeatureCollection",
		Name:     name,
		CRS:      "EPSG:4326",
		Features: make([]v1Feature, 0, len(records)),
	}
	for _, r := range records {
		geom := r.Geometry
		if len(geom) == 0 {
			geom = json.RawMessage("null")
		}
		props := v1Props{
			FeatureID:   r.FeatureID,
			SourceIndex: r.SourceIndex,
			Codigo:      r.Codigo,
			Nombre:      r.Nombre,
			Uso:         r.Uso,
			ProyRiego:   r.ProyRiego,
			RiegoAct:    r.RiegoAct,
			Referencia:  r.Referencia,
			Clase:       r.Clase,
			Pertenecen:  r.Pertenecen,
			PerimetroM:  r.PerimetroRaw,
			AreaM2:      r.AreaRaw,
		}
		if r.Capa == "jardines_reserva" || r.Capa == "xerofitica" {
			props.Capa = r.Capa
		}
		fc.Features = append(fc.Features, v1Feature{
			Type:       "Feature",
			ID:         r.FeatureID,
			Geometry:   geom,
			Properties: props,
		})
	}

	body, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name+".geojson"), body, 0o644)
}

// Manifest describe los datasets v1 y la omisión de PII.
type Manifest struct {
	CRS        string             `json:"crs"`
	GeneradoEn string             `json:"generado_en"`
	NotaPII    string             `json:"nota_pii"`
	Datasets   map[string]Dataset `json:"datasets"`
}

// Dataset es un archivo normalizado.
type Dataset struct {
	Archivo      string `json:"archivo"`
	Fuente       string `json:"fuente"`
	Features     int    `json:"features"`
	SHA256Fuente string `json:"sha256_fuente"`
}

func WriteManifest(dir string, m Manifest) error {
	body, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), body, 0o644)
}

func (m Manifest) String() string {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", m.Datasets)
	}
	return string(b)
}
