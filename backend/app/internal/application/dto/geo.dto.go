package dto

// FiltroGeoDTO bounds a geo query. Nil BBox and zero Limit return the full collection.
type FiltroGeoDTO struct {
	BBox  *BBoxDTO
	Limit int
}

// ResumenDTO contains counts of cadastral data.
type ResumenDTO struct {
	CRS               string         `json:"crs"`
	Areas             int64          `json:"areas"`
	AreasConGeometria int64          `json:"areas_con_geometria"`
	Zonas             int64          `json:"zonas"`
	ZonasConGeometria int64          `json:"zonas_con_geometria"`
	Capas             []CapaCountDTO `json:"capas"`
}

// CapaCountDTO represents the feature count of an auxiliary layer.
type CapaCountDTO struct {
	Capa     string `json:"capa"`
	Features int64  `json:"features"`
}

// CapasIndexDTO lists known and loaded layers.
type CapasIndexDTO struct {
	CapasConocidas []string       `json:"capas_conocidas"`
	Cargadas       []CapaCountDTO `json:"cargadas"`
}
