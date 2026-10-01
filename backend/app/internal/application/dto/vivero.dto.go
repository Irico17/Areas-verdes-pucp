// Package dto defines data transfer objects for the application layer.
package dto

// ViveroDTO represents a nursery activity record in JSON responses.
type ViveroDTO struct {
	ID            string `json:"id"`
	Fecha         string `json:"fecha,omitempty"`
	Area          string `json:"area"`
	Subproceso    string `json:"subproceso"`
	Etapa         string `json:"etapa"`
	Descripcion   string `json:"descripcion"`
	Observaciones string `json:"observaciones,omitempty"`
	Responsables  string `json:"responsables,omitempty"`
	LugarID       string `json:"lugar_id,omitempty"`
	LugarLibre    string `json:"lugar_libre,omitempty"`
}

// GuardarViveroDTO represents payload to create or edit a nursery activity record.
type GuardarViveroDTO struct {
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

// ViveroResponseDTO wraps the list of nursery activity records.
type ViveroResponseDTO struct {
	Registros []ViveroDTO `json:"registros"`
}
