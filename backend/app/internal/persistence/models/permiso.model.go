package models

// PermisoModel represents a row in the permisos table.
type PermisoModel struct {
	Rol    string `gorm:"column:rol;primaryKey"`
	Accion string `gorm:"column:accion;primaryKey"`
}

// TableName returns the table name in postgres.
func (PermisoModel) TableName() string {
	return "permisos"
}
