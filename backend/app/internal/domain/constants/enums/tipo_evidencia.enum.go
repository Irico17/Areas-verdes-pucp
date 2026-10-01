// Package enums defines domain enumeration types and their string values.
package enums

// TipoEvidencia represents the category of an evidence attachment.
type TipoEvidencia string

const (
	// TipoEvidenciaFoto indicates a photographic evidence item.
	TipoEvidenciaFoto TipoEvidencia = "foto"

	// TipoEvidenciaDocumento indicates a document evidence item.
	TipoEvidenciaDocumento TipoEvidencia = "documento"
)

// String returns the string representation of TipoEvidencia.
func (t TipoEvidencia) String() string {
	return string(t)
}
