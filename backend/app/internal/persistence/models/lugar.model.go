package models

// LugarModel represents a place in the lugares table.
type LugarModel struct {
	ID                int64   `gorm:"primaryKey;column:id"`
	Nombre            string  `gorm:"column:nombre"`
	NombreNorm        string  `gorm:"column:nombre_norm"`
	Lat               float64 `gorm:"column:lat"`
	Lon               float64 `gorm:"column:lon"`
	ZonaSupervisionID *int64  `gorm:"column:zona_supervision_id"`
	Activo            bool    `gorm:"column:activo"`
}

// TableName returns the table name for LugarModel.
func (LugarModel) TableName() string {
	return "lugares"
}
