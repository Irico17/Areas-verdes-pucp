// Package enums defines domain enumerations and constant values.
package enums

import (
	"strings"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// TiposVegetacion contains the allowed values for ejemplares.tipo_vegetacion.
var TiposVegetacion = []string{
	"Árbol", "Palmera", "Arbusto", "Herbácea", "Trepadora", "Suculenta", "cafeto",
}

// ValidarTipoVegetacion accepts empty string or a valid category from TiposVegetacion.
func ValidarTipoVegetacion(tipo string) error {
	tipo = strings.TrimSpace(tipo)
	if tipo == "" {
		return nil
	}
	for _, t := range TiposVegetacion {
		if tipo == t {
			return nil
		}
	}
	return domainErrors.ErrEntrada
}
