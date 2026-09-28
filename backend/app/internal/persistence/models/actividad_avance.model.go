// Package models defines GORM database models mapping to PostgreSQL tables.
package models

import "time"

// ActividadAvanceModel maps to the actividad_avances table in PostgreSQL.
type ActividadAvanceModel struct {
	ID            string    `gorm:"primaryKey;type:uuid;column:id"`
	ActividadID   string    `gorm:"type:uuid;column:actividad_id"`
	Fecha         string    `gorm:"column:fecha"`
	Nota          string    `gorm:"column:nota"`
	AreaFeatureID *string   `gorm:"column:area_feature_id"`
	EjemplarRef   *string   `gorm:"column:ejemplar_ref"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

// TableName returns the table name in the database.
func (ActividadAvanceModel) TableName() string {
	return "actividad_avances"
}
