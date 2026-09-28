package dto

import (
	"strconv"
	"strings"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// BBoxDTO represents minLon, minLat, maxLon, maxLat in EPSG:4326.
type BBoxDTO struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

// ParseBBox parses a bounding box string formatted as "minLon,minLat,maxLon,maxLat".
func ParseBBox(raw string) (*BBoxDTO, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	if len(parts) != 4 {
		return nil, domainErrors.ErrBBox
	}
	vals := make([]float64, 4)
	for i, p := range parts {
		n, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil, domainErrors.ErrBBox
		}
		vals[i] = n
	}
	if vals[0] >= vals[2] || vals[1] >= vals[3] {
		return nil, domainErrors.ErrBBox
	}
	return &BBoxDTO{MinX: vals[0], MinY: vals[1], MaxX: vals[2], MaxY: vals[3]}, nil
}

// RoundedFloat writes measurements rounded to 3 decimal places as the original cadastral data.
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

// CatastroPropertiesDTO represents non-sensitive cadastral properties shared by areas and zones.
type CatastroPropertiesDTO struct {
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

// CapaPropertiesDTO adds auxiliary layer attributes.
type CapaPropertiesDTO struct {
	CatastroPropertiesDTO
	Capa       string  `json:"capa"`
	Clase      *string `json:"clase,omitempty"`
	Pertenecen *string `json:"pertenecen,omitempty"`
}

// ZonaPropertiesDTO adds operational sector info.
type ZonaPropertiesDTO struct {
	CatastroPropertiesDTO
	Sector         *string `json:"sector,omitempty"`
	SectorEtiqueta *string `json:"sector_etiqueta,omitempty"`
}
