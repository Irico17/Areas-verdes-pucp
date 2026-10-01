// Package models defines GORM database models mapping to PostgreSQL tables.
package models

import "time"

// ViveroRegistroModel maps to public.vivero_registros table.
type ViveroRegistroModel struct {
	ID            string     `gorm:"column:id;primaryKey"`
	Fecha         *time.Time `gorm:"column:fecha"`
	Area          string     `gorm:"column:area;not null;default:''"`
	Subproceso    string     `gorm:"column:subproceso;not null;default:''"`
	Etapa         string     `gorm:"column:etapa;not null;default:''"`
	Descripcion   string     `gorm:"column:descripcion;not null;default:''"`
	Observaciones string     `gorm:"column:observaciones;not null;default:''"`
	Responsables  string     `gorm:"column:responsables;not null;default:''"`
	LugarID       *string    `gorm:"column:lugar_id"`
	LugarLibre    string     `gorm:"column:lugar_libre;not null;default:''"`
	OrigenRef     *string    `gorm:"column:origen_ref"`
	ArchivadaEn   *time.Time `gorm:"column:archivada_en"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

// TableName returns the table name in postgres.
func (ViveroRegistroModel) TableName() string {
	return "vivero_registros"
}
