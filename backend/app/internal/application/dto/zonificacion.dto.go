package dto

// SectorCapatazDTO is one row of the capataz sector catalog.
type SectorCapatazDTO struct {
	ID     int64  `json:"id"`
	Codigo string `json:"codigo"`
	Nombre string `json:"nombre"`
	Color  string `json:"color"`
	Activo bool   `json:"activo"`
}

// SectorListDTO lists capataz sectors.
type SectorListDTO struct {
	Sectores []SectorCapatazDTO `json:"sectores"`
}

// CrearSectorDTO is the payload for a new sector.
type CrearSectorDTO struct {
	Codigo    string `json:"codigo"`
	Nombre    string `json:"nombre"`
	Color     string `json:"color"`
	UsuarioID int64  `json:"-"`
}

// ActualizarSectorDTO corrects name, color or the active flag.
type ActualizarSectorDTO struct {
	Codigo    string `json:"codigo"`
	Nombre    string `json:"nombre"`
	Color     string `json:"color"`
	Activo    *bool  `json:"activo,omitempty"`
	UsuarioID int64  `json:"-"`
}

// ImportacionSectorDTO reports an import that does not duplicate codes.
type ImportacionSectorDTO struct {
	Creados      int `json:"creados"`
	Actualizados int `json:"actualizados"`
}

// LugarCatalogoDTO is a place already stored. It is never created from free text.
type LugarCatalogoDTO struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
}

// ResolverLugarDTO asks to resolve a catalog id or a name. LugarLibre is read-only.
type ResolverLugarDTO struct {
	LugarID    *int64 `json:"lugar_id"`
	Nombre     string `json:"nombre"`
	LugarLibre string `json:"lugar_libre"`
}

// LugarResueltoDTO is the catalog match, plus any historical free text left unread as a new row.
type LugarResueltoDTO struct {
	LugarID    *int64 `json:"lugar_id,omitempty"`
	Nombre     string `json:"nombre,omitempty"`
	LugarLibre string `json:"lugar_libre,omitempty"`
}

// ImportacionViaDTO reports how many road features were written.
type ImportacionViaDTO struct {
	Creadas      int `json:"creadas"`
	Actualizadas int `json:"actualizadas"`
}

// EdificioRefDTO is a building chosen by id.
type EdificioRefDTO struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
}

// CrearReferenteDTO pairs a catalog place with a building id.
type CrearReferenteDTO struct {
	LugarID    int64  `json:"lugar_id"`
	EdificioID string `json:"edificio_id"`
	UsuarioID  int64  `json:"-"`
}

// ReferenteDTO is the stored pair.
type ReferenteDTO struct {
	ID         int64  `json:"id"`
	LugarID    int64  `json:"lugar_id"`
	EdificioID string `json:"edificio_id"`
	Creado     bool   `json:"creado"`
}

// ViaAltaDTO is one line ready to persist.
type ViaAltaDTO struct {
	FeatureID string
	Nombre    string
	GeoJSON   string
}
