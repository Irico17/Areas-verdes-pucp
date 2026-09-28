// Package enums defines domain enumeration types and constants.
package enums

// EstadoIntervencion represents the lifecycle state of a field intervention.
type EstadoIntervencion string

const (
	// EstadoSinEstado indicates an intervention without an explicit state in original monitoring data.
	EstadoSinEstado EstadoIntervencion = "sin_estado"
	// EstadoPendiente indicates an intervention waiting to start.
	EstadoPendiente EstadoIntervencion = "pendiente"
	// EstadoEnProceso indicates an intervention currently in progress.
	EstadoEnProceso EstadoIntervencion = "en_proceso"
	// EstadoBloqueada indicates an intervention temporarily blocked.
	EstadoBloqueada EstadoIntervencion = "bloqueada"
	// EstadoCerrada indicates an intervention that has been successfully completed.
	EstadoCerrada EstadoIntervencion = "cerrada"
	// EstadoCancelada indicates an intervention that has been cancelled.
	EstadoCancelada EstadoIntervencion = "cancelada"
)

// EstadosIntervencion lists all recognized intervention states.
var EstadosIntervencion = []EstadoIntervencion{
	EstadoSinEstado,
	EstadoPendiente,
	EstadoEnProceso,
	EstadoBloqueada,
	EstadoCerrada,
	EstadoCancelada,
}
