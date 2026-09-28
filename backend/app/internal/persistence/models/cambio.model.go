package models

import "time"

// CambioModel represents a record in the cambios audit table.
type CambioModel struct {
	ID        int64     `gorm:"primaryKey;column:id"`
	Entidad   string    `gorm:"column:entidad"`
	EntidadID string    `gorm:"column:entidad_id"`
	Accion    string    `gorm:"column:accion"`
	Antes     *string   `gorm:"column:antes"`
	Despues   *string   `gorm:"column:despues"`
	UsuarioID *int64    `gorm:"column:usuario_id"`
	LoteID    *int64    `gorm:"column:lote_id"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// TableName returns the table name for CambioModel.
func (CambioModel) TableName() string {
	return "cambios"
}
