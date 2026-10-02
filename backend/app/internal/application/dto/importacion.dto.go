// Package dto provides data transfer objects for application layer operations.
package dto

// ErrorFilaDTO represents a rejected row or warning from file import.
type ErrorFilaDTO struct {
	Fila   int    `json:"fila"`
	Campo  string `json:"campo"`
	Motivo string `json:"motivo"`
}

// VistaPreviaDTO represents the preview analysis of an imported file.
type VistaPreviaDTO struct {
	Entidad          string           `json:"entidad"`
	Formato          string           `json:"formato"`
	Validas          int              `json:"validas"`
	Errores          []ErrorFilaDTO   `json:"errores"`
	Filas            []map[string]any `json:"filas"`
	ColumnasOmitidas []string         `json:"columnas_omitidas,omitempty"`
	AvisoOmitidas    string           `json:"aviso_omitidas,omitempty"`
	Avisos           []string         `json:"avisos,omitempty"`
	Escrito          bool             `json:"escrito"`
}
