// Package dto contains data transfer objects.
package dto

// EstadoSaludDTO represents health check results.
type EstadoSaludDTO struct {
	Database string `json:"database,omitempty"`
	PostGIS  string `json:"postgis,omitempty"`
}
