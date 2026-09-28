// Package entities defines core domain business models.
package entities

// Tacho represents a waste bin point with 11 distinct counts.
type Tacho struct {
	ID                  int64
	Codigo              string
	Lat                 *float64
	Lon                 *float64
	Nota                string
	Lugar               string
	Espacios            string
	Accion              string
	TachoActual         string
	TachoNuevo          string
	Recomendaciones     string
	NoAprovechables     int
	PapelCarton         int
	Plastico            int
	Vidrio              int
	Pilas               int
	Peligrosos          int
	RAEE                int
	Metales             int
	Aniquem             int
	IntermediosPlastico int
	IntermediosMetal    int
	Activo              bool
}
