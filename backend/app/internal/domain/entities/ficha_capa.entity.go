// Package entities defines core domain business models.
package entities

// FichaCapa represents an editable auxiliary layer record with geometry and attributes.
type FichaCapa struct {
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
