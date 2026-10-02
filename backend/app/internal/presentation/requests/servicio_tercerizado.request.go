// Package requests defines request payload structures for HTTP controllers.
package requests

import "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"

// CrearOrdenRequest represents HTTP body for creating a work order.
type CrearOrdenRequest struct {
	ID           string `json:"id"`
	ActividadID  string `json:"actividad_id"`
	Empresa      string `json:"empresa"`
	EmpresaID    int64  `json:"empresa_id"`
	Referencia   string `json:"referencia"`
	Frecuencia   string `json:"frecuencia"`
	FrecuenciaID int64  `json:"frecuencia_id"`
	Conformidad  string `json:"conformidad"`
}

// ToDTO converts CrearOrdenRequest to CrearOrdenDTO.
func (r CrearOrdenRequest) ToDTO() dto.CrearOrdenDTO {
	return dto.CrearOrdenDTO{
		ID:           r.ID,
		ActividadID:  r.ActividadID,
		Empresa:      r.Empresa,
		EmpresaID:    r.EmpresaID,
		Referencia:   r.Referencia,
		Frecuencia:   r.Frecuencia,
		FrecuenciaID: r.FrecuenciaID,
		Conformidad:  r.Conformidad,
	}
}

// EditarOrdenRequest represents HTTP body for editing a work order.
type EditarOrdenRequest struct {
	Conformidad      string `json:"conformidad"`
	PeriodoInicio    string `json:"periodo_inicio"`
	PeriodoFin       string `json:"periodo_fin"`
	ReporteProveedor string `json:"reporte_proveedor"`
	Estado           string `json:"estado"`
}

// ToDTO converts EditarOrdenRequest to EditarOrdenDTO.
func (r EditarOrdenRequest) ToDTO(id string) dto.EditarOrdenDTO {
	return dto.EditarOrdenDTO{
		ID:               id,
		Conformidad:      r.Conformidad,
		PeriodoInicio:    r.PeriodoInicio,
		PeriodoFin:       r.PeriodoFin,
		ReporteProveedor: r.ReporteProveedor,
		Estado:           r.Estado,
	}
}
