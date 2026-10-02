// Package models defines GORM database models mapping to PostgreSQL tables.
package models

import "time"

// ActividadEventoModel maps to the actividad_eventos table in PostgreSQL.
type ActividadEventoModel struct {
	ID              int64     `gorm:"primaryKey;autoIncrement;column:id"`
	ActividadID     string    `gorm:"type:uuid;column:actividad_id"`
	Tipo            string    `gorm:"column:tipo"`
	Estado          *string   `gorm:"column:estado"`
	CapatazID       *string   `gorm:"column:capataz_id"`
	CapatazAnterior *string   `gorm:"column:capataz_anterior"`
	ActorRol        string    `gorm:"column:actor_rol"`
	Nota            string    `gorm:"column:nota"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UsuarioID       *int64    `gorm:"column:usuario_id"`
	UUIDCliente     *string   `gorm:"type:uuid;column:uuid_cliente"`
}

// TableName returns the table name in the database.
func (ActividadEventoModel) TableName() string {
	return "actividad_eventos"
}
