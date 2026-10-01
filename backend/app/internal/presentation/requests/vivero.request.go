// Package requests defines incoming HTTP payload validation structures.
package requests

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

// GuardarViveroRequest represents payload to create or edit a vivero record.
type GuardarViveroRequest struct {
	ID            string `json:"id"`
	Fecha         string `json:"fecha"`
	Area          string `json:"area"`
	Subproceso    string `json:"subproceso"`
	Etapa         string `json:"etapa"`
	Descripcion   string `json:"descripcion"`
	Observaciones string `json:"observaciones"`
	Responsables  string `json:"responsables"`
	LugarID       string `json:"lugar_id"`
	LugarLibre    string `json:"lugar_libre"`
}

// ToDTO converts request to application DTO.
func (r *GuardarViveroRequest) ToDTO() dto.GuardarViveroDTO {
	return dto.GuardarViveroDTO{
		ID:            r.ID,
		Fecha:         r.Fecha,
		Area:          r.Area,
		Subproceso:    r.Subproceso,
		Etapa:         r.Etapa,
		Descripcion:   r.Descripcion,
		Observaciones: r.Observaciones,
		Responsables:  r.Responsables,
		LugarID:       r.LugarID,
		LugarLibre:    r.LugarLibre,
	}
}

// ToDTOWithID converts request to application DTO with route ID parameter.
func (r *GuardarViveroRequest) ToDTOWithID(id string) dto.GuardarViveroDTO {
	d := r.ToDTO()
	d.ID = id
	return d
}
