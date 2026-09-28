// Package entities defines core domain entities.
package entities

// CapasConocidas defines the list of supported legacy inventory overlay layers.
var CapasConocidas = []string{
	"bebederos",
	"fauna",
	"playas_estacionamiento",
	"puertas",
	"area_vereda_peligro",
	"flora",
	"cafetos",
	"tachos",
	"xerofitica",
	"jardines_reserva",
	"puntos_pucp",
}

// EsCapaConocida checks whether the given layer name belongs to known inventory layers.
func EsCapaConocida(capa string) bool {
	for _, name := range CapasConocidas {
		if name == capa {
			return true
		}
	}
	return false
}
