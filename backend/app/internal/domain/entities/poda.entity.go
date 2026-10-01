// Package entities defines domain entities for the application.
package entities

import "time"

// Poda represents a pruning / tree trimming task record.
type Poda struct {
	ID                string
	Codigo            string
	CodigoExterno     *string
	Tipo              string
	TipoActividad     string
	FechaReporte      *string
	FechaEjecucion    *string
	Personal          string
	Ubicacion         string
	Unidad            string
	CantidadPedida    float64
	CantidadEjecutada float64
	Prioridad         string
	Comentario        string
	NombreComun       string
	NombreCientifico  string
	ArchivadaEn       *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// GuardarPoda contains domain input parameters to create or edit a pruning record.
type GuardarPoda struct {
	ID                string
	Codigo            string
	CodigoExterno     string
	Tipo              string
	TipoActividad     string
	FechaReporte      string
	FechaEjecucion    string
	Personal          string
	Ubicacion         string
	Unidad            string
	CantidadPedida    float64
	CantidadEjecutada float64
	Prioridad         string
	Comentario        string
	NombreComun       string
	NombreCientifico  string
}
