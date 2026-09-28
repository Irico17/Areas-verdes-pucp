package entities

import "strconv"

// BBox represents minLon, minLat, maxLon, maxLat in EPSG:4326.
type BBox struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

// FiltroGeo bounds a geo query in the domain layer.
type FiltroGeo struct {
	BBox  *BBox
	Limit int
}

// CapaResumen represents the feature count of an auxiliary layer.
type CapaResumen struct {
	Capa     string
	Features int64
}

// ResumenCatastro contains counts of cadastral data.
type ResumenCatastro struct {
	CRS               string
	Areas             int64
	AreasConGeometria int64
	Zonas             int64
	ZonasConGeometria int64
	Capas             []CapaResumen
}

// CapasIndex lists known and loaded layers in the domain layer.
type CapasIndex struct {
	CapasConocidas []string
	Cargadas       []CapaResumen
}

// RoundedFloat formats coordinates / measurements with up to 3 decimal places in JSON.
type RoundedFloat float64

// MarshalJSON formats RoundedFloat with 3 decimal places.
func (f RoundedFloat) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(f), 'f', 3, 64)), nil
}

// FloatPtr converts a *float64 to *RoundedFloat.
func FloatPtr(v *float64) *RoundedFloat {
	if v == nil {
		return nil
	}
	r := RoundedFloat(*v)
	return &r
}

// CatastroProperties represents non-sensitive cadastral properties shared by areas and zones.
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

// CapaProperties adds auxiliary layer attributes.
type CapaProperties struct {
	CatastroProperties
	Capa       string  `json:"capa"`
	Clase      *string `json:"clase,omitempty"`
	Pertenecen *string `json:"pertenecen,omitempty"`
}

// ZonaProperties adds operational sector info.
type ZonaProperties struct {
	CatastroProperties
	Sector         *string `json:"sector,omitempty"`
	SectorEtiqueta *string `json:"sector_etiqueta,omitempty"`
}
