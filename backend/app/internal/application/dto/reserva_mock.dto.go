// Package dto defines data transfer objects for application use cases.
package dto

// ReservaItemDTO represents a single mock reservation record.
type ReservaItemDTO struct {
	ID                string `json:"id"`
	Jardin            string `json:"jardin"`
	JardinCodigo      string `json:"jardin_codigo,omitempty"`
	Fecha             string `json:"fecha"`
	Hora              string `json:"hora"`
	Evento            string `json:"evento"`
	UnidadResponsable string `json:"unidad_responsable,omitempty"`
	Estado            string `json:"estado"`
	Notas             string `json:"notas"`
	Fake              bool   `json:"fake"`
}

// ReservasMockResponseDTO represents the mock reservations agenda response.
type ReservasMockResponseDTO struct {
	Fake     bool             `json:"fake"`
	Aviso    string           `json:"aviso"`
	Total    int              `json:"total"`
	Reservas []ReservaItemDTO `json:"reservas"`
}
