package models

import "time"

// LoteImportacionModel represents a record in the lotes_importacion table.
type LoteImportacionModel struct {
	ID            int64      `gorm:"primaryKey;column:id"`
	Entidad       string     `gorm:"column:entidad"`
	Estado        string     `gorm:"column:estado"`
	UsuarioID     int64      `gorm:"column:usuario_id"`
	Filas         int        `gorm:"column:filas"`
	Contenido     []byte     `gorm:"column:contenido"`
	NombreArchivo string     `gorm:"column:nombre_archivo"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	RevertidoEn   *time.Time `gorm:"column:revertido_en"`
}

// TableName returns the table name for LoteImportacionModel.
func (LoteImportacionModel) TableName() string {
	return "lotes_importacion"
}
