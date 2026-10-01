// Package entities contains enterprise domain models.
package entities

// DefinicionPendiente is the status string for indicators without an agreed formula.
const DefinicionPendiente = "definición pendiente"

// FilaReporte represents a single activity row in the basic report.
type FilaReporte struct {
	ID             string
	Titulo         string
	Tipo           string
	Estado         string
	Ejecutor       string
	Equipo         string
	Zona           string
	CodigoExterno  string
	Fuente         string
	CreatedAt      string
	Clase          string
	Lugar          string
	Cuadrilla      string
	FechaSolicitud string
	FechaAtencion  string
}

// ConteoReporte represents activity count aggregated by status.
type ConteoReporte struct {
	Estado string
	N      int
}

// HuecoIndicador represents an indicator without an agreed formula.
type HuecoIndicador struct {
	Clave  string
	Nombre string
	Estado string
	Nota   string
}

// FiltroReporte filters basic report data.
type FiltroReporte struct {
	Estado    string
	Desde     string
	Hasta     string
	Zona      string
	Cuadrilla string
	Origen    string
}

// Reporte represents the complete basic report payload with counts, rows and pending indicators.
type Reporte struct {
	Aviso      string
	PorEstado  []ConteoReporte
	Filas      []FilaReporte
	Pendientes []HuecoIndicador
}

// HuecosIndicador returns indicators that lack an agreed calculation formula.
func HuecosIndicador() []HuecoIndicador {
	return []HuecoIndicador{
		{
			Clave:  "cobertura",
			Nombre: "Cobertura",
			Estado: DefinicionPendiente,
			Nota:   "El registro de riego no produce un porcentaje.",
		},
		{
			Clave:  "rendimiento",
			Nombre: "Rendimiento",
			Estado: DefinicionPendiente,
			Nota:   "No hay horas, personal ni superficie para un reporte de proceso.",
		},
		{
			Clave:  "metricas_proveedor",
			Nombre: "Métricas de proveedor",
			Estado: DefinicionPendiente,
			Nota:   "Jefatura no ha acordado la fórmula. No hay portal del proveedor.",
		},
	}
}
