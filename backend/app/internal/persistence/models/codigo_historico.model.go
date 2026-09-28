package models

import "time"

// CodigoHistoricoModel represents historical codes in the codigos_historicos table.
type CodigoHistoricoModel struct {
	ID             int64     `gorm:"primaryKey;column:id"`
	EjemplarID     int64     `gorm:"column:ejemplar_id"`
	CodigoAnterior string    `gorm:"column:codigo_anterior"`
	CodigoNuevo    *string   `gorm:"column:codigo_nuevo"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

// TableName returns the table name for CodigoHistoricoModel.
func (CodigoHistoricoModel) TableName() string {
	return "codigos_historicos"
}
