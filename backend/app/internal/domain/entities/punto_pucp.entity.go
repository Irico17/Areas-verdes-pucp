// Package entities defines core domain business models.
package entities

// PuntoPUCP represents a public point of interest in the PUCP campus.
type PuntoPUCP struct {
	ID     int64
	Titulo string
	Lat    float64
	Lon    float64
	URL    string
	Activo bool
}
