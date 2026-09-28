// Package enums defines domain enumerations and constant values.
package enums

// SectorSlugs are the known operational sectors.
const (
	SectorCuadrillaValeria = "cua-valeria"
	SectorCuadrillaMateo   = "cua-mateo"
	SectorCuadrillaRenato  = "cua-renato"
	SectorCampoDeportivo   = "campo-deportivo"
	SectorBosqueHumedo     = "bosque-humedo"
)

var etiquetasSector = map[string]string{
	SectorCuadrillaValeria: "Cuadrilla Valeria Quispe",
	SectorCuadrillaMateo:   "Cuadrilla Mateo Salazar",
	SectorCuadrillaRenato:  "Cuadrilla Renato Cárdenas",
	SectorCampoDeportivo:   "Campo deportivo",
	SectorBosqueHumedo:     "Bosque húmedo",
}

// EtiquetaSector returns the human-readable label for an operational sector slug.
// An unknown or empty slug returns an empty string.
func EtiquetaSector(slug string) string {
	return etiquetasSector[slug]
}
