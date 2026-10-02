package dto

// FichaDTO represents an area's metadata for display and editing.
type FichaDTO struct {
	FeatureID  string   `json:"feature_id"`
	Nombre     string   `json:"nombre"`
	Uso        string   `json:"uso"`
	RiegoAct   string   `json:"riego_act"`
	Referencia string   `json:"referencia"`
	AreaM2     *float64 `json:"area_m2,omitempty"`
	ConGeom    bool     `json:"con_geometria"`
	Activo     bool     `json:"activo"`
}

// ActualizarFichaDTO contains editable fields for a ficha.
type ActualizarFichaDTO struct {
	Nombre     string `json:"nombre"`
	Uso        string `json:"uso"`
	RiegoAct   string `json:"riego_act"`
	Referencia string `json:"referencia"`
}

// CrearAreaSinGeomDTO contains fields for creating an area without GPS.
type CrearAreaSinGeomDTO struct {
	FeatureID string `json:"feature_id"`
	Nombre    string `json:"nombre"`
	Uso       string `json:"uso"`
}
