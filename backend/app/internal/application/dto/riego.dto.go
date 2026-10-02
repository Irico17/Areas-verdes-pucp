// Package dto defines data transfer objects for the application layer.
package dto

// RiegoDTO represents an irrigation log item returned in JSON responses.
type RiegoDTO struct {
	ID        string `json:"id"`
	Sector    string `json:"sector"`
	Turno     string `json:"turno"`
	CapatazID string `json:"capataz_id,omitempty"`
	Equipo    string `json:"equipo,omitempty"`
	Fecha     string `json:"fecha"`
	Nota      string `json:"nota,omitempty"`
	ZonaID    string `json:"zona_supervision_id,omitempty"`
	Ciclo     string `json:"ciclo,omitempty"`
}

// CrearRiegoDTO represents payload to create an irrigation log.
type CrearRiegoDTO struct {
	ID         string  `json:"id"`
	Sector     string  `json:"sector"`
	SectorID   int64   `json:"sector_id"`
	Turno      string  `json:"turno"`
	CapatazID  string  `json:"capataz_id"`
	Fecha      string  `json:"fecha"`
	Nota       string  `json:"nota"`
	ZonaID     string  `json:"zona_supervision_id"`
	Ciclo      string  `json:"ciclo"`
	Superficie float64 `json:"superficie_m2"`
}

// RiegoResponseDTO wraps list of irrigation logs.
// Provisional is always true: the coverage percentage is not the official formula.
type RiegoResponseDTO struct {
	Aviso       string     `json:"aviso"`
	Provisional bool       `json:"provisional"`
	Cobertura   int        `json:"cobertura"`
	Registros   []RiegoDTO `json:"registros"`
}

// CrearRiegoResponseDTO wraps the response after creating an irrigation log.
type CrearRiegoResponseDTO struct {
	ID string `json:"id"`
}
