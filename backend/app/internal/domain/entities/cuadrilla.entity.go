package entities

import (
	"strings"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// Cuadrilla represents an operational work team.
type Cuadrilla struct {
	ID             string `json:"id"`
	NombreFicticio string `json:"nombre_ficticio"`
	Turno          string `json:"turno"`
	Activo         bool   `json:"activo"`
}

// Validar checks if the cuadrilla fields are valid for creation.
func (c *Cuadrilla) Validar() error {
	c.ID = strings.TrimSpace(c.ID)
	c.NombreFicticio = strings.TrimSpace(c.NombreFicticio)
	c.Turno = strings.TrimSpace(c.Turno)
	if c.ID == "" || c.NombreFicticio == "" || (c.Turno != "manana" && c.Turno != "tarde") {
		return domainErrors.ErrEntrada
	}
	return nil
}
