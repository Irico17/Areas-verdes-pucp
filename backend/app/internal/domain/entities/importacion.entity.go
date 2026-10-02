// Package entities contains business domain model representations.
package entities

import "encoding/json"

// Rechazo represents a row rejected during parsing or normalization.
type Rechazo struct {
	Fuente string `json:"fuente"`
	Fila   int    `json:"fila"`
	Campo  string `json:"campo"`
	Motivo string `json:"motivo"`
}

// ReporteLote represents summary metrics and rejections of a batch load.
type ReporteLote struct {
	LoteID     int64             `json:"lote_id"`
	Origen     map[string]string `json:"origen"`
	Cargados   map[string]int    `json:"cargados"`
	Rechazados []Rechazo         `json:"rechazados"`
	Avisos     []string          `json:"avisos,omitempty"`
}

// ReporteCapas represents summary metrics and warnings for auxiliary layers 2B.
type ReporteCapas struct {
	Cargados         map[string]int `json:"cargados"`
	Rechazados       []Rechazo      `json:"rechazados"`
	Avisos           []string       `json:"avisos,omitempty"`
	ColumnasOmitidas []string       `json:"columnas_omitidas,omitempty"`
}

// PoligonoSector represents a sector assignment for a polygon.
type PoligonoSector struct {
	SourceIndex int    `json:"source_index"`
	FeatureID   string `json:"feature_id"`
	Sector      string `json:"sector"`
}

// ArchivoSectores represents the generated sector assignment file content.
type ArchivoSectores struct {
	Version   int              `json:"version"`
	Fuente    string           `json:"fuente"`
	Clave     string           `json:"clave"`
	Nota      string           `json:"nota"`
	Conteos   map[string]int   `json:"conteos"`
	Poligonos []PoligonoSector `json:"poligonos"`
}

// ReporteETL represents summary metrics of an ETL pipeline run.
type ReporteETL struct {
	Areas      int
	Zonas      int
	Capas      map[string]int
	Inventario map[string]int
	Manifest   ManifestETL
}

// ManifestETL contains metadata describing v1 datasets.
type ManifestETL struct {
	CRS        string
	GeneradoEn string
	NotaPII    string
	Datasets   map[string]DatasetETL
}

// DatasetETL represents metadata of a single dataset.
type DatasetETL struct {
	Archivo      string
	Fuente       string
	Features     int
	SHA256Fuente string
}

// CatastroRecord represents a normalized catastro record.
type CatastroRecord struct {
	FeatureID    string
	SourceIndex  int
	Capa         string
	Codigo       *string
	Nombre       *string
	Uso          *string
	Clase        *string
	ProyRiego    *string
	RiegoAct     *string
	Referencia   *string
	Pertenecen   *string
	PerimetroM   *float64
	AreaM2       *float64
	PerimetroRaw json.RawMessage
	AreaRaw      json.RawMessage
	Geometry     json.RawMessage
}

// InventarioRecord represents an inventory record to be loaded.
type InventarioRecord struct {
	Capa      string
	FeatureID string
	Nombre    string
	Subtipo   string
	Detalle   string
	Lugar     string
	Foto      string
	Geometry  json.RawMessage
}

// FuentesLote contains raw source payloads for batch loading.
type FuentesLote struct {
	Zonas     []byte
	Jefes     []byte
	Lugares   []byte
	Gviz      []byte
	Palmeras  []byte
	Cafetos   []byte
	Tipos     []byte
	Monitoreo []byte
	Poda      []byte
	Vivero    []byte
	Origen    map[string]string
}
