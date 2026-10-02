package entities

// RolCatalogo is a row of the roles catalog. The code is stable.
type RolCatalogo struct {
	Codigo string
	Nombre string
	Activo bool
	Orden  int
}
