// Package enums defines domain enumeration types.
package enums

// EstadoOrden represents the lifecycle status of a work order.
type EstadoOrden string

const (
	// EstadoOrdenEnProceso means in progress.
	EstadoOrdenEnProceso EstadoOrden = "en_proceso"
	// EstadoOrdenEjecutada means executed.
	EstadoOrdenEjecutada EstadoOrden = "ejecutada"
	// EstadoOrdenConforme means approved / compliant.
	EstadoOrdenConforme EstadoOrden = "conforme"
)

// String returns the string representation.
func (e EstadoOrden) String() string {
	return string(e)
}

// EsEstadoOrdenValido checks if the order state is recognized.
func EsEstadoOrdenValido(v string) bool {
	switch EstadoOrden(v) {
	case EstadoOrdenEnProceso, EstadoOrdenEjecutada, EstadoOrdenConforme:
		return true
	default:
		return false
	}
}
