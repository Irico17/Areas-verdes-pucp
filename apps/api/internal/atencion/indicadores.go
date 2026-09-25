package atencion

// Hueco es un indicador sin fórmula acordada. No lleva número ni meta.
type Hueco struct {
	Clave  string `json:"clave"`
	Nombre string `json:"nombre"`
	Estado string `json:"estado"`
	Nota   string `json:"nota"`
}

const definicionPendiente = "definición pendiente"

// HuecosIndicador nombra el hueco de HUID 16, 24 y 26. No calcula cobertura,
// rendimiento ni métricas de proveedor.
func HuecosIndicador() []Hueco {
	return []Hueco{
		{
			Clave:  "cobertura",
			Nombre: "Cobertura",
			Estado: definicionPendiente,
			Nota:   "El registro de riego no produce un porcentaje.",
		},
		{
			Clave:  "rendimiento",
			Nombre: "Rendimiento",
			Estado: definicionPendiente,
			Nota:   "No hay horas, personal ni superficie para un reporte de proceso.",
		},
		{
			Clave:  "metricas_proveedor",
			Nombre: "Métricas de proveedor",
			Estado: definicionPendiente,
			Nota:   "Jefatura no ha acordado la fórmula. No hay portal del proveedor.",
		},
	}
}
