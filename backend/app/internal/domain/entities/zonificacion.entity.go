package entities

// SectorCapataz is one row of the capataz sector catalog.
type SectorCapataz struct {
	ID     int64  `json:"id"`
	Codigo string `json:"codigo"`
	Nombre string `json:"nombre"`
	Color  string `json:"color"`
	Activo bool   `json:"activo"`
}

// NuevoSectorCapataz is the payload for a new sector.
type NuevoSectorCapataz struct {
	Codigo    string `json:"codigo"`
	Nombre    string `json:"nombre"`
	Color     string `json:"color"`
	UsuarioID int64  `json:"-"`
}

// CambioSectorCapataz corrects name, color or the active flag.
type CambioSectorCapataz struct {
	Codigo    string `json:"codigo"`
	Nombre    string `json:"nombre"`
	Color     string `json:"color"`
	Activo    *bool  `json:"activo,omitempty"`
	UsuarioID int64  `json:"-"`
}

// ResumenImportacionSector reports an import that does not duplicate codes.
type ResumenImportacionSector struct {
	Creados      int `json:"creados"`
	Actualizados int `json:"actualizados"`
}

// LugarCatalogo is a place already stored. It is never created from free text.
type LugarCatalogo struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
}

// ResumenImportacionVia reports how many road features were written.
type ResumenImportacionVia struct {
	Creadas      int `json:"creadas"`
	Actualizadas int `json:"actualizadas"`
}

// NuevoReferenteEdificio pairs a catalog place with a building id.
type NuevoReferenteEdificio struct {
	LugarID    int64  `json:"lugar_id"`
	EdificioID string `json:"edificio_id"`
	UsuarioID  int64  `json:"-"`
}

// ReferenteEdificio is the stored pair.
type ReferenteEdificio struct {
	ID         int64  `json:"id"`
	LugarID    int64  `json:"lugar_id"`
	EdificioID string `json:"edificio_id"`
	Creado     bool   `json:"creado"`
}

// ViaAlta is one line ready to persist.
type ViaAlta struct {
	FeatureID string
	Nombre    string
	GeoJSON   string
}
