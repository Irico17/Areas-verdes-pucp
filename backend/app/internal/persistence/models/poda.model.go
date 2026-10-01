// Package models defines GORM database models mapping to PostgreSQL tables.
package models

import "time"

// PodaModel maps to public.podas table.
type PodaModel struct {
	ID                string     `gorm:"column:id;primaryKey"`
	Codigo            string     `gorm:"column:codigo;not null;unique"`
	CodigoExterno     *string    `gorm:"column:codigo_externo"`
	Tipo              string     `gorm:"column:tipo;not null;default:''"`
	TipoActividad     string     `gorm:"column:tipo_actividad;not null;default:''"`
	FechaReporte      *time.Time `gorm:"column:fecha_reporte"`
	FechaEjecucion    *time.Time `gorm:"column:fecha_ejecucion"`
	PersonalFicticio  string     `gorm:"column:personal_ficticio;not null;default:''"`
	Ubicacion         string     `gorm:"column:ubicacion;not null;default:''"`
	LugarID           *string    `gorm:"column:lugar_id"`
	Unidad            string     `gorm:"column:unidad;not null;default:''"`
	CantidadPedida    float64    `gorm:"column:cantidad_pedida;not null;default:0"`
	CantidadEjecutada float64    `gorm:"column:cantidad_ejecutada;not null;default:0"`
	TipoVegetacion    string     `gorm:"column:tipo_vegetacion;not null;default:''"`
	NombreComun       string     `gorm:"column:nombre_comun;not null;default:''"`
	NombreCientifico  string     `gorm:"column:nombre_cientifico;not null;default:''"`
	EspecieID         *string    `gorm:"column:especie_id"`
	Prioridad         string     `gorm:"column:prioridad;not null;default:'media'"`
	Comentario        string     `gorm:"column:comentario;not null;default:''"`
	SolicitudID       *string    `gorm:"column:solicitud_id"`
	ActividadID       *string    `gorm:"column:actividad_id"`
	OrigenRef         *string    `gorm:"column:origen_ref"`
	ArchivadaEn       *time.Time `gorm:"column:archivada_en"`
	CreatedAt         time.Time  `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;not null;default:now()"`
}

// TableName returns the table name in postgres.
func (PodaModel) TableName() string {
	return "podas"
}
