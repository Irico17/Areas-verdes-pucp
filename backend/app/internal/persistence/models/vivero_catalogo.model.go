// Package models defines GORM database models mapping to PostgreSQL tables.
package models

// ViveroCatalogoModel maps to public.vivero_catalogo table.
type ViveroCatalogoModel struct {
	ID     int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Clase  string `gorm:"column:clase;not null"`
	Nombre string `gorm:"column:nombre;not null"`
	Activo bool   `gorm:"column:activo;not null;default:true"`
}

// TableName returns the table name in postgres.
func (ViveroCatalogoModel) TableName() string {
	return "vivero_catalogo"
}
