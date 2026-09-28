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

// ElementoInventario represents an item in the legacy inventory catalog.
type ElementoInventario struct {
	ID        int64  `json:"id"`
	Capa      string `json:"capa"`
	FeatureID string `json:"feature_id"`
	Nombre    string `json:"nombre,omitempty"`
	Subtipo   string `json:"subtipo,omitempty"`
	Detalle   string `json:"detalle,omitempty"`
	Lugar     string `json:"lugar,omitempty"`
	Foto      string `json:"foto,omitempty"`
}

// CapaInventarioResumen summarizes the count of loaded features in an inventory layer.
type CapaInventarioResumen struct {
	Capa     string `json:"capa"`
	Features int64  `json:"features"`
}
