// Package dto contains data transfer objects for application use cases.
package dto

import "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"

// CapatazDTO represents a demonstration team.
type CapatazDTO struct {
	ID     string `json:"id"`
	Equipo string `json:"equipo"`
	Turno  string `json:"turno"`
}

// CapatacesResponseDTO wraps a list of capataces.
type CapatacesResponseDTO struct {
	Capataces []CapatazDTO `json:"capataces"`
}

// FiltroIntervencionesDTO contains parameters for querying activities.
type FiltroIntervencionesDTO struct {
	Rol               string
	CapatazID         string
	Estado            string
	Tipo              string
	ZonaSupervisionID string
	CuadrillaID       string
	Origen            string
	SoloAbiertas      bool
}

// CrearIntervencionDTO contains input parameters to create an activity.
type CrearIntervencionDTO struct {
	ID                string   `json:"id"`
	Tipo              string   `json:"tipo"`
	Titulo            string   `json:"titulo"`
	Detalle           string   `json:"detalle"`
	Lon               float64  `json:"lon"`
	Lat               float64  `json:"lat"`
	AreaFeatureID     string   `json:"area_feature_id"`
	ZonaFeatureID     string   `json:"zona_feature_id"`
	AssignedCapatazID string   `json:"assigned_capataz_id"`
	ActorRol          string   `json:"actor_rol"`
	Ejecutor          string   `json:"ejecutor"`
	UsuarioID         int64    `json:"usuario_id"`
	LugarID           string   `json:"lugar_id"`
	ZonaSupervisionID string   `json:"zona_supervision_id"`
	Origen            string   `json:"origen"`
	CodigoExterno     string   `json:"codigo_externo"`
	UnidadSolicitante string   `json:"unidad_solicitante"`
	NivelRiesgo       string   `json:"nivel_riesgo"`
	FechaProgramada   string   `json:"fecha_programada"`
	Cantidad          *float64 `json:"cantidad"`
	Subtipo           string   `json:"subtipo"`
	Clase             string   `json:"clase"`
	Personal          []string `json:"personal"`
	LugarLibre        string   `json:"lugar_libre"`
	LugarTexto        string   `json:"lugar"`
}

// CrearIntervencionResponseDTO represents the response for creating an activity.
type CrearIntervencionResponseDTO struct {
	Creada  bool             `json:"creada"`
	Feature entities.Feature `json:"feature"`
}

// AsignarIntervencionDTO contains input parameters for assigning an activity to a capataz.
type AsignarIntervencionDTO struct {
	ID        string
	CapatazID string
	ActorRol  string
	UsuarioID int64
}

// CambiarEstadoDTO contains input parameters for updating the state of an activity.
type CambiarEstadoDTO struct {
	ID        string
	Estado    string
	ActorRol  string
	CapatazID string
	UsuarioID int64
}

// ArchivarIntervencionDTO contains input parameters for archiving an activity.
type ArchivarIntervencionDTO struct {
	ID        string
	ActorRol  string
	Motivo    string
	UsuarioID int64
}

// ArchivarIntervencionResponseDTO represents the response for archiving an activity.
type ArchivarIntervencionResponseDTO struct {
	Archivada bool   `json:"archivada"`
	ID        string `json:"id"`
}

// FichaIntervencionDTO contains input parameters for editing activity metadata from desktop.
type FichaIntervencionDTO struct {
	ID             string `json:"id"`
	Clase          string `json:"clase"`
	FechaSolicitud string `json:"fecha_solicitud"`
	FechaAtencion  string `json:"fecha_atencion"`
	Lugar          string `json:"lugar"`
	Comentario     string `json:"comentario"`
	ActorRol       string `json:"actor_rol"`
	CapatazID      string `json:"capataz_id"`
}

// EventoTimelineDTO represents an individual event in an activity timeline.
type EventoTimelineDTO struct {
	ID          int64   `json:"id"`
	Tipo        string  `json:"tipo"`
	Estado      *string `json:"estado,omitempty"`
	CapatazID   *string `json:"capataz_id,omitempty"`
	Equipo      *string `json:"equipo,omitempty"`
	ActorRol    string  `json:"actor_rol,omitempty"`
	UsuarioID   *int64  `json:"usuario_id,omitempty"`
	Usuario     string  `json:"usuario,omitempty"`
	Nombre      string  `json:"usuario_nombre,omitempty"`
	Nota        string  `json:"nota,omitempty"`
	UUIDCliente *string `json:"uuid_cliente,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

// TimelineResponseDTO represents the full timeline response for an activity.
type TimelineResponseDTO struct {
	ActividadID string              `json:"actividad_id"`
	Eventos     []EventoTimelineDTO `json:"eventos"`
}

// OpcionCatalogoDTO is a catalog code with its visible name.
type OpcionCatalogoDTO struct {
	Codigo string `json:"codigo"`
	Nombre string `json:"nombre"`
}

// ClaseActividadDTO is the first level of the activity taxonomy and its types.
type ClaseActividadDTO struct {
	Codigo string              `json:"codigo"`
	Nombre string              `json:"nombre"`
	Tipos  []OpcionCatalogoDTO `json:"tipos"`
}

// PersonalFicticioDTO is a fictional name offered when creating an activity.
type PersonalFicticioDTO struct {
	ID             string `json:"id"`
	NombreFicticio string `json:"nombre_ficticio"`
}

// TaxonomiaActividadDTO is the catalog payload for the create form.
type TaxonomiaActividadDTO struct {
	Clases   []ClaseActividadDTO   `json:"clases"`
	Riesgos  []OpcionCatalogoDTO   `json:"riesgos"`
	Origenes []OpcionCatalogoDTO   `json:"origenes"`
	Personal []PersonalFicticioDTO `json:"personal"`
}

// PersonalLaborDTO is a fictional participant stored on an activity.
type PersonalLaborDTO struct {
	ID             string `json:"id"`
	NombreFicticio string `json:"nombre_ficticio"`
	RolCampo       string `json:"rol_campo"`
}

// PersonalLaborResponseDTO lists the fictional participants of an activity.
type PersonalLaborResponseDTO struct {
	Personal []PersonalLaborDTO `json:"personal"`
}

// RegistrarPersonalDTO adds fictional names to an activity. It does not create accounts.
type RegistrarPersonalDTO struct {
	ActividadID string
	Nombres     []string
	ActorRol    string
}

// CrearAvanceDTO contains input parameters for recording progress on an activity.
type CrearAvanceDTO struct {
	ActividadID   string
	ID            string
	Fecha         string
	Nota          string
	AreaFeatureID string
	EjemplarRef   string
	ActorRol      string
	CapatazID     string
}
