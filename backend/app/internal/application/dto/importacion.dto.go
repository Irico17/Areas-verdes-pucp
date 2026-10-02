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

// VistaPreviaResponseDTO represents the HTTP response for POST /importaciones.
type VistaPreviaResponseDTO struct {
	ID               int64            `json:"id"`
	LoteID           int64            `json:"lote_id"`
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

// ConfirmarImportacionResponseDTO represents the response for POST /importaciones/:id/confirmar.
type ConfirmarImportacionResponseDTO struct {
	LoteID  int64 `json:"lote_id"`
	Validas int   `json:"validas"`
	Escrito bool  `json:"escrito"`
}

// EntidadesImportablesResponseDTO represents the response for GET /importaciones/entidades.
type EntidadesImportablesResponseDTO struct {
	Entidades []string `json:"entidades"`
}

// OpcionesETLDTO defines options for running the initial ETL.
type OpcionesETLDTO struct {
	RawDir   string
	V1Dir    string
	SkipLoad bool
	Strict   bool
}

// ReporteETLDTO represents the output summary of an ETL run.
type ReporteETLDTO struct {
	Areas      int            `json:"areas"`
	Zonas      int            `json:"zonas"`
	Capas      map[string]int `json:"capas"`
	Inventario map[string]int `json:"inventario"`
}

// OpcionesCargaLoteDTO defines options for running etl-lote.
type OpcionesCargaLoteDTO struct {
	SoloLectura bool
	RawDir      string
}

// ReporteLoteDTO represents the complete summary of a batch load.
type ReporteLoteDTO struct {
	LoteID     int64             `json:"lote_id"`
	Origen     map[string]string `json:"origen"`
	Cargados   map[string]int    `json:"cargados"`
	Rechazados []RechazoDTO      `json:"rechazados"`
	Avisos     []string          `json:"avisos,omitempty"`
}

// OpcionesSectoresDTO defines options for the cmd/sectores tool.
type OpcionesSectoresDTO struct {
	RawDir      string
	V1Dir       string
	Salida      string
	ImprimirSQL bool
	Vivo        bool
}

// ResultadoSectoresDTO represents the result of the cmd/sectores execution.
type ResultadoSectoresDTO struct {
	Desde   string         `json:"desde"`
	Filas   int            `json:"filas"`
	Conteos map[string]int `json:"conteos"`
	SQL     string         `json:"sql,omitempty"`
}
