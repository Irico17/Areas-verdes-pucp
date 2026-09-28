package models

import "time"

// SesionModel represents a row in the sesiones table.
type SesionModel struct {
	TokenHash string    `gorm:"column:token_hash;primaryKey"`
	UsuarioID int64     `gorm:"column:usuario_id"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
}

// TableName returns the table name in postgres.
func (SesionModel) TableName() string {
	return "sesiones"
}
