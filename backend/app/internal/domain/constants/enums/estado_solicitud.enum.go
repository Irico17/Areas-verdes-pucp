// Package enums defines domain enumeration types.
package enums

// EstadoSolicitud represents the lifecycle status of a service request.
type EstadoSolicitud string

const (
	// EstadoSolicitudPorIniciar means pending/not started.
	EstadoSolicitudPorIniciar EstadoSolicitud = "por_iniciar"
	// EstadoSolicitudEnProceso means in progress.
	EstadoSolicitudEnProceso EstadoSolicitud = "en_proceso"
	// EstadoSolicitudEjecutado means executed.
	EstadoSolicitudEjecutado EstadoSolicitud = "ejecutado"
	// EstadoSolicitudCerrado means closed.
	EstadoSolicitudCerrado EstadoSolicitud = "cerrado"
	// EstadoSolicitudCancelado means cancelled.
	EstadoSolicitudCancelado EstadoSolicitud = "cancelado"
)

// String returns the string representation.
func (e EstadoSolicitud) String() string {
	return string(e)
}
