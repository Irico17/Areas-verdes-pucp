// Package models defines GORM database models.
package models

// FichaCapaModel represents a row scanned from one of the 6 auxiliary reference layer tables.
type FichaCapaModel struct {
	ID         int64
	FeatureID  string
	Nombre     string
	Codigo     string
	Nota       string
	Clase      string
	Riego      string
	AreaM2     *float64
	PerimetroM *float64
	Pertenecen string
	Uso        string
	GeoJSON    string
	Activo     bool
}
