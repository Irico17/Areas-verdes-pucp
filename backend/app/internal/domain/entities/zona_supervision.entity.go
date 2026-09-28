package entities

import (
	"strings"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// ZonaSupervision represents one of the four supervision zones Z1–Z4.
type ZonaSupervision struct {
	ID      int64    `json:"id"`
	Codigo  string   `json:"codigo"`
	Nombre  string   `json:"nombre"`
	AreaM2  *float64 `json:"area_m2,omitempty"`
	ConGeom bool     `json:"con_geometria"`
	Activo  bool     `json:"activo"`
}

// Validar checks if the zona supervision fields are valid for creation.
func (z *ZonaSupervision) Validar(geojson string) error {
	z.Codigo = strings.TrimSpace(z.Codigo)
	z.Nombre = strings.TrimSpace(z.Nombre)
	if err := ValidarCodigoZona(z.Codigo); err != nil || z.Nombre == "" || strings.TrimSpace(geojson) == "" {
		return domainErrors.ErrEntrada
	}
	return nil
}
