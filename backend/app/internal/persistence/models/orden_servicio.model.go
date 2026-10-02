// Package models defines GORM database models mapping to PostgreSQL tables.
package models

import "time"

// OrdenServicioModel maps to public.ordenes_servicio table.
type OrdenServicioModel struct {
	ID               string     `gorm:"column:id;primaryKey"`
	ActividadID      string     `gorm:"column:actividad_id;not null"`
	Empresa          string     `gorm:"column:empresa;not null"`
	EmpresaID        *int64     `gorm:"column:empresa_id"`
	Referencia       string     `gorm:"column:referencia;not null"`
	Frecuencia       string     `gorm:"column:frecuencia;not null;default:''"`
	FrecuenciaID     *int64     `gorm:"column:frecuencia_id"`
	Estado           string     `gorm:"column:estado;not null;default:'en_proceso'"`
	Conformidad      string     `gorm:"column:conformidad;not null;default:''"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null;default:now()"`
	PeriodoInicio    *time.Time `gorm:"column:periodo_inicio"`
	PeriodoFin       *time.Time `gorm:"column:periodo_fin"`
	ReporteProveedor string     `gorm:"column:reporte_proveedor;not null;default:''"`
}

// TableName returns the table name in postgres.
func (OrdenServicioModel) TableName() string {
	return "ordenes_servicio"
}
