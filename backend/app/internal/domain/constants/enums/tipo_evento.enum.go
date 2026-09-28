// Package enums defines domain enumeration types and constants.
package enums

// TipoEvento represents the kind of event logged in an activity timeline.
type TipoEvento string

const (
	// TipoEventoCreada represents the initial creation event.
	TipoEventoCreada TipoEvento = "creada"
	// TipoEventoAsignada represents the initial assignment to a capataz.
	TipoEventoAsignada TipoEvento = "asignada"
	// TipoEventoReasignada represents reassignment to another capataz.
	TipoEventoReasignada TipoEvento = "reasignada"
	// TipoEventoEstado represents a state transition event.
	TipoEventoEstado TipoEvento = "estado"
	// TipoEventoCancelada represents a cancellation event.
	TipoEventoCancelada TipoEvento = "cancelada"
	// TipoEventoArchivada represents a soft-delete / archival event.
	TipoEventoArchivada TipoEvento = "archivada"
	// TipoEventoEvidencia represents an evidence upload event.
	TipoEventoEvidencia TipoEvento = "evidencia"
)
