package models

// CuadrillaModel represents a work team in the cuadrillas table.
type CuadrillaModel struct {
	ID             string `gorm:"primaryKey;column:id"`
	NombreFicticio string `gorm:"column:nombre_ficticio"`
	Turno          string `gorm:"column:turno"`
	Activo         bool   `gorm:"column:activo"`
}

// TableName returns the table name for CuadrillaModel.
func (CuadrillaModel) TableName() string {
	return "cuadrillas"
}
