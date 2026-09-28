package models

import "time"

// PoligonoCuadrillaModel represents a zone in the zonas view / poligonos_cuadrilla table.
type PoligonoCuadrillaModel struct {
	ID                int64     `gorm:"primaryKey;column:id"`
	FeatureID         string    `gorm:"column:feature_id"`
	SourceIndex       int       `gorm:"column:source_index"`
	Codigo            *string   `gorm:"column:codigo"`
	Nombre            *string   `gorm:"column:nombre"`
	Uso               *string   `gorm:"column:uso"`
	ProyRiego         *string   `gorm:"column:proy_riego"`
	RiegoAct          *string   `gorm:"column:riego_act"`
	Referencia        *string   `gorm:"column:referencia"`
	PerimetroM        *float64  `gorm:"column:perimetro_m"`
	AreaM2            *float64  `gorm:"column:area_m2"`
	CuadrillaID       *string   `gorm:"column:cuadrilla_id"`
	ZonaSupervisionID *int64    `gorm:"column:zona_supervision_id"`
	Activo            bool      `gorm:"column:activo"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}

// TableName returns the table name for PoligonoCuadrillaModel.
func (PoligonoCuadrillaModel) TableName() string {
	return "poligonos_cuadrilla"
}
