package requests

// ActualizarFichaRequest represents the JSON body to update an area ficha.
type ActualizarFichaRequest struct {
	Nombre     string `json:"nombre"`
	Uso        string `json:"uso"`
	RiegoAct   string `json:"riego_act"`
	Referencia string `json:"referencia"`
}

// CrearAreaSinGeomRequest represents the JSON body to create an area without GPS.
type CrearAreaSinGeomRequest struct {
	FeatureID string `json:"feature_id"`
	Nombre    string `json:"nombre"`
	Uso       string `json:"uso"`
}
