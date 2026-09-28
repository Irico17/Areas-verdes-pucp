package entities

import (
	"strings"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// Lugar represents a named campus location with geographic coordinates.
type Lugar struct {
	ID                int64   `json:"id"`
	Nombre            string  `json:"nombre"`
	NombreNorm        string  `json:"nombre_norm"`
	Lat               float64 `json:"lat"`
	Lon               float64 `json:"lon"`
	ZonaSupervisionID *int64  `json:"zona_supervision_id,omitempty"`
	Activo            bool    `json:"activo"`
}

// Validar checks if the place fields are valid for creation.
func (l *Lugar) Validar() error {
	l.Nombre = strings.TrimSpace(l.Nombre)
	l.NombreNorm = NormalizarNombre(l.Nombre)
	if l.Nombre == "" || l.NombreNorm == "" {
		return domainErrors.ErrEntrada
	}
	return ValidarPuntoCampus(l.Lat, l.Lon)
}
