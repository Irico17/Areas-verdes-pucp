package requests

// CrearZonaSupervisionRequest binds input JSON for creating a supervision zone.
type CrearZonaSupervisionRequest struct {
	Codigo  string   `json:"codigo"`
	Nombre  string   `json:"nombre"`
	AreaM2  *float64 `json:"area_m2"`
	GeoJSON string   `json:"geojson"`
}

// CrearCuadrillaRequest binds input JSON for creating a work team.
type CrearCuadrillaRequest struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre_ficticio"`
	Turno  string `json:"turno"`
}

// CrearLugarRequest binds input JSON for creating a place.
type CrearLugarRequest struct {
	Nombre string  `json:"nombre"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	ZonaID *int64  `json:"zona_supervision_id"`
}

// CrearEspecieRequest binds input JSON for creating a botanical species.
type CrearEspecieRequest struct {
	Cientifico string `json:"nombre_cientifico"`
	Comun      string `json:"nombre_comun"`
}

// CrearEjemplarRequest binds input JSON for creating a flora specimen.
type CrearEjemplarRequest struct {
	NumeroOrigen       *int     `json:"numero_origen,omitempty"`
	Codigo             string   `json:"codigo"`
	EspecieID          *int64   `json:"especie_id,omitempty"`
	NombreComun        string   `json:"nombre_comun"`
	TipoVegetacion     string   `json:"tipo_vegetacion"`
	Cantidad           int      `json:"cantidad"`
	UbicacionLugarID   *int64   `json:"ubicacion_lugar_id,omitempty"`
	Referencia         string   `json:"referencia"`
	Lat                *float64 `json:"lat,omitempty"`
	Lon                *float64 `json:"lon,omitempty"`
	ObservacionFen2026 string   `json:"observacion_fen_2026"`
}

// RecodificarRequest binds input JSON for recodifying a specimen.
type RecodificarRequest struct {
	Codigo string `json:"codigo_nuevo"`
}
