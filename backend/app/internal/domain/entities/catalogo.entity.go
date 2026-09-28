package entities

// CatalogoItem represents a catalog item entity.
type CatalogoItem struct {
	ID     int64
	Clase  string
	Codigo string
	Nombre string
	Activo bool
	Orden  int
}
