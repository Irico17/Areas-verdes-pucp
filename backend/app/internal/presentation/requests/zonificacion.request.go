package requests

// CrearSectorRequest is the body for a new capataz sector.
type CrearSectorRequest struct {
	Codigo string `json:"codigo"`
	Nombre string `json:"nombre"`
	Color  string `json:"color"`
}

// ActualizarSectorRequest corrects the visible name, the map color or the active flag.
type ActualizarSectorRequest struct {
	Nombre string `json:"nombre"`
	Color  string `json:"color"`
	Activo *bool  `json:"activo"`
}

// ImportarSectoresRequest carries catalog rows. A repeated code updates, it does not insert twice.
type ImportarSectoresRequest struct {
	Sectores []CrearSectorRequest `json:"sectores"`
}

// ResolverLugarRequest resolves a catalog place. A free-text name that is not already stored is rejected.
type ResolverLugarRequest struct {
	LugarID    *int64 `json:"lugar_id"`
	Nombre     string `json:"nombre"`
	LugarLibre string `json:"lugar_libre"`
}

// CrearReferenteRequest stores a building id next to a catalog place.
type CrearReferenteRequest struct {
	LugarID    int64  `json:"lugar_id"`
	EdificioID string `json:"edificio_id"`
}
