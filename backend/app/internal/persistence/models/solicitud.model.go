// Package models defines GORM database models mapping to PostgreSQL tables.
package models

import "time"

// SolicitudModel maps to public.solicitudes table.
type SolicitudModel struct {
	ID                 string     `gorm:"column:id;primaryKey"`
	CodigoExterno      *string    `gorm:"column:codigo_externo"`
	Fuente             string     `gorm:"column:fuente;not null"`
	Titulo             string     `gorm:"column:titulo;not null"`
	Detalle            string     `gorm:"column:detalle;not null;default:''"`
	Prioridad          string     `gorm:"column:prioridad;not null;default:'media'"`
	Estado             string     `gorm:"column:estado;not null;default:'por_iniciar'"`
	Lugar              *string    `gorm:"column:lugar"`
	LugarID            *int64     `gorm:"column:lugar_id"`
	Lat                *float64   `gorm:"column:lat"`
	Lon                *float64   `gorm:"column:lon"`
	Cantidad           *int       `gorm:"column:cantidad"`
	CantidadSolicitada *int       `gorm:"column:cantidad_solicitada"`
	CantidadEjecutada  *int       `gorm:"column:cantidad_ejecutada"`
	ActividadID        *string    `gorm:"column:actividad_id"`
	CreatedAt          time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;not null;default:now()"`
	OrigenRef          *string    `gorm:"column:origen_ref"`
	ArchivadaEn        *time.Time `gorm:"column:archivada_en"`
}

// TableName returns the table name in postgres.
func (SolicitudModel) TableName() string {
	return "solicitudes"
}
