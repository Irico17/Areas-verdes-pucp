// Package entities defines core domain entities.
package entities

// ActividadProperties represents the properties embedded in an activity GeoJSON Feature.
type ActividadProperties struct {
	ID                string   `json:"id"`
	Tipo              string   `json:"tipo"`
	Estado            string   `json:"estado"`
	Titulo            string   `json:"titulo"`
	Detalle           string   `json:"detalle,omitempty"`
	AreaFeatureID     *string  `json:"area_feature_id,omitempty"`
	ZonaFeatureID     *string  `json:"zona_feature_id,omitempty"`
	AssignedCapatazID *string  `json:"assigned_capataz_id,omitempty"`
	Equipo            *string  `json:"equipo,omitempty"`
	Ejecutor          string   `json:"ejecutor,omitempty"`
	EstadoEtiqueta    string   `json:"estado_etiqueta,omitempty"`
	Origen            string   `json:"origen"`
	CodigoExterno     *string  `json:"codigo_externo,omitempty"`
	UnidadSolicitante *string  `json:"unidad_solicitante,omitempty"`
	NivelRiesgo       *string  `json:"nivel_riesgo,omitempty"`
	FechaProgramada   *string  `json:"fecha_programada,omitempty"`
	Cantidad          *float64 `json:"cantidad,omitempty"`
	Subtipo           *string  `json:"subtipo,omitempty"`
	Clase             *string  `json:"clase,omitempty"`
	Personal          []string `json:"personal,omitempty"`
	Archivada         bool     `json:"archivada"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
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
	Origen            string
	CodigoExterno     string
	UnidadSolicitante string
	NivelRiesgo       string
	FechaProgramada   string
	Cantidad          *float64
	Subtipo           string
	Clase             string
	Personal          []string
	LugarLibre        string
	LugarTexto        string
}

// PersonalLabor is a fictional name assigned to an activity. It is not a user account.
type PersonalLabor struct {
	ID             string
	ActividadID    string
	NombreFicticio string
	RolCampo       string
}

// OpcionCatalogo is a code and its visible name.
type OpcionCatalogo struct {
	Codigo string
	Nombre string
	Padre  string
}

// ClaseActividad groups second-level activity types under a class.
type ClaseActividad struct {
	Codigo string
	Nombre string
	Tipos  []OpcionCatalogo
}

// PersonalFicticio is a seeded name that can be chosen for an activity.
type PersonalFicticio struct {
	ID             string
	NombreFicticio string
}

// TaxonomiaActividad is the catalog the create form reads.
type TaxonomiaActividad struct {
	Clases   []ClaseActividad
	Riesgos  []OpcionCatalogo
	Origenes []OpcionCatalogo
	Personal []PersonalFicticio
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
