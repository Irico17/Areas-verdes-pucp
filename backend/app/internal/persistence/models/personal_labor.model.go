// Package models defines GORM database models mapping to PostgreSQL tables.
package models

import "time"

// PersonalLaborModel maps to the personal_labor table in PostgreSQL.
type PersonalLaborModel struct {
	ID             string    `gorm:"primaryKey;type:uuid;column:id"`
	ActividadID    string    `gorm:"type:uuid;column:actividad_id"`
	NombreFicticio string    `gorm:"column:nombre_ficticio"`
	RolCampo       string    `gorm:"column:rol_campo"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

// TableName returns the table name in the database.
func (PersonalLaborModel) TableName() string {
	return "personal_labor"
}
