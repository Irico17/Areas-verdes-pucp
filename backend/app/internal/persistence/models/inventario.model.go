// Package models contains GORM database schema representations.
package models

// InventarioModel represents a row in the inventario database table.
type InventarioModel struct {
	ID        int64   `gorm:"column:id;primaryKey;autoIncrement"`
	Capa      string  `gorm:"column:capa;not null"`
	FeatureID string  `gorm:"column:feature_id;not null"`
	Nombre    *string `gorm:"column:nombre"`
	Subtipo   *string `gorm:"column:subtipo"`
	Detalle   *string `gorm:"column:detalle"`
	Lugar     *string `gorm:"column:lugar"`
	Foto      *string `gorm:"column:foto"`
}

// TableName returns the physical table name for InventarioModel.
func (InventarioModel) TableName() string {
	return "inventario"
}
