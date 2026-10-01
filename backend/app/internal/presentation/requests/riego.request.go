// Package requests defines request payload structures for HTTP controllers.
package requests

import "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"

// CrearRiegoRequest represents HTTP body for creating an irrigation record.
type CrearRiegoRequest struct {
	ID         string  `json:"id"`
	Sector     string  `json:"sector"`
	Turno      string  `json:"turno"`
	CapatazID  string  `json:"capataz_id"`
	Fecha      string  `json:"fecha"`
	Nota       string  `json:"nota"`
	ZonaID     string  `json:"zona_supervision_id"`
	Ciclo      string  `json:"ciclo"`
	Superficie float64 `json:"superficie_m2"`
}

// ToDTO converts CrearRiegoRequest to CrearRiegoDTO.
func (r CrearRiegoRequest) ToDTO() dto.CrearRiegoDTO {
	return dto.CrearRiegoDTO{
		ID:         r.ID,
		Sector:     r.Sector,
		Turno:      r.Turno,
		CapatazID:  r.CapatazID,
		Fecha:      r.Fecha,
		Nota:       r.Nota,
		ZonaID:     r.ZonaID,
		Ciclo:      r.Ciclo,
		Superficie: r.Superficie,
	}
}
