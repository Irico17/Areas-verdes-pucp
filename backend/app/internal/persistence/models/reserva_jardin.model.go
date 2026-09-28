// Package models defines GORM database models.
package models

import "time"

// ReservaJardinModel maps to the reservas_jardin table in PostgreSQL.
type ReservaJardinModel struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	JardinID   *int64    `gorm:"column:jardin_id"`
	Fecha      time.Time `gorm:"column:fecha"`
	HoraInicio time.Time `gorm:"column:hora_inicio"`
	HoraFin    time.Time `gorm:"column:hora_fin"`
	Estado     string    `gorm:"column:estado"`
	Evento     string    `gorm:"column:evento"`
	Unidad     *string   `gorm:"column:unidad"`
	Origen     string    `gorm:"column:origen"`
	OrigenRef  string    `gorm:"column:origen_ref"`
	Activo     bool      `gorm:"column:activo"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

// TableName returns the table name for ReservaJardinModel.
func (ReservaJardinModel) TableName() string {
	return "reservas_jardin"
}
