package entities

// CapaFicha represents a record from an auxiliary reference layer (fauna, puertas, playas, etc.).
type CapaFicha struct {
	ID         int64  `json:"id"`
	FeatureID  string `json:"feature_id"`
	Nombre     string `json:"nombre,omitempty"`
	Codigo     string `json:"codigo,omitempty"`
	Nota       string `json:"nota,omitempty"`
	Clase      string `json:"clase,omitempty"`
	Riego      string `json:"riego,omitempty"`
	Pertenecen string `json:"pertenecen,omitempty"`
	Activo     bool   `json:"activo"`
}
