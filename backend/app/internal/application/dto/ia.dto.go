// Package dto contains data transfer objects for presentation and application layers.
package dto

// SugerenciaIADTO represents the JSON response for activity type suggestions.
type SugerenciaIADTO struct {
	Codigo         string `json:"codigo"`
	Etiqueta       string `json:"etiqueta"`
	Explicacion    string `json:"explicacion"`
	Confianza      string `json:"confianza"`
	RequiereHumano bool   `json:"requiere_humano"`
}
