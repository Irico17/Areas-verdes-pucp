// Package entities contains enterprise domain models.
package entities

// SugerenciaIA represents a heuristic activity type suggestion from a title.
type SugerenciaIA struct {
	Codigo         string
	Etiqueta       string
	Explicacion    string
	Confianza      string
	RequiereHumano bool
}
