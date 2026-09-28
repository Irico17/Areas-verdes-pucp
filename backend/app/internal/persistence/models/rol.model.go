package models

// RolModel represents a row in the roles table.
type RolModel struct {
	Codigo string `gorm:"column:codigo;primaryKey"`
	Nombre string `gorm:"column:nombre"`
	Orden  int    `gorm:"column:orden"`
}

// TableName returns the table name in postgres.
func (RolModel) TableName() string {
	return "roles"
}
