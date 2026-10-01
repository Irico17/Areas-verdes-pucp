// Package enums defines domain enumeration types.
package enums

// PrioridadSolicitud represents the priority of a service request.
type PrioridadSolicitud string

const (
	// PrioridadBaja represents low priority.
	PrioridadBaja PrioridadSolicitud = "baja"
	// PrioridadMedia represents medium priority.
	PrioridadMedia PrioridadSolicitud = "media"
	// PrioridadAlta represents high priority.
	PrioridadAlta PrioridadSolicitud = "alta"
)

// String returns the string representation.
func (p PrioridadSolicitud) String() string {
	return string(p)
}

// EsPrioridadSolicitudValida checks if the priority is valid.
func EsPrioridadSolicitudValida(v string) bool {
	switch PrioridadSolicitud(v) {
	case PrioridadBaja, PrioridadMedia, PrioridadAlta:
		return true
	default:
		return false
	}
}
