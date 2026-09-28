package models

// ZonaSupervisionModel represents a supervision zone in the zonas_supervision table.
type ZonaSupervisionModel struct {
	ID     int64    `gorm:"primaryKey;column:id"`
	Codigo string   `gorm:"column:codigo"`
	Nombre string   `gorm:"column:nombre"`
	AreaM2 *float64 `gorm:"column:area_m2"`
	Activo bool     `gorm:"column:activo"`
}

// TableName returns the table name for ZonaSupervisionModel.
func (ZonaSupervisionModel) TableName() string {
	return "zonas_supervision"
}
