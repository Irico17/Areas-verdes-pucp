// Package models defines GORM database models.
package models

import "time"

// PuntoPUCPModel maps to the puntos_pucp table in PostgreSQL.
type PuntoPUCPModel struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Titulo    string    `gorm:"column:titulo"`
	Lat       float64   `gorm:"column:lat"`
	Lon       float64   `gorm:"column:lon"`
	URL       *string   `gorm:"column:url"`
	OrigenRef string    `gorm:"column:origen_ref"`
	Activo    bool      `gorm:"column:activo"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName returns the table name for PuntoPUCPModel.
func (PuntoPUCPModel) TableName() string {
	return "puntos_pucp"
}
