// Package requests defines incoming HTTP request bodies for controllers.
package requests

// CrearIntervencionRequest represents the payload for POST /operacion/actividades.
type CrearIntervencionRequest struct {
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
	LugarID           string   `json:"lugar_id"`
	LugarLibre        string   `json:"lugar_libre"`
	Lugar             string   `json:"lugar"`
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
}

// RegistrarPersonalRequest adds fictional names to an existing activity.
type RegistrarPersonalRequest struct {
	NombreFicticio string   `json:"nombre_ficticio"`
	Nombres        []string `json:"nombres"`
}

// AsignarIntervencionRequest represents the payload for PATCH /operacion/actividades/:id/asignacion.
type AsignarIntervencionRequest struct {
	CapatazID string `json:"capataz_id"`
	ActorRol  string `json:"actor_rol"`
}

// CambiarEstadoRequest represents the payload for PATCH /operacion/actividades/:id/estado.
type CambiarEstadoRequest struct {
	Estado    string `json:"estado"`
	ActorRol  string `json:"actor_rol"`
	CapatazID string `json:"capataz_id"`
}

// ArchivarIntervencionRequest represents the payload for POST /operacion/actividades/:id/archivar.
type ArchivarIntervencionRequest struct {
	ActorRol string `json:"actor_rol"`
	Motivo   string `json:"motivo"`
}

// FichaIntervencionRequest represents the payload for PATCH /operacion/actividades/:id/ficha.
type FichaIntervencionRequest struct {
	Clase          string `json:"clase"`
	FechaSolicitud string `json:"fecha_solicitud"`
	FechaAtencion  string `json:"fecha_atencion"`
	Lugar          string `json:"lugar"`
	Comentario     string `json:"comentario"`
}

// RegistrarHitoRequest is the body of POST /operacion/actividades/:id/hitos.
type RegistrarHitoRequest struct {
	Tipo  string `json:"tipo"`
	Texto string `json:"texto"`
}

// CrearAvanceRequest represents the payload for POST /operacion/actividades/:id/avances.
type CrearAvanceRequest struct {
	ID            string `json:"id"`
	Fecha         string `json:"fecha"`
	Nota          string `json:"nota"`
	AreaFeatureID string `json:"area_feature_id"`
	EjemplarRef   string `json:"ejemplar_ref"`
}
