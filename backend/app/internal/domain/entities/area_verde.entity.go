package entities

import "time"

// AreaVerde represents a cadastral green area polygon.
type AreaVerde struct {
	ID          int64
	FeatureID   string
	SourceIndex int
	Codigo      *string
	Nombre      *string
	Uso         *string
	ProyRiego   *string
	RiegoAct    *string
	Referencia  *string
	PerimetroM  *float64
	AreaM2      *float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AreaVerdeFicha represents green area card metadata for display and editing.
type AreaVerdeFicha struct {
	FeatureID  string
	Nombre     string
	Uso        string
	RiegoAct   string
	Referencia string
	AreaM2     *float64
	ConGeom    bool
}
