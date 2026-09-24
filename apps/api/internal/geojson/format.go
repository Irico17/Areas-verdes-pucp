package geojson

import (
	"errors"
	"strconv"
	"strings"
)

// ErrBBox indica un query bbox inválido.
var ErrBBox = errors.New("bbox debe ser minLon,minLat,maxLon,maxLat con min < max")

func format3(v float64) string {
	return strconv.FormatFloat(v, 'f', 3, 64)
}

// BBox es minLon,minLat,maxLon,maxLat en EPSG:4326.
type BBox struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

func ParseBBox(raw string) (*BBox, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	if len(parts) != 4 {
		return nil, ErrBBox
	}
	vals := make([]float64, 4)
	for i, p := range parts {
		n, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil, ErrBBox
		}
		vals[i] = n
	}
	if vals[0] >= vals[2] || vals[1] >= vals[3] {
		return nil, ErrBBox
	}
	return &BBox{MinX: vals[0], MinY: vals[1], MaxX: vals[2], MaxY: vals[3]}, nil
}
