package geojson

import "encoding/json"

// FeatureCollection es GeoJSON RFC 7946. Las coordenadas van en EPSG:4326 (lon/lat).
type FeatureCollection struct {
	Type     string    `json:"type"`
	Name     string    `json:"name,omitempty"`
	Features []Feature `json:"features"`
}

// Feature usa geometry en crudo para no reescribir coordenadas.
type Feature struct {
	Type       string          `json:"type"`
	ID         string          `json:"id"`
	Geometry   json.RawMessage `json:"geometry"`
	Properties any             `json:"properties"`
}

// Collection vacía serializa features como [] y no como null.
func Collection(name string) FeatureCollection {
	return FeatureCollection{
		Type:     "FeatureCollection",
		Name:     name,
		Features: []Feature{},
	}
}

// CatastroProperties son atributos no sensibles compartidos por áreas y zonas.
type CatastroProperties struct {
	ID          int64         `json:"id"`
	FeatureID   string        `json:"feature_id"`
	SourceIndex int           `json:"source_index"`
	Codigo      *string       `json:"codigo,omitempty"`
	Nombre      *string       `json:"nombre,omitempty"`
	Uso         *string       `json:"uso,omitempty"`
	ProyRiego   *string       `json:"proy_riego,omitempty"`
	RiegoAct    *string       `json:"riego_act,omitempty"`
	Referencia  *string       `json:"referencia,omitempty"`
	PerimetroM  *RoundedFloat `json:"perimetro_m,omitempty"`
	AreaM2      *RoundedFloat `json:"area_m2,omitempty"`
}

// CapaProperties añade la capa auxiliar y atributos que no son de persona.
type CapaProperties struct {
	CatastroProperties
	Capa       string  `json:"capa"`
	Clase      *string `json:"clase,omitempty"`
	Pertenecen *string `json:"pertenecen,omitempty"`
}

// RoundedFloat escribe mediciones con 3 decimales, como el catastro de origen.
type RoundedFloat float64

func (f RoundedFloat) MarshalJSON() ([]byte, error) {
	return []byte(format3(float64(f))), nil
}

func FloatPtr(v *float64) *RoundedFloat {
	if v == nil {
		return nil
	}
	r := RoundedFloat(*v)
	return &r
}
