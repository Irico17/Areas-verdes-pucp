// Package models defines GORM database models mapping to PostgreSQL tables.
package models

// CapatazModel maps to the capataces table in PostgreSQL.
type CapatazModel struct {
	ID     string `gorm:"primaryKey;column:id"`
	Equipo string `gorm:"column:equipo"`
	Turno  string `gorm:"column:turno"`
	Activo bool   `gorm:"column:activo"`
}

// TableName returns the table name in the database.
func (CapatazModel) TableName() string {
	return "capataces"
}
