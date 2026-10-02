// Package models defines GORM database models.
package models

import "time"

// AreaVerdeModel represents a cadastral green area polygon (areas_verdes table).
type AreaVerdeModel struct {
	ID          int64     `gorm:"primaryKey;column:id"`
	FeatureID   string    `gorm:"column:feature_id"`
	SourceIndex int       `gorm:"column:source_index"`
	Codigo      *string   `gorm:"column:codigo"`
	Nombre      *string   `gorm:"column:nombre"`
	Uso         *string   `gorm:"column:uso"`
	ProyRiego   *string   `gorm:"column:proy_riego"`
	RiegoAct    *string   `gorm:"column:riego_act"`
	Referencia  *string   `gorm:"column:referencia"`
	PerimetroM  *float64  `gorm:"column:perimetro_m"`
	AreaM2      *float64  `gorm:"column:area_m2"`
	Activo      bool      `gorm:"column:activo"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// TableName returns the table name for AreaVerdeModel.
func (AreaVerdeModel) TableName() string {
	return "areas_verdes"
}
