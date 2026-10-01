// Package entities defines core domain entities.
package entities

// ActividadProperties represents the properties embedded in an activity GeoJSON Feature.
type ActividadProperties struct {
	ID                string  `json:"id"`
	Tipo              string  `json:"tipo"`
	Estado            string  `json:"estado"`
	Titulo            string  `json:"titulo"`
	Detalle           string  `json:"detalle,omitempty"`
	AreaFeatureID     *string `json:"area_feature_id,omitempty"`
	ZonaFeatureID     *string `json:"zona_feature_id,omitempty"`
	AssignedCapatazID *string `json:"assigned_capataz_id,omitempty"`
	Equipo            *string `json:"equipo,omitempty"`
	Ejecutor          string  `json:"ejecutor,omitempty"`
	Archivada         bool    `json:"archivada"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// FiltroIntervenciones contains domain criteria for filtering activities.
type FiltroIntervenciones struct {
	Rol               string
	CapatazID         string
	Estado            string
	Tipo              string
	ZonaSupervisionID string
	CuadrillaID       string
	Origen            string
	SoloAbiertas      bool
}

// NuevaIntervencion contains domain input parameters to create an activity.
type NuevaIntervencion struct {
	ID                string
	Tipo              string
	Titulo            string
	Detalle           string
	Lon               float64
	Lat               float64
	AreaFeatureID     string
	ZonaFeatureID     string
	AssignedCapatazID string
	ActorRol          string
	Ejecutor          string
	UsuarioID         int64
	LugarID           string
	ZonaSupervisionID string
}

// AsignarIntervencion contains domain input for assigning an activity to a capataz.
type AsignarIntervencion struct {
	ID        string
	CapatazID string
	ActorRol  string
	UsuarioID int64
}

// CambiarEstadoIntervencion contains domain input for updating activity state.
type CambiarEstadoIntervencion struct {
	ID        string
	Estado    string
	ActorRol  string
	CapatazID string
	UsuarioID int64
}

// ArchivarIntervencion contains domain input for archiving an activity.
type ArchivarIntervencion struct {
	ID        string
	ActorRol  string
	Motivo    string
	UsuarioID int64
}

// FichaIntervencion contains domain input for editing activity metadata from desktop form.
type FichaIntervencion struct {
	ID             string
	Clase          string
	FechaSolicitud string
	FechaAtencion  string
	Lugar          string
	Comentario     string
	ActorRol       string
	CapatazID      string
}

// NuevoAvance contains domain input for recording progress on an activity.
type NuevoAvance struct {
	ActividadID   string
	ID            string
	Fecha         string
	Nota          string
	AreaFeatureID string
	EjemplarRef   string
	ActorRol      string
	CapatazID     string
}
