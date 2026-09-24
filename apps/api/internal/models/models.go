package models

import "time"

// AreaVerde es un polígono de catastro (semilla areas_verdes.geojson).
// La columna geom vive en PostGIS; el esquema lo definen las migraciones SQL, no AutoMigrate.
type AreaVerde struct {
	ID          int64     `gorm:"primaryKey"`
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
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (AreaVerde) TableName() string { return "areas_verdes" }

// Zona es una zona de catastro. No incluye nombres de jefes de grupo.
type Zona struct {
	ID          int64     `gorm:"primaryKey"`
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
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Zona) TableName() string { return "zonas" }

// CapaAuxiliar guarda jardines de reserva y áreas xerofíticas.
type CapaAuxiliar struct {
	ID          int64     `gorm:"primaryKey"`
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

func (CapaAuxiliar) TableName() string { return "capas_auxiliares" }
