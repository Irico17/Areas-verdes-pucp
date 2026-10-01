// Package models defines GORM database models mapping to PostgreSQL tables.
package models

import "time"

// RiegoRegistroModel maps to public.riego_registros table.
type RiegoRegistroModel struct {
	ID                string    `gorm:"column:id;primaryKey"`
	Sector            string    `gorm:"column:sector;not null"`
	Turno             string    `gorm:"column:turno;not null"`
	CapatazID         *string   `gorm:"column:capataz_id"`
	Fecha             time.Time `gorm:"column:fecha;not null"`
	Nota              string    `gorm:"column:nota;not null;default:''"`
	CreatedAt         time.Time `gorm:"column:created_at;not null;default:now()"`
	ZonaSupervisionID *int64    `gorm:"column:zona_supervision_id"`
	Ciclo             string    `gorm:"column:ciclo;not null;default:''"`
	SuperficieM2      *float64  `gorm:"column:superficie_m2"`
}

// TableName returns the table name in postgres.
func (RiegoRegistroModel) TableName() string {
	return "riego_registros"
}
