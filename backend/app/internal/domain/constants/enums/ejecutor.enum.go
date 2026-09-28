// Package enums defines domain enumeration types and constants.
package enums

// Ejecutor represents the execution entity of an activity (internal crew vs third-party contractor).
type Ejecutor string

const (
	// EjecutorPropia represents work executed by internal university personnel.
	EjecutorPropia Ejecutor = "propia"
	// EjecutorTercerizada represents work executed by an outsourced contractor company.
	EjecutorTercerizada Ejecutor = "tercerizada"
)
