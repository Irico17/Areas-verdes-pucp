package entities

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// CodigosZonaSupervision contains valid codes for supervision zones (Z1–Z4).
var CodigosZonaSupervision = []string{"Z1", "Z2", "Z3", "Z4"}

// NormalizarNombre normalizes a place name to lowercase, trimmed and without accents.
func NormalizarNombre(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = norm.NFD.String(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// ValidarCodigoZona accepts only Z1–Z4.
func ValidarCodigoZona(codigo string) error {
	codigo = strings.TrimSpace(codigo)
	for _, c := range CodigosZonaSupervision {
		if codigo == c {
			return nil
		}
	}
	return domainErrors.ErrEntrada
}

// ValidarPuntoCampus verifies that coordinates fall within campus bounds.
func ValidarPuntoCampus(lat, lon float64) error {
	if lat < -12.20 || lat > -11.90 || lon < -77.30 || lon > -76.90 {
		return domainErrors.ErrEntrada
	}
	return nil
}

// ValidarCantidad ensures count is at least 1.
func ValidarCantidad(n int) error {
	if n < 1 {
		return domainErrors.ErrEntrada
	}
	return nil
}

// ValidarReferencia ensures reference text does not exceed 500 characters.
func ValidarReferencia(s string) error {
	if len([]rune(strings.TrimSpace(s))) > 500 {
		return domainErrors.ErrEntrada
	}
	return nil
}
