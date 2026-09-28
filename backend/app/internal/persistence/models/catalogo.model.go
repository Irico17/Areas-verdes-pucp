package models

// CatalogoModel represents a row in the catalogos table.
type CatalogoModel struct {
	ID     int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Clase  string `gorm:"column:clase"`
	Codigo string `gorm:"column:codigo"`
	Nombre string `gorm:"column:nombre"`
	Activo bool   `gorm:"column:activo"`
	Orden  int    `gorm:"column:orden"`
}

// TableName returns the table name in postgres.
func (CatalogoModel) TableName() string {
	return "catalogos"
}
