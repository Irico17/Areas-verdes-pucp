// Package entities defines domain entities for the application.
package entities

import "time"

// Vivero represents a nursery activity record.
type Vivero struct {
	ID            string
	Fecha         *string
	Area          string
	Subproceso    string
	Etapa         string
	Descripcion   string
	Observaciones string
	Responsables  string
	LugarID       *string
	LugarLibre    string
	ArchivadaEn   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
