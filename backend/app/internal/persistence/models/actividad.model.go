// Package models defines GORM database models mapping to PostgreSQL tables.
package models

import "time"

// ActividadModel maps to the actividades table in PostgreSQL.
type ActividadModel struct {
	ID                string     `gorm:"primaryKey;type:uuid;column:id"`
	Tipo              string     `gorm:"column:tipo"`
	Estado            string     `gorm:"column:estado"`
	Titulo            string     `gorm:"column:titulo"`
	Detalle           string     `gorm:"column:detalle"`
	AreaFeatureID     *string    `gorm:"column:area_feature_id"`
	ZonaFeatureID     *string    `gorm:"column:zona_feature_id"`
	AssignedCapatazID *string    `gorm:"column:assigned_capataz_id"`
	ArchivadaEn       *time.Time `gorm:"column:archivada_en"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
	Ejecutor          string     `gorm:"column:ejecutor"`
	MotivoArchivo     *string    `gorm:"column:motivo_archivo"`
	LugarID           *int64     `gorm:"column:lugar_id"`
	ZonaSupervisionID *int64     `gorm:"column:zona_supervision_id"`
	FechaSolicitud    *string    `gorm:"column:fecha_solicitud"`
	FechaAtencion     *string    `gorm:"column:fecha_atencion"`
	CuadrillaID       *string    `gorm:"column:cuadrilla_id"`
	ClaseCodigo       *string    `gorm:"column:clase_codigo"`
	TipoCodigo        *string    `gorm:"column:tipo_codigo"`
	Comentario        string     `gorm:"column:comentario"`
	LugarLibre        string     `gorm:"column:lugar_libre"`
	OrigenRef         *string    `gorm:"column:origen_ref"`
	Origen            string     `gorm:"column:origen"`
}

// TableName returns the table name in the database.
func (ActividadModel) TableName() string {
	return "actividades"
}
