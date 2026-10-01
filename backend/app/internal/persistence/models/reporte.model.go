// Package models contains GORM and raw SQL database models.
package models

import (
	"database/sql"
	"time"
)

// FilaReporteModel represents the database query row for a basic report.
type FilaReporteModel struct {
	ID             string
	Titulo         string
	Tipo           string
	Estado         string
	Ejecutor       string
	Equipo         string
	Zona           string
	CodigoExterno  string
	Fuente         string
	CreatedAt      time.Time
	Clase          string
	Lugar          string
	Cuadrilla      string
	FechaSolicitud sql.NullTime
	FechaAtencion  sql.NullTime
}

// ConteoReporteModel represents activity counts grouped by status.
type ConteoReporteModel struct {
	Estado string `gorm:"column:estado"`
	N      int    `gorm:"column:n"`
}
