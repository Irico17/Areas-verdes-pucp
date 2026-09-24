package operacion

import "time"

// Roles de supervisión. No hay sesión institucional: el cliente declara el rol.
const (
	RolJefatura     = "jefatura"
	RolCoordinacion = "coordinacion"
	RolCapataz      = "capataz"
	RolAdmin        = "admin"
)

// Tipos de labor que el mapa distingue.
var Tipos = []string{"riego", "poda", "limpieza", "incidencia", "inspeccion"}

// Estados operativos. cerrada y cancelada salen del listado abierto.
var Estados = []string{"pendiente", "en_proceso", "bloqueada", "cerrada", "cancelada"}

// Recinto aproximado del campus Pando (EPSG:4326), con un margen corto.
const (
	MinLon = -77.090
	MaxLon = -77.070
	MinLat = -12.080
	MaxLat = -12.060
)

// Capataz es un equipo ficticio, no una persona.
type Capataz struct {
	ID     string `json:"id"`
	Equipo string `json:"equipo"`
	Turno  string `json:"turno"`
}

// ActividadProperties viajan dentro del GeoJSON de labores.
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

// Evento es una fila de la bitácora.
type Evento struct {
	ID        int64   `json:"id"`
	Tipo      string  `json:"tipo"`
	Estado    *string `json:"estado,omitempty"`
	CapatazID *string `json:"capataz_id,omitempty"`
	Equipo    *string `json:"equipo,omitempty"`
	ActorRol  string  `json:"actor_rol"`
	Nota      string  `json:"nota,omitempty"`
	CreatedAt string  `json:"created_at"`
}

// Timeline agrupa la bitácora de una labor.
type Timeline struct {
	ActividadID string   `json:"actividad_id"`
	Eventos     []Evento `json:"eventos"`
}

// CreateInput es el alta desde un pin. ID lo genera el cliente (UUID idempotente).
type CreateInput struct {
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
}

// Query filtra el listado GeoJSON.
type Query struct {
	Rol          string
	CapatazID    string
	Estado       string
	Tipo         string
	SoloAbiertas bool
}

// Saved es la fila que se compara en un reintento idempotente.
type Saved struct {
	Tipo              string
	Titulo            string
	Detalle           string
	AssignedCapatazID string
	AreaFeatureID     string
	ZonaFeatureID     string
	Lon               float64
	Lat               float64
	Ejecutor          string
	CreatedAt         time.Time
}
