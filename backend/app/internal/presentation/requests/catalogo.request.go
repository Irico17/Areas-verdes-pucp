package requests

// CrearCatalogoRequest defines the JSON payload for creating or updating a catalog item.
type CrearCatalogoRequest struct {
	Clase  string `json:"clase"`
	Codigo string `json:"codigo"`
	Nombre string `json:"nombre"`
}
