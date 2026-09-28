package entities

// IndiceInventario represents the inventory overlay catalogue and currently loaded counts.
type IndiceInventario struct {
	Capas    []string
	Cargadas []CapaResumen
}

// InventarioProperties defines GeoJSON feature properties for legacy inventory overlays.
type InventarioProperties struct {
	FeatureID string  `json:"feature_id"`
	Capa      string  `json:"capa"`
	Nombre    *string `json:"nombre,omitempty"`
	Subtipo   *string `json:"subtipo,omitempty"`
	Detalle   *string `json:"detalle,omitempty"`
	Lugar     *string `json:"lugar,omitempty"`
	Foto      *string `json:"foto,omitempty"`
}
