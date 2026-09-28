// Package models defines GORM persistence structs pointing to database tables.
package models

// UsuarioModel represents a row in the usuarios table.
type UsuarioModel struct {
	ID           int64   `gorm:"column:id;primaryKey;autoIncrement"`
	Usuario      string  `gorm:"column:usuario"`
	Nombre       string  `gorm:"column:nombre"`
	Rol          string  `gorm:"column:rol"`
	RolNombre    *string `gorm:"column:rol_nombre;->"`
	CapatazID    *string `gorm:"column:capataz_id"`
	PasswordHash string  `gorm:"column:password_hash"`
	Activo       bool    `gorm:"column:activo"`
}

// TableName returns the table name in postgres.
func (UsuarioModel) TableName() string {
	return "usuarios"
}
