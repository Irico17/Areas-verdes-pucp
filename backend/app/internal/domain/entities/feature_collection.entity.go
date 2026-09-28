// Package entities defines core domain entities.
package entities

import "encoding/json"

// FeatureCollection represents a GeoJSON RFC 7946 FeatureCollection.
// Coordinates are in EPSG:4326 (lon/lat).
type FeatureCollection struct {
	Type     string    `json:"type"`
	Name     string    `json:"name,omitempty"`
	Features []Feature `json:"features"`
}

// Feature represents a single GeoJSON Feature with raw geometry.
type Feature struct {
	Type       string          `json:"type"`
	ID         string          `json:"id"`
	Geometry   json.RawMessage `json:"geometry"`
	Properties any             `json:"properties"`
}

// Collection returns an empty FeatureCollection with non-nil features slice.
func Collection(name string) FeatureCollection {
	return FeatureCollection{
		Type:     "FeatureCollection",
		Name:     name,
		Features: []Feature{},
	}
}
