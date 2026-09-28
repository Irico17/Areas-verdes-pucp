package models

import "time"

// CapaAuxiliarModel stores auxiliary layers such as reserve gardens and xerophytic areas.
type CapaAuxiliarModel struct {
	ID          int64     `gorm:"primaryKey;column:id"`
	Capa        string    `gorm:"column:capa"`
	FeatureID   string    `gorm:"column:feature_id"`
	SourceIndex int       `gorm:"column:source_index"`
	Codigo      *string   `gorm:"column:codigo"`
	Nombre      *string   `gorm:"column:nombre"`
	Uso         *string   `gorm:"column:uso"`
	ProyRiego   *string   `gorm:"column:proy_riego"`
	Clase       *string   `gorm:"column:clase"`
	RiegoAct    *string   `gorm:"column:riego_act"`
	Referencia  *string   `gorm:"column:referencia"`
	Pertenecen  *string   `gorm:"column:pertenecen"`
	PerimetroM  *float64  `gorm:"column:perimetro_m"`
	AreaM2      *float64  `gorm:"column:area_m2"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// TableName returns the table name for CapaAuxiliarModel.
func (CapaAuxiliarModel) TableName() string {
	return "capas_auxiliares"
}
