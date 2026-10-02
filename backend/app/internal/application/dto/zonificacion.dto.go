package dto

import "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"

// SectorCapatazDTO is one row of the capataz sector catalog.
type SectorCapatazDTO = entities.SectorCapataz

// SectorListDTO lists capataz sectors.
type SectorListDTO struct {
	Sectores []SectorCapatazDTO `json:"sectores"`
}

// CrearSectorDTO is the payload for a new sector.
type CrearSectorDTO = entities.NuevoSectorCapataz

// ActualizarSectorDTO corrects name, color or the active flag.
type ActualizarSectorDTO = entities.CambioSectorCapataz

// ImportacionSectorDTO reports an import that does not duplicate codes.
type ImportacionSectorDTO = entities.ResumenImportacionSector

// LugarCatalogoDTO is a place already stored. It is never created from free text.
type LugarCatalogoDTO = entities.LugarCatalogo

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
type ImportacionViaDTO = entities.ResumenImportacionVia

// EdificioRefDTO is a building chosen by id.
type EdificioRefDTO struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
}

// CrearReferenteDTO pairs a catalog place with a building id.
type CrearReferenteDTO = entities.NuevoReferenteEdificio

// ReferenteDTO is the stored pair.
type ReferenteDTO = entities.ReferenteEdificio

// ViaAltaDTO is one line ready to persist.
type ViaAltaDTO = entities.ViaAlta
