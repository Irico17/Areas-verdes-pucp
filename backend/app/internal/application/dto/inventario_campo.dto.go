// Package dto defines data transfer objects for API requests and responses.
package dto

// TachoDTO represents a waste bin point with 11 count columns.
type TachoDTO struct {
	ID                  int64    `json:"id"`
	Codigo              string   `json:"codigo"`
	Lat                 *float64 `json:"lat,omitempty"`
	Lon                 *float64 `json:"lon,omitempty"`
	Nota                string   `json:"nota"`
	Lugar               string   `json:"lugar"`
	Espacios            string   `json:"espacios"`
	Accion              string   `json:"accion"`
	TachoActual         string   `json:"tacho_actual"`
	TachoNuevo          string   `json:"tacho_nuevo"`
	Recomendaciones     string   `json:"recomendaciones"`
	NoAprovechables     int      `json:"no_aprovechables"`
	PapelCarton         int      `json:"papel_carton"`
	Plastico            int      `json:"plastico"`
	Vidrio              int      `json:"vidrio"`
	Pilas               int      `json:"pilas"`
	Peligrosos          int      `json:"peligrosos"`
	RAEE                int      `json:"raee"`
	Metales             int      `json:"metales"`
	Aniquem             int      `json:"aniquem"`
	IntermediosPlastico int      `json:"intermedios_plastico"`
	IntermediosMetal    int      `json:"intermedios_metal"`
	Activo              bool     `json:"activo"`
}

// ListarTachosResponseDTO wraps the list of tachos.
type ListarTachosResponseDTO struct {
	Tachos []TachoDTO `json:"tachos"`
}

// BebederoDTO represents a drinking fountain point.
type BebederoDTO struct {
	ID      int64    `json:"id"`
	Codigo  string   `json:"codigo"`
	Subtipo string   `json:"subtipo"`
	Estado  string   `json:"estado"`
	Sede    string   `json:"sede"`
	Lat     *float64 `json:"lat,omitempty"`
	Lon     *float64 `json:"lon,omitempty"`
	Activo  bool     `json:"activo"`
}

// ListarBebederosResponseDTO wraps the list of bebederos.
type ListarBebederosResponseDTO struct {
	Bebederos []BebederoDTO `json:"bebederos"`
}

// PuntoDTO represents a public campus point of interest.
type PuntoDTO struct {
	ID     int64   `json:"id"`
	Titulo string  `json:"titulo"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	URL    string  `json:"url"`
	Activo bool    `json:"activo"`
}

// ListarPuntosResponseDTO wraps the list of points of interest.
type ListarPuntosResponseDTO struct {
	Puntos []PuntoDTO `json:"puntos"`
}

// ReservaDTO represents a garden reservation record.
type ReservaDTO struct {
	ID         int64  `json:"id"`
	JardinID   *int64 `json:"jardin_id,omitempty"`
	Fecha      string `json:"fecha"`
	HoraInicio string `json:"hora_inicio"`
	HoraFin    string `json:"hora_fin"`
	Estado     string `json:"estado"`
	Evento     string `json:"evento"`
	Unidad     string `json:"unidad"`
	Origen     string `json:"origen"`
	Activo     bool   `json:"activo"`
}

// ListarReservasResponseDTO wraps the list of reservations.
type ListarReservasResponseDTO struct {
	Origen   string       `json:"origen"`
	Aviso    string       `json:"aviso"`
	Reservas []ReservaDTO `json:"reservas"`
}

// FichaCapaDTO represents an editable layer record.
type FichaCapaDTO struct {
	ID         int64    `json:"id"`
	FeatureID  string   `json:"feature_id"`
	Nombre     string   `json:"nombre"`
	Codigo     string   `json:"codigo"`
	Nota       string   `json:"nota"`
	Clase      string   `json:"clase"`
	Riego      string   `json:"riego"`
	AreaM2     *float64 `json:"area_m2,omitempty"`
	PerimetroM *float64 `json:"perimetro_m,omitempty"`
	Pertenecen string   `json:"pertenecen"`
	Uso        string   `json:"uso"`
	GeoJSON    string   `json:"geojson,omitempty"`
	Activo     bool     `json:"activo"`
}

// ListarFichasCapaResponseDTO wraps the list of layer records.
type ListarFichasCapaResponseDTO struct {
	Filas []FichaCapaDTO `json:"filas"`
}

// BajaResponseDTO represents the response for a successful soft delete.
type BajaResponseDTO struct {
	Activo bool `json:"activo"`
}

// RechazoDTO represents an individual row rejection in CSV format parsing.
type RechazoDTO struct {
	Fuente string `json:"fuente"`
	Fila   int    `json:"fila"`
	Campo  string `json:"campo"`
	Motivo string `json:"motivo"`
}

// FormatoPuntosResponseDTO represents the result of the formato/puntos endpoint.
type FormatoPuntosResponseDTO struct {
	Filas            int          `json:"filas"`
	Rechazados       []RechazoDTO `json:"rechazados"`
	ColumnasOmitidas []string     `json:"columnas_omitidas"`
	Nota             string       `json:"nota"`
}
