// Package requests defines incoming HTTP payload validation structures.
package requests

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

// GuardarPodaRequest represents payload to create or edit a poda record.
type GuardarPodaRequest struct {
	ID                string  `json:"id"`
	Codigo            string  `json:"codigo"`
	CodigoExterno     string  `json:"codigo_externo"`
	Tipo              string  `json:"tipo"`
	TipoActividad     string  `json:"tipo_actividad"`
	FechaReporte      string  `json:"fecha_reporte"`
	FechaEjecucion    string  `json:"fecha_ejecucion"`
	Personal          string  `json:"personal"`
	Ubicacion         string  `json:"ubicacion"`
	Unidad            string  `json:"unidad"`
	CantidadPedida    float64 `json:"cantidad_pedida"`
	CantidadEjecutada float64 `json:"cantidad_ejecutada"`
	Prioridad         string  `json:"prioridad"`
	Comentario        string  `json:"comentario"`
	NombreComun       string  `json:"nombre_comun"`
	NombreCientifico  string  `json:"nombre_cientifico"`
}

// ToDTO converts request to application DTO.
func (r *GuardarPodaRequest) ToDTO() dto.GuardarPodaDTO {
	return dto.GuardarPodaDTO{
		ID:                r.ID,
		Codigo:            r.Codigo,
		CodigoExterno:     r.CodigoExterno,
		Tipo:              r.Tipo,
		TipoActividad:     r.TipoActividad,
		FechaReporte:      r.FechaReporte,
		FechaEjecucion:    r.FechaEjecucion,
		Personal:          r.Personal,
		Ubicacion:         r.Ubicacion,
		Unidad:            r.Unidad,
		CantidadPedida:    r.CantidadPedida,
		CantidadEjecutada: r.CantidadEjecutada,
		Prioridad:         r.Prioridad,
		Comentario:        r.Comentario,
		NombreComun:       r.NombreComun,
		NombreCientifico:  r.NombreCientifico,
	}
}

// ToDTOWithID converts request to application DTO with route ID parameter.
func (r *GuardarPodaRequest) ToDTOWithID(id string) dto.GuardarPodaDTO {
	d := r.ToDTO()
	d.ID = id
	return d
}
