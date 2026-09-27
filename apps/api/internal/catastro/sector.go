package catastro

// etiquetasSector solo sirven para mostrar; no se guardan en la BD. Los
// nombres son ficticios (ver apps/api/internal/etl/sector.go y
// docs/MAPA-DATOS-Y-EDICION.md §2), no identifican a ninguna persona real.
var etiquetasSector = map[string]string{
	"cua-valeria":     "Cuadrilla Valeria Quispe",
	"cua-mateo":       "Cuadrilla Mateo Salazar",
	"cua-renato":      "Cuadrilla Renato Cárdenas",
	"campo-deportivo": "Campo deportivo",
	"bosque-humedo":   "Bosque húmedo",
}

// EtiquetaSector devuelve el rótulo legible de un slug de sector. Un slug
// desconocido o vacío devuelve "".
func EtiquetaSector(slug string) string {
	return etiquetasSector[slug]
}
