// Package entities defines core domain business models.
package entities

// RechazoFormato represents a row rejected during CSV parsing.
type RechazoFormato struct {
	Fuente string
	Fila   int
	Campo  string
	Motivo string
}

// PuntoCarga represents a parsed point before persistence.
type PuntoCarga struct {
	Titulo string
	Lat    float64
	Lon    float64
	URL    string
	Origen string
}
