// Package entities defines core domain entities.
package entities

import "time"

// EvidenciaDeEvento is a file hanging from one timeline event.
type EvidenciaDeEvento struct {
	ID     string
	Nombre string
	Mime   string
}

// ActividadEvento represents an event log row in the timeline of an activity.
type ActividadEvento struct {
	ID                int64
	ActividadID       string
	Tipo              string
	Estado            *string
	CapatazID         *string
	Equipo            *string
	CapatazAnterior   *string
	CuadrillaAnterior *string
	ActorRol          string
	UsuarioID         *int64
	Usuario           string
	Nombre            string
	Nota              string
	UUIDCliente       *string
	Evidencias        []EvidenciaDeEvento
	CreatedAt         time.Time
}
