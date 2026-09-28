package models

import "time"

// EjemplarModel represents an individual plant specimen in the ejemplares table.
type EjemplarModel struct {
	ID                 int64     `gorm:"primaryKey;column:id"`
	NumeroOrigen       *int      `gorm:"column:numero_origen"`
	Codigo             *string   `gorm:"column:codigo"`
	EspecieID          *int64    `gorm:"column:especie_id"`
	NombreComun        *string   `gorm:"column:nombre_comun"`
	TipoVegetacion     *string   `gorm:"column:tipo_vegetacion"`
	Cantidad           int       `gorm:"column:cantidad"`
	UbicacionLugarID   *int64    `gorm:"column:ubicacion_lugar_id"`
	Referencia         *string   `gorm:"column:referencia"`
	Lat                *float64  `gorm:"column:lat"`
	Lon                *float64  `gorm:"column:lon"`
	ObservacionFen2026 *string   `gorm:"column:observacion_fen_2026"`
	Salud              *string   `gorm:"column:salud"`
	Activo             bool      `gorm:"column:activo"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at"`
}

// TableName returns the table name for EjemplarModel.
func (EjemplarModel) TableName() string {
	return "ejemplares"
}
