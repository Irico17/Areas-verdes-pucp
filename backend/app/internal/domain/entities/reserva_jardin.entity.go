// Package entities defines core domain business models.
package entities

// ReservaJardin represents a garden reservation event.
type ReservaJardin struct {
	ID         int64
	JardinID   *int64
	Fecha      string
	HoraInicio string
	HoraFin    string
	Estado     string
	Evento     string
	Unidad     string
	Origen     string
	Activo     bool
}
