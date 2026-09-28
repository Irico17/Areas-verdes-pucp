// Package dto defines data transfer objects for application use cases.
package dto

// IndiceInventarioDTO represents the inventory overlay catalogue and currently loaded counts.
type IndiceInventarioDTO struct {
	Capas    []string       `json:"capas"`
	Cargadas []CapaCountDTO `json:"cargadas"`
}

// InventarioPropsDTO defines the GeoJSON feature properties for legacy inventory overlays.
type InventarioPropsDTO struct {
	FeatureID string `json:"feature_id"`
	Capa      string `json:"capa"`
	Nombre    string `json:"nombre,omitempty"`
	Subtipo   string `json:"subtipo,omitempty"`
	Detalle   string `json:"detalle,omitempty"`
	Lugar     string `json:"lugar,omitempty"`
	Foto      string `json:"foto,omitempty"`
}

// CapaDesconocidaErrorDTO represents the 404 response payload for an unknown inventory layer.
type CapaDesconocidaErrorDTO struct {
	Capas []string `json:"capas"`
	Error string   `json:"error"`
}
