package dto

// ZonaSupervisionDTO represents one of the four supervision zones Z1–Z4.
type ZonaSupervisionDTO struct {
	ID      int64    `json:"id"`
	Codigo  string   `json:"codigo"`
	Nombre  string   `json:"nombre"`
	AreaM2  *float64 `json:"area_m2,omitempty"`
	ConGeom bool     `json:"con_geometria"`
	Activo  bool     `json:"activo"`
}

// CuadrillaDTO represents an operational work team.
type CuadrillaDTO struct {
	ID             string `json:"id"`
	NombreFicticio string `json:"nombre_ficticio"`
	Turno          string `json:"turno"`
	Activo         bool   `json:"activo"`
}

// LugarDTO represents a named campus location with geographic coordinates.
type LugarDTO struct {
	ID                int64   `json:"id"`
	Nombre            string  `json:"nombre"`
	NombreNorm        string  `json:"nombre_norm"`
	Lat               float64 `json:"lat"`
	Lon               float64 `json:"lon"`
	ZonaSupervisionID *int64  `json:"zona_supervision_id,omitempty"`
	Activo            bool    `json:"activo"`
}

// EspecieDTO classifies botanical specimens by scientific and common names.
type EspecieDTO struct {
	ID               int64  `json:"id"`
	NombreCientifico string `json:"nombre_cientifico"`
	NombreComun      string `json:"nombre_comun"`
	Activo           bool   `json:"activo"`
}

// EjemplarDTO represents an individual flora specimen on campus.
type EjemplarDTO struct {
	ID                  int64    `json:"id"`
	NumeroOrigen        *int     `json:"numero_origen,omitempty"`
	Codigo              string   `json:"codigo"`
	EspecieID           *int64   `json:"especie_id,omitempty"`
	NombreComun         string   `json:"nombre_comun"`
	TipoVegetacion      string   `json:"tipo_vegetacion"`
	Cantidad            int      `json:"cantidad"`
	UbicacionLugarID    *int64   `json:"ubicacion_lugar_id,omitempty"`
	Referencia          string   `json:"referencia"`
	Lat                 *float64 `json:"lat,omitempty"`
	Lon                 *float64 `json:"lon,omitempty"`
	ObservacionFen2026  string   `json:"observacion_fen_2026"`
	Salud               *string  `json:"salud"`
	SectorCuartelID     *int64   `json:"sector_cuartel_id,omitempty"`
	SectorCuartelNombre string   `json:"sector_cuartel_nombre,omitempty"`
	SectorCuartelClase  string   `json:"sector_cuartel_clase,omitempty"`
	EspecieCientifico   string   `json:"especie_cientifico,omitempty"`
	LugarNombre         string   `json:"lugar_nombre,omitempty"`
	Activo              bool     `json:"activo"`
}

// ActualizarEjemplarDTO is a partial edit. Tiene* means the key came in the JSON.
type ActualizarEjemplarDTO struct {
	Codigo           *string
	TieneCodigo      bool
	EspecieID        *int64
	TieneEspecie     bool
	Salud            *string
	TieneSalud       bool
	Lat              *float64
	Lon              *float64
	TieneLat         bool
	TieneLon         bool
	UbicacionLugarID *int64
	TieneLugar       bool
	SectorCuartelID  *int64
	TieneSector      bool
}

// CodigoHistoricoDTO preserves previous and new codes of a specimen when recodified.
type CodigoHistoricoDTO struct {
	ID             int64  `json:"id"`
	EjemplarID     int64  `json:"ejemplar_id"`
	CodigoAnterior string `json:"codigo_anterior"`
	CodigoNuevo    string `json:"codigo_nuevo"`
}

// PoligonoCuadrillaDTO represents an operational work team sector polygon.
type PoligonoCuadrillaDTO struct {
	ID                int64  `json:"id"`
	FeatureID         string `json:"feature_id"`
	Codigo            string `json:"codigo"`
	Nombre            string `json:"nombre"`
	CuadrillaID       string `json:"cuadrilla_id"`
	ZonaSupervisionID *int64 `json:"zona_supervision_id,omitempty"`
	ConGeom           bool   `json:"con_geometria"`
	Activo            bool   `json:"activo"`
}

// CapaFichaDTO represents a record from an auxiliary reference layer.
type CapaFichaDTO struct {
	ID         int64  `json:"id"`
	FeatureID  string `json:"feature_id"`
	Nombre     string `json:"nombre,omitempty"`
	Codigo     string `json:"codigo,omitempty"`
	Nota       string `json:"nota,omitempty"`
	Clase      string `json:"clase,omitempty"`
	Riego      string `json:"riego,omitempty"`
	Pertenecen string `json:"pertenecen,omitempty"`
	Activo     bool   `json:"activo"`
}

// CrearZonaSupervisionDTO represents the input data to create a supervision zone.
type CrearZonaSupervisionDTO struct {
	Codigo  string   `json:"codigo"`
	Nombre  string   `json:"nombre"`
	AreaM2  *float64 `json:"area_m2"`
	GeoJSON string   `json:"geojson"`
}

// ActualizarZonaSupervisionDTO edits a supervision zone. Empty GeoJSON keeps the current polygon.
type ActualizarZonaSupervisionDTO struct {
	Nombre  string   `json:"nombre"`
	AreaM2  *float64 `json:"area_m2"`
	GeoJSON string   `json:"geojson"`
}

// CrearCuadrillaDTO represents the input data to create a work team.
type CrearCuadrillaDTO struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre_ficticio"`
	Turno  string `json:"turno"`
}

// CrearLugarDTO represents the input data to create a place.
type CrearLugarDTO struct {
	Nombre string  `json:"nombre"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	ZonaID *int64  `json:"zona_supervision_id"`
}

// CrearEspecieDTO represents the input data to create a botanical species.
type CrearEspecieDTO struct {
	Cientifico string `json:"nombre_cientifico"`
	Comun      string `json:"nombre_comun"`
}

// EjemplaresPaginadosDTO represents paginated response for ejemplares.
type EjemplaresPaginadosDTO struct {
	Ejemplares []EjemplarDTO `json:"ejemplares"`
	Total      int           `json:"total"`
	Limit      int           `json:"limit"`
	Offset     int           `json:"offset"`
}

// RecodificarDTO represents input data to recodify a specimen.
type RecodificarDTO struct {
	Codigo string `json:"codigo_nuevo"`
}
