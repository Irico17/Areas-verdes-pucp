package entities

import (
	"strings"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// Especie classifies botanical specimens by scientific and common names.
type Especie struct {
	ID               int64  `json:"id"`
	NombreCientifico string `json:"nombre_cientifico"`
	NombreComun      string `json:"nombre_comun"`
	Activo           bool   `json:"activo"`
}

// Validar checks if the species fields are valid for creation.
func (e *Especie) Validar() error {
	e.NombreCientifico = strings.TrimSpace(e.NombreCientifico)
	e.NombreComun = strings.TrimSpace(e.NombreComun)
	if e.NombreCientifico == "" {
		return domainErrors.ErrEntrada
	}
	return nil
}
