// Package models defines GORM database models.
package models

import "time"

// BebederoModel maps to the bebederos table in PostgreSQL.
type BebederoModel struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Codigo    string    `gorm:"column:codigo"`
	Subtipo   string    `gorm:"column:subtipo"`
	Estado    string    `gorm:"column:estado"`
	Sede      *string   `gorm:"column:sede"`
	Lat       *float64  `gorm:"column:lat"`
	Lon       *float64  `gorm:"column:lon"`
	Foto      *string   `gorm:"column:foto"`
	OrigenRef string    `gorm:"column:origen_ref"`
	Activo    bool      `gorm:"column:activo"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName returns the table name for BebederoModel.
func (BebederoModel) TableName() string {
	return "bebederos"
}
