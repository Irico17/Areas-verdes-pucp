// Package models defines GORM database models.
package models

import "time"

// TachoModel maps to the tachos table in PostgreSQL.
type TachoModel struct {
	ID                  int64     `gorm:"column:id;primaryKey"`
	Codigo              string    `gorm:"column:codigo"`
	Lat                 *float64  `gorm:"column:lat"`
	Lon                 *float64  `gorm:"column:lon"`
	Nota                *string   `gorm:"column:nota"`
	Lugar               *string   `gorm:"column:lugar"`
	Espacios            *string   `gorm:"column:espacios"`
	Accion              *string   `gorm:"column:accion"`
	TachoActual         *string   `gorm:"column:tacho_actual"`
	TachoNuevo          *string   `gorm:"column:tacho_nuevo"`
	Recomendaciones     *string   `gorm:"column:recomendaciones"`
	NoAprovechables     int       `gorm:"column:no_aprovechables"`
	PapelCarton         int       `gorm:"column:papel_carton"`
	Plastico            int       `gorm:"column:plastico"`
	Vidrio              int       `gorm:"column:vidrio"`
	Pilas               int       `gorm:"column:pilas"`
	Peligrosos          int       `gorm:"column:peligrosos"`
	RAEE                int       `gorm:"column:raee"`
	Metales             int       `gorm:"column:metales"`
	Aniquem             int       `gorm:"column:aniquem"`
	IntermediosPlastico int       `gorm:"column:intermedios_plastico"`
	IntermediosMetal    int       `gorm:"column:intermedios_metal"`
	Foto                *string   `gorm:"column:foto"`
	ZonaSupervisionID   *int64    `gorm:"column:zona_supervision_id"`
	OrigenRef           string    `gorm:"column:origen_ref"`
	Activo              bool      `gorm:"column:activo"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at"`
}

// TableName returns the table name for TachoModel.
func (TachoModel) TableName() string {
	return "tachos"
}
