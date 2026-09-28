// Package entities defines core domain business models.
package entities

// Bebedero represents a drinking water fountain point.
type Bebedero struct {
	ID      int64
	Codigo  string
	Subtipo string
	Estado  string
	Sede    string
	Lat     *float64
	Lon     *float64
	Activo  bool
}
