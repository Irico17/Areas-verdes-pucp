package dto

// CatalogoItemDTO represents a catalog item in API responses.
type CatalogoItemDTO struct {
	ID          int64  `json:"id"`
	Clase       string `json:"clase"`
	Codigo      string `json:"codigo"`
	Nombre      string `json:"nombre"`
	Activo      bool   `json:"activo"`
	Orden       int    `json:"orden"`
	Provisional bool   `json:"provisional"`
}

// CatalogoListResponseDTO represents the response for listing catalog items.
type CatalogoListResponseDTO struct {
	Items  []CatalogoItemDTO `json:"items"`
	Clases []string          `json:"clases"`
}

// DesactivarCatalogoResponseDTO represents the response after deactivating a catalog item.
type DesactivarCatalogoResponseDTO struct {
	Activo bool  `json:"activo"`
	ID     int64 `json:"id"`
}

// CrearCatalogoDTO represents input for creating or updating a catalog item.
type CrearCatalogoDTO struct {
	Clase  string `json:"clase"`
	Codigo string `json:"codigo"`
	Nombre string `json:"nombre"`
}

// FiltroCatalogoDTO represents parameters for filtering catalog items.
type FiltroCatalogoDTO struct {
	Clase       string
	SoloActivos bool
}

// RenombrarCatalogoDTO corrects the visible name of a catalog item.
type RenombrarCatalogoDTO struct {
	ID        int64
	Nombre    string
	UsuarioID int64
}
