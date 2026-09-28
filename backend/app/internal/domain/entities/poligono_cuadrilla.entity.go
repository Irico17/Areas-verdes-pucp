package entities

// PoligonoCuadrilla represents an operational work team sector polygon.
type PoligonoCuadrilla struct {
	ID                int64  `json:"id"`
	FeatureID         string `json:"feature_id"`
	Codigo            string `json:"codigo"`
	Nombre            string `json:"nombre"`
	CuadrillaID       string `json:"cuadrilla_id"`
	ZonaSupervisionID *int64 `json:"zona_supervision_id,omitempty"`
	ConGeom           bool   `json:"con_geometria"`
	Activo            bool   `json:"activo"`
}
