// Package entities defines domain entities for the application.
package entities

import "time"

// Evidencia represents an evidence attachment associated with an activity or order.
type Evidencia struct {
	ID          string
	ActividadID string
	SolicitudID *string
	OrdenID     *string
	Nombre      string
	Mime        string
	Bytes       int
	Ruta        string
	Nota        string
	SHA256      *string
	Lat         *float64
	Lon         *float64
	Exif        *string
	CreatedAt   time.Time
}

// GuardarEvidencia represents domain parameters for saving an evidence record.
type GuardarEvidencia struct {
	ID          string
	ActividadID string
	OrdenID     string
	Nombre      string
	Mime        string
	Ext         string
	Bytes       int
	Contenido   []byte
	Ruta        string
	Nota        string
	Hash        string
	Lat         *float64
	Lon         *float64
	Exif        any
	Rol         string
	CapatazID   string
	UsuarioID   int64
}

// ResultadoEvidencia represents the result of saving an evidence record.
type ResultadoEvidencia struct {
	ID          string
	Idempotente bool
}
