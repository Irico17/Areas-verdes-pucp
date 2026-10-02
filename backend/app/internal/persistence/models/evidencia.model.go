// Package models defines GORM database models mapping to PostgreSQL tables.
package models

import "time"

// EvidenciaModel maps to the evidencias table in PostgreSQL.
type EvidenciaModel struct {
	ID          string    `gorm:"primaryKey;type:uuid;column:id"`
	ActividadID *string   `gorm:"type:uuid;column:actividad_id"`
	SolicitudID *string   `gorm:"type:uuid;column:solicitud_id"`
	OrdenID     *string   `gorm:"type:uuid;column:orden_id"`
	Nombre      string    `gorm:"column:nombre;not null"`
	Mime        string    `gorm:"column:mime;not null"`
	Bytes       int       `gorm:"column:bytes;not null"`
	Ruta        string    `gorm:"column:ruta;not null"`
	Nota        string    `gorm:"column:nota;not null;default:''"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:now()"`
	SHA256      *string   `gorm:"column:sha256"`
	Lat         *float64  `gorm:"column:lat"`
	Lon         *float64  `gorm:"column:lon"`
	Exif        *string   `gorm:"type:jsonb;column:exif"`
	EventoID    *int64    `gorm:"column:evento_id"`
}

// TableName returns the table name in the database.
func (EvidenciaModel) TableName() string {
	return "evidencias"
}
