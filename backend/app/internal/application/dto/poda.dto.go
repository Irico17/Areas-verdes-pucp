// Package dto defines data transfer objects for the application layer.
package dto

// PodaDTO represents a pruning record in JSON responses.
type PodaDTO struct {
	ID                string  `json:"id"`
	Codigo            string  `json:"codigo"`
	CodigoExterno     string  `json:"codigo_externo,omitempty"`
	Tipo              string  `json:"tipo"`
	TipoActividad     string  `json:"tipo_actividad"`
	FechaReporte      string  `json:"fecha_reporte,omitempty"`
	FechaEjecucion    string  `json:"fecha_ejecucion,omitempty"`
	Personal          string  `json:"personal"`
	Ubicacion         string  `json:"ubicacion"`
	Unidad            string  `json:"unidad"`
	CantidadPedida    float64 `json:"cantidad_pedida"`
	CantidadEjecutada float64 `json:"cantidad_ejecutada"`
	Prioridad         string  `json:"prioridad"`
	Comentario        string  `json:"comentario,omitempty"`
	NombreComun       string  `json:"nombre_comun,omitempty"`
	NombreCientifico  string  `json:"nombre_cientifico,omitempty"`
}

// GuardarPodaDTO represents payload to create or edit a pruning record.
type GuardarPodaDTO struct {
	ID                string  `json:"id"`
	Codigo            string  `json:"codigo"`
	CodigoExterno     string  `json:"codigo_externo"`
	Tipo              string  `json:"tipo"`
	TipoActividad     string  `json:"tipo_actividad"`
	FechaReporte      string  `json:"fecha_reporte"`
	FechaEjecucion    string  `json:"fecha_ejecucion"`
	Personal          string  `json:"personal"`
	Ubicacion         string  `json:"ubicacion"`
	Unidad            string  `json:"unidad"`
	CantidadPedida    float64 `json:"cantidad_pedida"`
	CantidadEjecutada float64 `json:"cantidad_ejecutada"`
	Prioridad         string  `json:"prioridad"`
	Comentario        string  `json:"comentario"`
	NombreComun       string  `json:"nombre_comun"`
	NombreCientifico  string  `json:"nombre_cientifico"`
}

// PodasResponseDTO wraps the list of pruning records.
type PodasResponseDTO struct {
	Podas []PodaDTO `json:"podas"`
}
