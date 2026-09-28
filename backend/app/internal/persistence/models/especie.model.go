package models

// EspecieModel represents a botanical species in the especies table.
type EspecieModel struct {
	ID               int64   `gorm:"primaryKey;column:id"`
	NombreCientifico string  `gorm:"column:nombre_cientifico"`
	NombreComun      *string `gorm:"column:nombre_comun"`
	Activo           bool    `gorm:"column:activo"`
}

// TableName returns the table name for EspecieModel.
func (EspecieModel) TableName() string {
	return "especies"
}
